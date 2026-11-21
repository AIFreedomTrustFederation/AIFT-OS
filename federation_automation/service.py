"""Independent AIFT App worker; HTTPS proxy and deployment require operator approval."""
import argparse
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
import os
from pathlib import Path
import threading
import time

from .core import (App, Blocked, OWNER, Queue, READ, doc_ok, read_private,
                   receive, repo_ok, signature_ok, work)


def make_handler(db_path, secret):
    class Webhooks(BaseHTTPRequestHandler):
        def log_message(self, *_args):
            pass  # GitHub payloads and credentials never enter access logs.

        def reply(self, code, payload):
            data = json.dumps(payload).encode()
            self.send_response(code)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(data)))
            self.end_headers()
            self.wfile.write(data)

        def do_GET(self):
            if self.path == "/":
                data = (Path(__file__).parent / "console.html").read_bytes()
                self.send_response(200)
                self.send_header("Content-Type", "text/html; charset=utf-8")
                self.send_header("Content-Length", str(len(data)))
                self.send_header("Cache-Control", "no-store")
                self.send_header("X-Content-Type-Options", "nosniff")
                self.send_header("Content-Security-Policy",
                                 "default-src 'self'; style-src 'unsafe-inline'; script-src 'unsafe-inline'; "
                                 "object-src 'none'; frame-ancestors 'none'")
                self.end_headers()
                return self.wfile.write(data)
            if self.path in ("/healthz", "/api/jobs"):
                queue = Queue(db_path)
                if self.path == "/api/jobs":
                    return self.reply(200, {"status": queue.status(),
                                            "recent": queue.recent()})
                return self.reply(200, {"service": "aift-automation",
                                        "status": queue.status()})
            self.reply(404, {"error": "not_found"})

        def do_POST(self):
            if self.path != "/webhook":
                return self.reply(404, {"error": "not_found"})
            try:
                size = int(self.headers.get("Content-Length", "0"))
            except ValueError:
                return self.reply(400, {"error": "invalid_length"})
            if not 0 < size <= 262144:
                return self.reply(413, {"error": "invalid_size"})
            body = self.rfile.read(size)
            if len(body) != size or not signature_ok(
                    secret, body, self.headers.get("X-Hub-Signature-256", "")):
                return self.reply(401, {"error": "invalid_signature"})
            try:
                event = self.headers.get("X-GitHub-Event", "")
                delivery = self.headers.get("X-GitHub-Delivery", "")
                state = receive(Queue(db_path), event, json.loads(body), delivery)
                self.reply(202, {"status": state})
            except (Blocked, ValueError, KeyError, TypeError):
                self.reply(400, {"error": "invalid_event"})
    return Webhooks


def schedule(queue, auth, seconds):
    if queue.paused() or time.time() - float(queue.get("last_scheduled", "0")) < seconds:
        return
    # App-only repository enumeration, never a personal GitHub token.
    api = auth.discovery_api()
    epoch = int(time.time() // seconds)
    for page in range(1, 11):
        results = api.request("GET", f"/installation/repositories?per_page=100&page={page}")
        repos = results.get("repositories", [])
        for repo in repos:
            name = repo.get("full_name", "")
            if name.startswith(OWNER + "/"):
                queue.enqueue("periodic:" + str(epoch) + ":" + name,
                              "audit", {"repo": repo_ok(name)})
        if len(repos) < 100:
            break
    queue.set("last_scheduled", time.time())


def daemon(args):
    secret = read_private(os.getenv("AIFT_WEBHOOK_SECRET_FILE", "")).rstrip(b"\r\n")
    if not 32 <= len(secret) <= 256:
        raise Blocked("Webhook secret has invalid length")
    auth = App(os.getenv("AIFT_APP_PRIVATE_KEY_FILE", ""))
    auth.identify()  # Fail closed until actual App ID + installation validated.
    write_repos = {repo_ok(s.strip()) for s in
                   os.getenv("AIFT_WRITE_REPOS", "").split(",") if s.strip()}
    queue = Queue(args.db)
    server = ThreadingHTTPServer(("127.0.0.1", args.port), make_handler(args.db, secret))
    thread = threading.Thread(target=server.serve_forever, daemon=True)
    thread.start()
    print(json.dumps({"event": "worker_started", "paused": queue.paused(),
                      "listen": "127.0.0.1", "port": args.port}), flush=True)
    try:
        while True:
            try:
                schedule(queue, auth, max(900, args.audit_interval))
                job = queue.claim()
                if job:
                    try:
                        queue.finish(job, result=work(job, auth, write_repos))
                    except Exception as error:
                        queue.finish(job, error=error)
                else:
                    time.sleep(3)
            except Exception:
                queue.log("loop_error", "redacted")
                time.sleep(30)
    finally:
        server.shutdown()
        server.server_close()


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--db", default="aift-worker.sqlite3")
    parser.add_argument("--port", type=int, default=8787)
    parser.add_argument("--audit-interval", type=int, default=3600)
    sub = parser.add_subparsers(dest="action", required=True)
    for cmd in ("run", "status", "pause", "resume", "check-auth"):
        sub.add_parser(cmd)
    audit = sub.add_parser("enqueue-audit")
    audit.add_argument("--repo", required=True)
    docs = sub.add_parser("enqueue-doc-format")
    docs.add_argument("--repo", required=True)
    docs.add_argument("--path", required=True)
    args = parser.parse_args(argv)
    if args.action == "run":
        return daemon(args)
    if args.action == "check-auth":
        app, install = App(os.getenv("AIFT_APP_PRIVATE_KEY_FILE", "")).identify()
        print(json.dumps({"app_id": app["id"], "client_id": app.get("client_id"),
                          "installation_id": install["id"], "installation_owner": OWNER,
                          "permission_names": sorted(install.get("permissions", {})),
                          "repository_selection": install.get("repository_selection")}))
        return
    queue = Queue(args.db)
    if args.action == "pause":
        queue.set("paused", "1")
    elif args.action == "resume":
        queue.set("paused", "0")
    elif args.action == "enqueue-audit":
        name = repo_ok(args.repo)
        queue.enqueue("manual:audit:" + name + ":" + str(time.time_ns()),
                      "audit", {"repo": name})
    elif args.action == "enqueue-doc-format":
        name, path = repo_ok(args.repo), doc_ok(args.path)
        queue.enqueue("manual:docfmt:" + name + ":" + path + ":" + str(time.time_ns()),
                      "doc_format", {"repo": name, "path": path})
    print(json.dumps(queue.status(), sort_keys=True))


if __name__ == "__main__":
    try:
        main()
    except Blocked as error:
        print(json.dumps({"status": "blocked", "reason": str(error)}))
        raise SystemExit(2)
