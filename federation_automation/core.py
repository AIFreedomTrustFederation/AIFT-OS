"""AIFT Federation Automation: bounded GitHub App identity, queue, and jobs."""
import base64
import hashlib
import hmac
import json
import os
import re
import sqlite3
import stat
import subprocess
import time
from urllib.error import HTTPError, URLError
from urllib.parse import quote
from urllib.request import Request, urlopen

APP_ID = 5861093
OWNER = "AIFreedomTrustFederation"
READ = {"contents": "read", "pull_requests": "read", "checks": "read"}
WRITE = {"contents": "write", "pull_requests": "write"}
EVIDENCE = {"contents": "read", "actions": "read"}
REPO = re.compile(r"AIFreedomTrustFederation/[A-Za-z0-9_.-]+\Z")
DOC = re.compile(r"docs/[A-Za-z0-9_./-]+\.md\Z")
SENSITIVE = re.compile(r"governan|doctrin|constitut|trust|identit|secur|policy|credential|deploy|finance|legal|auth|secret", re.I)


class Blocked(Exception):
    pass


class ApiError(Blocked):
    def __init__(self, status):
        self.status = status
        super().__init__(f"GitHub API status {status}")


def repo_ok(name):
    if not isinstance(name, str) or not REPO.fullmatch(name):
        raise Blocked("Repository outside allowed account")
    return name


def doc_ok(path):
    if (not isinstance(path, str) or not DOC.fullmatch(path) or
            ".." in path.split("/") or SENSITIVE.search(path)):
        raise Blocked("Not approved for routine Markdown formatting")
    return path


def signature_ok(secret, body, header):
    if not secret or not re.fullmatch(r"sha256=[0-9a-f]{64}", header or ""):
        return False
    return hmac.compare_digest("sha256=" + hmac.new(secret, body, hashlib.sha256).hexdigest(), header)


def read_private(path):
    if not path:
        raise Blocked("Credential path missing")
    try:
        fd = os.open(path, os.O_RDONLY | getattr(os, "O_NOFOLLOW", 0))
        try:
            info = os.fstat(fd)
            if not stat.S_ISREG(info.st_mode) or info.st_mode & 0o077:
                raise Blocked("Credential must be an owner-private regular file")
            with os.fdopen(fd, "rb", closefd=False) as source:
                data = source.read(65537)
        finally:
            os.close(fd)
    except OSError:
        raise Blocked("Cannot access credential securely") from None
    if not 1 <= len(data) <= 65536:
        raise Blocked("Invalid credential length")
    return data


class GitHub:
    def __init__(self, token):
        self.token = token

    def request(self, method, path, body=None):
        if not path.startswith("/") or path.startswith("//"):
            raise Blocked("Invalid API path")
        data = None if body is None else json.dumps(body).encode()
        headers = {"Accept": "application/vnd.github+json",
                   "User-Agent": "aift-federation-automation/0.1",
                   "X-GitHub-Api-Version": "2022-11-28",
                   "Authorization": "Bearer " + self.token}
        if data is not None:
            headers["Content-Type"] = "application/json"
        req = Request("https://api.github.com" + path, data=data, headers=headers, method=method)
        try:
            with urlopen(req, timeout=20) as response:
                return json.loads(response.read(1024 * 1024))
        except HTTPError as error:
            raise ApiError(error.code) from None
        except (URLError, TimeoutError, ValueError):
            raise Blocked("GitHub network or response failure") from None


class App:
    def __init__(self, key_path, factory=GitHub):
        self.key_path, self.factory = key_path, factory

    def jwt(self):
        pem = read_private(self.key_path)
        if not pem.startswith((b"-----BEGIN RSA PRIVATE KEY-----", b"-----BEGIN PRIVATE KEY-----")):
            raise Blocked("Expected GitHub App RSA PEM")
        del pem
        now = int(time.time())
        b64 = lambda s: base64.urlsafe_b64encode(s).rstrip(b"=").decode()
        head = b64(b'{"alg":"RS256","typ":"JWT"}')
        claims = b64(json.dumps({"iat": now - 30, "exp": now + 480,
                                  "iss": str(APP_ID)}, separators=(",", ":")).encode())
        msg = (head + "." + claims).encode()
        try:
            result = subprocess.run(["openssl", "dgst", "-sha256", "-sign", self.key_path],
                                    input=msg, stdout=subprocess.PIPE,
                                    stderr=subprocess.DEVNULL, check=True, timeout=10)
        except (OSError, subprocess.SubprocessError):
            raise Blocked("App PEM signing failed") from None
        return msg.decode() + "." + b64(result.stdout)

    def identify(self):
        api = self.factory(self.jwt())
        meta = api.request("GET", "/app")
        if meta.get("id") != APP_ID:
            raise Blocked("Authenticated GitHub App ID mismatch")
        items = api.request("GET", "/app/installations?per_page=100")
        matches = [x for x in items if x.get("account", {}).get("login") == OWNER]
        if len(matches) != 1 or not isinstance(matches[0].get("id"), int):
            raise Blocked("Unique federation installation not verified")
        return meta, matches[0]

    def token(self, repo, permissions):
        repo_ok(repo)
        if permissions not in (READ, WRITE, EVIDENCE):
            raise Blocked("Unsupported permission grant")
        _, installation = self.identify()
        api = self.factory(self.jwt())
        result = api.request("POST", f"/app/installations/{installation['id']}/access_tokens",
                             {"repositories": [repo.split("/")[1]], "permissions": permissions})
        if not isinstance(result.get("token"), str):
            raise Blocked("Installation token not issued")
        if any(result.get("permissions", {}).get(k) != v for k, v in permissions.items()):
            raise Blocked("Installation token permissions not verified")
        extras = set(result.get("permissions", {})) - set(permissions) - {"metadata"}
        if extras:
            raise Blocked("Unexpected installation token permission grants")
        repositories = result.get("repositories", [])
        if len(repositories) != 1 or repositories[0].get("full_name") != repo:
            raise Blocked("Installation token repository scope not verified")
        return self.factory(result["token"])

    def discovery_api(self):
        """Read installation repository metadata without granting contents or PR writes."""
        _, installation = self.identify()
        api = self.factory(self.jwt())
        result = api.request("POST", f"/app/installations/{installation['id']}/access_tokens",
                             {"permissions": {"metadata": "read"}})
        if not isinstance(result.get("token"), str):
            raise Blocked("Discovery token not issued")
        actual = result.get("permissions", {})
        if actual.get("metadata") != "read" or set(actual) != {"metadata"}:
            raise Blocked("Discovery token is not metadata-only")
        return self.factory(result["token"])


class Queue:
    def __init__(self, path, clock=time.time):
        self.clock = clock
        self.db = sqlite3.connect(path, timeout=30)
        self.db.row_factory = sqlite3.Row
        self.db.execute("PRAGMA journal_mode=WAL")
        self.db.executescript("""
          CREATE TABLE IF NOT EXISTS jobs(
            id TEXT PRIMARY KEY,kind TEXT NOT NULL,payload TEXT NOT NULL,
            state TEXT NOT NULL,attempts INTEGER NOT NULL DEFAULT 0,
            due REAL NOT NULL,lease REAL NOT NULL DEFAULT 0,result TEXT,updated REAL NOT NULL);
          CREATE TABLE IF NOT EXISTS settings(key TEXT PRIMARY KEY,value TEXT NOT NULL);
          CREATE TABLE IF NOT EXISTS deliveries(id TEXT PRIMARY KEY,received REAL NOT NULL);
          CREATE TABLE IF NOT EXISTS audit(id INTEGER PRIMARY KEY AUTOINCREMENT,
            ts REAL NOT NULL,event TEXT NOT NULL,subject TEXT NOT NULL,detail TEXT NOT NULL);
        """)

    def log(self, event, subject, detail=""):
        with self.db:
            self.db.execute("INSERT INTO audit(ts,event,subject,detail) VALUES(?,?,?,?)",
                            (self.clock(), event, subject, str(detail)[:300]))

    def get(self, key, default=""):
        item = self.db.execute("SELECT value FROM settings WHERE key=?", (key,)).fetchone()
        return item[0] if item else default

    def set(self, key, value):
        with self.db:
            self.db.execute("INSERT OR REPLACE INTO settings VALUES(?,?)", (key, str(value)))
        self.log("setting", key, value)

    def paused(self):
        return self.get("paused", "1") != "0"

    def enqueue(self, ident, kind, payload):
        if kind not in ("audit", "doc_format") or not isinstance(ident, str) or len(ident) > 180:
            raise Blocked("Unknown job")
        repo_ok(payload["repo"])
        with self.db:
            active = self.db.execute(
                "SELECT COUNT(*) FROM jobs WHERE state IN ('queued','retry','running')").fetchone()[0]
            if active >= 300:
                raise Blocked("Queue limit reached")
            inserted = self.db.execute(
                "INSERT OR IGNORE INTO jobs(id,kind,payload,state,due,updated)"
                " VALUES(?,?,?,'queued',?,?)",
                (ident, kind, json.dumps(payload, sort_keys=True), self.clock(), self.clock()))
        if inserted.rowcount:
            self.log("enqueued", ident, kind)
        return bool(inserted.rowcount)

    def delivery(self, ident):
        if not re.fullmatch(r"[0-9a-fA-F-]{16,64}", ident or ""):
            raise Blocked("Invalid GitHub delivery ID")
        with self.db:
            inserted = self.db.execute("INSERT OR IGNORE INTO deliveries VALUES(?,?)",
                                       (ident, self.clock()))
        return bool(inserted.rowcount)

    def claim(self):
        if self.paused():
            return None
        now = self.clock()
        with self.db:
            self.db.execute("UPDATE jobs SET state='retry',due=?,lease=0"
                            " WHERE state='running' AND lease<=?", (now, now))
            row = self.db.execute(
                "SELECT * FROM jobs WHERE state IN ('queued','retry') AND due<=?"
                " ORDER BY due,id LIMIT 1", (now,)).fetchone()
            if row:
                self.db.execute("UPDATE jobs SET state='running',attempts=attempts+1,"
                                " lease=?,updated=? WHERE id=?", (now + 240, now, row["id"]))
        return dict(row) if row else None

    def finish(self, job, result=None, error=None):
        attempt = job["attempts"] + 1
        state = "done" if error is None else ("dead" if attempt >= 4 else "retry")
        delay = 0 if error is None else min(3600, 10 * 2 ** (attempt - 1))
        safe = result if error is None else {"error_type": type(error).__name__,
                                              "http_status": getattr(error, "status", None)}
        with self.db:
            self.db.execute("UPDATE jobs SET state=?,result=?,due=?,lease=0,updated=? WHERE id=?",
                            (state, json.dumps(safe), self.clock() + delay,
                             self.clock(), job["id"]))
        self.log(state, job["id"], type(error).__name__ if error else "complete")
        return state

    def status(self):
        rows = self.db.execute("SELECT state,COUNT(*) FROM jobs GROUP BY state").fetchall()
        return {"paused": self.paused(), "jobs": dict(rows)}


def receive(queue, event, payload, delivery_id):
    if not queue.delivery(delivery_id):
        return "duplicate_delivery"
    sender = payload.get("sender") or {}
    if sender.get("type") == "Bot" or str(sender.get("login", "")).endswith("[bot]"):
        return "bot_ignored"
    name = repo_ok((payload.get("repository") or {}).get("full_name"))
    if event == "push":
        if payload.get("ref") != "refs/heads/" + payload["repository"].get("default_branch", ""):
            return "non_default_push"
        sha = payload.get("after", "")
    elif event == "pull_request":
        if payload.get("action") not in ("opened", "reopened", "synchronize", "ready_for_review"):
            return "ignored_action"
        sha = (payload.get("pull_request") or {}).get("head", {}).get("sha", "")
    else:
        return "ignored_event"
    if not re.fullmatch(r"[0-9a-f]{40}", sha):
        raise Blocked("Invalid exact-head commit")
    return ("queued" if queue.enqueue("audit:" + name + ":" + sha, "audit",
                                      {"repo": name, "head": sha}) else "duplicate_head")


def work(job, auth, write_repos=frozenset()):
    data = json.loads(job["payload"])
    name = repo_ok(data["repo"])
    if job["kind"] == "audit":
        api = auth.token(name, READ)
        meta = api.request("GET", "/repos/" + name)
        if meta.get("full_name") != name:
            raise Blocked("Repository identity mismatch")
        prs = api.request("GET", f"/repos/{name}/pulls?state=open&per_page=15")
        reports = []
        for pr in prs[:15]:
            sha = pr.get("head", {}).get("sha", "")
            if not re.fullmatch(r"[0-9a-f]{40}", sha):
                continue
            checks = api.request("GET", f"/repos/{name}/commits/{sha}/check-runs?per_page=100")
            runs = checks.get("check_runs", [])
            reports.append({"number": pr.get("number"), "head_sha": sha, "draft": pr.get("draft"),
                            "base_sha": pr.get("base", {}).get("sha"),
                            "checks": [{"name": c.get("name"), "status": c.get("status"),
                                        "conclusion": c.get("conclusion")}
                                       for c in runs if c.get("head_sha") == sha],
                            "checks_truncated": len(runs) >= 100, "merge_ready": None})
        return {"repo": name, "open_pr_sample": reports, "merge_ready": "unverified"}
    if job["kind"] != "doc_format":
        raise Blocked("Unsupported operation")
    if os.getenv("AIFT_ENABLE_DOC_PROPOSALS") != "1" or name not in write_repos:
        raise Blocked("Draft-PR proposals not enabled")
    path = doc_ok(data.get("path"))
    read = auth.token(name, READ)
    meta = read.request("GET", "/repos/" + name)
    branch = meta["default_branch"]
    base = read.request("GET", f"/repos/{name}/branches/{quote(branch, safe='')}")["commit"]["sha"]
    item = read.request("GET", f"/repos/{name}/contents/{quote(path)}?ref={base}")
    if item.get("encoding") != "base64" or item.get("size", 999999) > 200000:
        raise Blocked("File is not small base64-encoded UTF-8")
    try:
        original = base64.b64decode(item["content"]).decode("utf-8")
    except (KeyError, UnicodeError, ValueError):
        raise Blocked("Unsupported document encoding") from None
    fixed = "\n".join(line.rstrip(" \t") for line in original.split("\n"))
    if original == fixed:
        return {"outcome": "no_change", "path": path}
    head = "aift-automation/docfmt-" + hashlib.sha256(
        (name + ":" + path + ":" + base).encode()).hexdigest()[:16]
    write = auth.token(name, WRITE)
    try:
        write.request("POST", f"/repos/{name}/git/refs",
                      {"ref": "refs/heads/" + head, "sha": base})
    except ApiError as error:
        if error.status == 422:
            return {"outcome": "already_proposed", "branch": head}
        raise
    write.request("PUT", f"/repos/{name}/contents/{quote(path)}", {
        "message": "docs: normalize Markdown trailing whitespace",
        "branch": head, "sha": item["sha"],
        "content": base64.b64encode(fixed.encode()).decode()})
    pr = write.request("POST", f"/repos/{name}/pulls", {
        "head": head, "base": branch, "draft": True,
        "title": "docs: normalize safe Markdown whitespace",
        "body": "Bounded whitespace-only automation; exact-head CI and human review required. No merge/deploy."})
    if not pr.get("draft") or not pr.get("html_url"):
        raise Blocked("Draft PR could not be verified")
    return {"outcome": "draft_pr", "url": pr["html_url"], "number": pr.get("number"),
            "base_sha": base}
