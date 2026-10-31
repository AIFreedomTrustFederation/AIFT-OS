#!/usr/bin/env python3
"""Black-box tests of the authenticated GitHub Actions evidence collector."""

import json
import hashlib
import importlib.util
import os
import subprocess
import sys
import tempfile
import textwrap
import unittest
from unittest import mock
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts/collect_federation_evidence.py"
REPOS = {
    "AIFreedomTrustFederation/.github": ("40146248552", "b14a458e6e6ef4bbda5cde96ee445e66678fee9c", "aift-catalog-verification", "aift-catalog-verification.json", 12580921836),
    "AIFreedomTrustFederation/AIFT-Runtime": ("40150125926", "d953ddccacdb5700d46e2d79d5af7c7dd309807f", "aift-runtime-verification", "aift-runtime-verification.json", 12582097262),
}

FAKE_GH = r'''#!/usr/bin/env python3
import hashlib,io,json,os,sys,zipfile
from pathlib import Path
args=sys.argv[1:]
repo_by_id={"40146248552":"AIFreedomTrustFederation/.github","40150125926":"AIFreedomTrustFederation/AIFT-Runtime"}
repos={
"AIFreedomTrustFederation/.github":("40146248552","b14a458e6e6ef4bbda5cde96ee445e66678fee9c","aift-catalog-verification", "aift-catalog-verification.json",12580921836),
"AIFreedomTrustFederation/AIFT-Runtime":("40150125926","d953ddccacdb5700d46e2d79d5af7c7dd309807f","aift-runtime-verification","aift-runtime-verification.json",12582097262),
}
mode=os.environ.get("FAKE_GH_MODE", "ok")
def archive_bytes(repo):
 filename=repos[repo][3]
 if mode=="corrupt_archive":return b"not a ZIP"
 fixture_name=next(v[3] for k,v in repos.items() if k!=repo) if mode=="swapped_producer_records" else filename
 data=(Path(os.environ["FAKE_GH_FIXTURES"])/fixture_name).read_bytes()
 if mode=="tampered" and repo.endswith("/.github"):
  record=json.loads(data);record["source"]["run_url"]="https://github.com/AIFreedomTrustFederation/.github/actions/runs/123"
  data=json.dumps(record).encode()
 if mode in ("nonfinite_producer_json", "deeply_nested_producer_json") and repo.endswith("/.github"):
  # Keep authenticated ZIP size/digest valid; only the embedded JSON is malformed.
  extra=(b"1e9999" if mode=="nonfinite_producer_json" else
         b"["*20000+b"0"+b"]"*20000)
  data=data.rstrip()[:-1]+b',"ignored":'+extra+b"}"
 buffer=io.BytesIO()
 with zipfile.ZipFile(buffer,"w",compression=zipfile.ZIP_DEFLATED) as archive:
  info=zipfile.ZipInfo(filename,date_time=(2020,1,1,0,0,0));info.compress_type=zipfile.ZIP_DEFLATED
  archive.writestr(info,data)
  if mode=="extra_member":
   extra=zipfile.ZipInfo("unexpected.txt",date_time=(2020,1,1,0,0,0));extra.compress_type=zipfile.ZIP_DEFLATED
   archive.writestr(extra,b"untrusted")
 if mode=="invalid_deflate":
  # Preserve a valid ZIP container and metadata while corrupting its deflate stream.
  damaged=bytearray(buffer.getvalue())
  offset=30+int.from_bytes(damaged[26:28],"little")+int.from_bytes(damaged[28:30],"little")
  damaged[offset]=0xFF
  return bytes(damaged)
 return buffer.getvalue()
if args[0]=="api":
 path=args[1].removeprefix("repos/")
 repo=next((r for r in repos if path==r or path.startswith(r+"/")),None)
 if not repo: sys.exit(2)
 run,sha,artifact,filename,artifact_id=repos[repo]
 suffix=path[len(repo):]
 if suffix=="" and mode=="oversized_metadata":
  sys.stdout.buffer.write(b"x"*(2*1024*1024));sys.exit(0)
 if suffix=="" and mode=="invalid_utf8_metadata":
  sys.stdout.buffer.write(bytes([255]));sys.exit(0)
 if suffix=="" and mode=="deeply_nested_metadata":
  sys.stdout.write('{"id":' + "[" * 20000 + "0" + "]" * 20000 + "}");sys.exit(0)
 if suffix=="" and mode in ("overflowing_exponent_metadata","negative_overflowing_exponent_metadata","finite_exponent_metadata"):
  repo_id=1271861965 if repo.endswith("/.github") else 1298897337
  exponent={"overflowing_exponent_metadata":"1e9999",
            "negative_overflowing_exponent_metadata":"-1e9999",
            "finite_exponent_metadata":"1e2"}[mode]
  sys.stdout.write('{"default_branch":"main","id":' + str(repo_id) + ',"extra":' + exponent + "}");sys.exit(0)
 if suffix=="":
  result={"default_branch":"main","id":1271861965 if repo.endswith("/.github") else 1298897337}
  if mode=="repo_id_changed" and (Path(os.environ["FAKE_GH_FIXTURES"])/".branch-advanced").exists():
   result["id"]+=1
 elif suffix=="/branches/main":
  advanced=mode=="branch_advanced" and (Path(os.environ["FAKE_GH_FIXTURES"])/".branch-advanced").exists()
  result={"commit":{"sha":"0"*40 if mode=="stale" or advanced else sha}}
 elif suffix=="/actions/runs/"+run:
  repository_id=1271861965 if repo.endswith("/.github") else 1298897337
  result={"id":int(run),"repository":{"full_name":repo,"id":repository_id},
          "head_repository":{"full_name":repo,"id":repository_id},
          "head_sha":sha,"head_branch":"main","event":"push","status":"completed","conclusion":"failure" if mode in ("failed_run", "failed_run_with_failed_gate", "failed_run_with_failed_check") else "success",
          "workflow_id":362136764 if repo.endswith("/.github") else 361044013,
          "path":".github/workflows/catalog-integrity.yml" if repo.endswith("/.github") else ".github/workflows/qa.yml",
          "run_attempt":2 if mode in ("artifact_created_before_run", "artifact_created_just_before_run") else 1,
          "run_started_at":"2026-10-25T00:19:35Z" if repo.endswith("/.github") else "2026-10-25T01:18:37Z",
          "updated_at":"2026-10-25T00:19:49Z" if repo.endswith("/.github") else "2026-10-25T01:18:49Z"}
  if mode=="wrong_run_id":result["id"]+=1
  if mode=="missing_run_id":result.pop("id")
  if mode=="wrong_run_repo_id":result["repository"]["id"]+=1
  if mode=="wrong_run_head_repo_id":result["head_repository"]["id"]+=1
  if mode=="missing_run_head_repo":result.pop("head_repository")
  if mode=="missing_run_timestamps":result.pop("run_started_at")
  if mode=="invalid_run_attempt":result["run_attempt"]=True
  if mode=="wrong_workflow":result["workflow_id"]=999
  if mode=="wrong_workflow_path":result["path"]=".github/workflows/untrusted.yml"
 elif suffix=="/actions/runs/"+run+"/artifacts?per_page=100":
  digest="sha256:"+hashlib.sha256(archive_bytes(repo)).hexdigest()
  if mode=="digest_mismatch":digest="sha256:"+"0"*64
  if mode=="invalid_digest":digest="sha256:not-a-digest"
  repository_id=1271861965 if repo.endswith("/.github") else 1298897337
  artifact_run={"id":int(run),"head_sha":sha,"head_branch":"main",
                "repository_id":repository_id,"head_repository_id":repository_id}
  if mode=="artifact_wrong_run_id":artifact_run["id"]+=1
  if mode=="artifact_wrong_run_sha":artifact_run["head_sha"]="0"*40
  if mode=="artifact_wrong_run_branch":artifact_run["head_branch"]="other"
  if mode=="artifact_wrong_repo_id":artifact_run["repository_id"]+=1
  if mode=="artifact_wrong_head_repo_id":artifact_run["head_repository_id"]+=1
  item={"name":artifact,"id":artifact_id,"expired":False,"digest":digest,
        "created_at":"2026-10-25T00:19:46Z" if repo.endswith("/.github") else "2026-10-25T01:18:46Z",
        "size_in_bytes":len(archive_bytes(repo)),"workflow_run":artifact_run}
  if mode=="artifact_created_before_run":item["created_at"]="2026-10-24T00:00:00Z"
  if mode=="artifact_created_after_run":item["created_at"]="2026-10-25T02:00:00Z"
  if mode=="artifact_created_just_before_run":
   item["created_at"]="2026-10-25T00:19:34Z" if repo.endswith("/.github") else "2026-10-25T01:18:36Z"
  if mode=="artifact_created_just_after_run":
   item["created_at"]="2026-10-25T00:19:50Z" if repo.endswith("/.github") else "2026-10-25T01:18:50Z"
  if mode=="artifact_missing_created_at":item.pop("created_at")
  if mode=="artifact_missing_run_identity":item.pop("workflow_run")
  if mode=="oversized_artifact":item["size_in_bytes"]=5*1024*1024
  if mode=="wrong_artifact_size":item["size_in_bytes"]-=1
  if mode=="missing_artifact_size":item.pop("size_in_bytes")
  if mode=="invalid_artifact_size":item["size_in_bytes"]=True
  if mode=="missing_digest":item.pop("digest")
  items=[item]
  if mode=="duplicate_artifact":items*=2
  if mode=="wrong_artifact_id":items[0]["id"]+=1
  result={"total_count":len(items),"artifacts":items}
  if mode=="artifact_count_boolean":result["total_count"]=True
  if mode=="artifact_count_float":result["total_count"]=1.0
 elif suffix=="/actions/artifacts/"+str(artifact_id)+"/zip":
  if mode=="oversized_stream":
   sys.stdout.buffer.write(b"x"*(5*1024*1024))
   sys.exit(0)
  sys.stdout.buffer.write(archive_bytes(repo))
  if mode in ("branch_advanced", "repo_id_changed") and repo.endswith("/AIFT-Runtime"): (Path(os.environ["FAKE_GH_FIXTURES"])/".branch-advanced").write_text("advanced")
  sys.exit(0)
 else: sys.exit(3)
 print(json.dumps(result))
else:sys.exit(5)
'''


class CollectionTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.bin = self.root / "bin"
        self.bin.mkdir()
        executable = self.bin / "gh"
        executable.write_text(FAKE_GH, encoding="utf-8")
        executable.chmod(0o755)

    def invoke(self, mode="ok", output=None, selected_repos=None):
        fixture_dir = self.root / "fixtures"
        fixture_dir.mkdir(exist_ok=True)
        for repo, (run, sha, _, filename, _) in REPOS.items():
            record = {"schema": "aift.verification.v1", "repository": repo, "commit_sha": sha,
                      "generated_at": "2026-10-25T00:19:40+00:00" if repo.endswith("/.github") else "2026-10-25T01:18:42+00:00",
                      "source": {"type": "github-actions", "run_url": f"https://github.com/{repo}/actions/runs/{run}"},
                      "checks": [{"id": "qa", "command": "python -m unittest", "status": "passed", "exit_code": 0}],
                      "overall_status": "passed", "security": {"status": "passed" if mode == "all_gates_pass" else "not_checked"},
                      "dependencies": {"status": "passed" if mode == "all_gates_pass" else "not_checked"},
                      "deployment": {"status": "passed" if mode == "all_gates_pass" else "not_applicable"},
                      "limitations": ["No live deployment tested."]}
            if mode == "future_generated_at":
                record["generated_at"] = "2999-01-01T00:00:00+00:00"
            if mode == "stale_generated_at":
                record["generated_at"] = "2026-10-24T00:00:00+00:00"
            if mode == "post_run_generated_at":
                record["generated_at"] = "2026-10-25T02:00:00+00:00"
            if mode in ("failed_gate", "failed_run_with_failed_gate"):
                record["security"]["status"] = "failed"
            if mode in ("failed_check", "failed_run_with_failed_check"):
                record["checks"][0].update(status="failed", exit_code=1)
                record["overall_status"] = "failed"
            (fixture_dir / filename).write_text(json.dumps(record), encoding="utf-8")
        output = output or self.root / "evidence"
        cmd = [sys.executable, "-B", str(SCRIPT), "--output-dir", str(output)]
        for repo, (run, sha, _, _, _) in REPOS.items():
            if selected_repos is not None and repo not in selected_repos:
                continue
            cmd.extend(["--run", f"{repo}={run}", "--expect", f"{repo}={sha}"])
        env = dict(os.environ, PATH=f"{self.bin}:{os.environ.get('PATH','')}",
                   FAKE_GH_MODE=mode, FAKE_GH_FIXTURES=str(fixture_dir))
        return subprocess.run(cmd, capture_output=True, text=True, env=env, check=False), output

    def test_collects_two_artifact_formats_and_reports_partial(self):
        result, output = self.invoke()
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads((output / "federation-assessment.json").read_text())
        self.assertEqual(report["status"], "partial")
        self.assertEqual(report["provenance"], "github_api_run_and_artifact_checked_unsigned")
        self.assertEqual(len(report["github_runs"]), 2)
        expected_repo_ids = {
            "AIFreedomTrustFederation/.github": 1271861965,
            "AIFreedomTrustFederation/AIFT-Runtime": 1298897337,
        }
        for run in report["github_runs"]:
            self.assertEqual(run["repository_id"], expected_repo_ids[run["repository"]])
        expected_workflows = {
            "AIFreedomTrustFederation/.github": (362136764, ".github/workflows/catalog-integrity.yml"),
            "AIFreedomTrustFederation/AIFT-Runtime": (361044013, ".github/workflows/qa.yml"),
        }
        for run in report["github_runs"]:
            self.assertEqual((run["workflow_id"], run["workflow_path"]),
                             expected_workflows[run["repository"]])
        self.assertEqual(len(list(output.iterdir())), 3)

    def test_rejects_cross_repository_artifact_record_swap(self):
        # Both GitHub artifact digests and the record set remain valid, but
        # each archive contains the other producer's verification payload.
        result, output = self.invoke("swapped_producer_records")
        self.assertEqual(result.returncode, 1, result.stderr)
        self.assertIn("artifact record does not match authenticated producer", result.stderr)
        self.assertFalse(output.exists())

    def test_assessment_exposes_authenticated_artifact_timing_and_size(self):
        # Preserve the GitHub API facts used by the temporal/size guards so a
        # downstream auditor can examine them without re-fetching the run.
        result, output = self.invoke()
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads((output / "federation-assessment.json").read_text())
        expected = {
            "AIFreedomTrustFederation/.github": (
                "2026-10-25T00:19:35Z", "2026-10-25T00:19:49Z", "2026-10-25T00:19:46Z"),
            "AIFreedomTrustFederation/AIFT-Runtime": (
                "2026-10-25T01:18:37Z", "2026-10-25T01:18:49Z", "2026-10-25T01:18:46Z"),
        }
        for entry in report["github_runs"]:
            started, updated, created = expected[entry["repository"]]
            self.assertEqual(entry["workflow_started_at"], started)
            self.assertEqual(entry["workflow_updated_at"], updated)
            self.assertEqual(entry["artifact_created_at"], created)
            self.assertIs(type(entry["artifact_size_in_bytes"]), int)
            self.assertGreater(entry["artifact_size_in_bytes"], 0)
            self.assertLessEqual(entry["artifact_size_in_bytes"], 4 * 1024 * 1024)

    def test_assessment_records_unsigned_verifier_source_fingerprints(self):
        result, output = self.invoke()
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads((output / "federation-assessment.json").read_text())
        verifier = report["verifier_implementation"]
        self.assertEqual(verifier["attestation"], "local_file_sha256_unsigned")
        for name, filename in (
            ("collector", "collect_federation_evidence.py"),
            ("structural_consumer", "check_federation_evidence.py"),
        ):
            source = ROOT / "scripts" / filename
            self.assertEqual(verifier[name]["path"], f"scripts/{filename}")
            self.assertEqual(verifier[name]["sha256"], hashlib.sha256(source.read_bytes()).hexdigest())
        self.assertIs(report["federation_scope"]["federation_wide_acceptance"], False)

    def test_assessment_exposes_reported_checks_gates_and_limitations(self):
        result, output = self.invoke()
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads((output / "federation-assessment.json").read_text())
        for entry in report["repositories"]:
            self.assertEqual(entry["reported_checks"], [
                {"id": "qa", "command": "python -m unittest", "status": "passed", "exit_code": 0}
            ])
            self.assertEqual(entry["reported_gates"], {
                "security": "not_checked", "dependencies": "not_checked",
                "deployment": "not_applicable",
            })
            self.assertEqual(entry["reported_limitations"], ["No live deployment tested."])

    def test_failed_checks_remain_visible_in_assessment(self):
        result, output = self.invoke("failed_run_with_failed_check")
        self.assertEqual(result.returncode, 1, result.stderr)
        report = json.loads((output / "federation-assessment.json").read_text())
        self.assertEqual(report["status"], "failed")
        for entry in report["repositories"]:
            self.assertEqual(entry["reported_checks"][0]["status"], "failed")
            self.assertEqual(entry["reported_checks"][0]["exit_code"], 1)

    def test_subset_collection_cannot_report_full_federation_pass(self):
        only_repo = "AIFreedomTrustFederation/.github"
        result, output = self.invoke("all_gates_pass", selected_repos={only_repo})
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads((output / "federation-assessment.json").read_text())
        self.assertEqual(report["status"], "partial")
        self.assertEqual(report["producer_coverage"]["collected"], [only_repo])
        self.assertEqual(report["producer_coverage"]["not_collected"],
                         ["AIFreedomTrustFederation/AIFT-Runtime"])
        self.assertEqual(len(report["repositories"]), 1)

    def test_complete_supported_coverage_can_report_pass(self):
        result, output = self.invoke("all_gates_pass")
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads((output / "federation-assessment.json").read_text())
        self.assertEqual(report["status"], "reported_pass")
        self.assertEqual(report["producer_coverage"]["collected"], sorted(REPOS))
        self.assertEqual(report["producer_coverage"]["not_collected"], [])

    def test_reported_pass_is_explicitly_not_federation_wide_acceptance(self):
        result, output = self.invoke("all_gates_pass")
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads((output / "federation-assessment.json").read_text())
        self.assertEqual(report["status"], "reported_pass")
        scope = report["federation_scope"]
        self.assertEqual(scope["verification_level"], "supported_producers_only")
        self.assertEqual(scope["supported_producers"], sorted(REPOS))
        self.assertIs(scope["federation_wide_acceptance"], False)
        self.assertIs(scope["known_unintegrated_repositories_exhaustive"], False)
        self.assertTrue(set(scope["known_unintegrated_repositories"]).isdisjoint(REPOS))
        for name in ("AIFT-Genesis", "AIFT-Forge", "TheMindofAll", "VPS",
                     "AI-Freedom-Trust", "www.aifreedomtrust.com"):
            self.assertIn(f"AIFreedomTrustFederation/{name}",
                          scope["known_unintegrated_repositories"])

    def test_publishes_authenticated_failed_security_gate(self):
        result, output = self.invoke("failed_gate")
        self.assertEqual(result.returncode, 1, result.stderr)
        report = json.loads((output / "federation-assessment.json").read_text())
        self.assertEqual(report["status"], "failed")
        self.assertTrue(all("security" in r["failed_gates"] for r in report["repositories"]))
        self.assertEqual(len(list(output.iterdir())), 3)

    def test_publishes_authenticated_failed_check(self):
        result, output = self.invoke("failed_check")
        self.assertEqual(result.returncode, 1, result.stderr)
        report = json.loads((output / "federation-assessment.json").read_text())
        self.assertEqual(report["status"], "failed")
        self.assertTrue(all(r["status"] == "failed" for r in report["repositories"]))
        self.assertEqual(len(list(output.iterdir())), 3)

    def test_rejects_future_generated_at(self):
        result, output = self.invoke("future_generated_at")
        self.assertEqual(result.returncode, 1, result.stderr)
        self.assertIn("verification evidence timestamp is in the future", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_stale_generated_at_relative_to_authenticated_run(self):
        result, output = self.invoke("stale_generated_at")
        self.assertEqual(result.returncode, 1, result.stderr)
        self.assertIn("timestamp outside authenticated workflow window", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_post_run_generated_at(self):
        result, output = self.invoke("post_run_generated_at")
        self.assertEqual(result.returncode, 1, result.stderr)
        self.assertIn("timestamp outside authenticated workflow window", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_missing_authenticated_run_timestamp(self):
        result, output = self.invoke("missing_run_timestamps")
        self.assertEqual(result.returncode, 1, result.stderr)
        self.assertIn("invalid authenticated workflow timestamps", result.stderr)
        self.assertFalse(output.exists())

    def test_records_authenticated_run_attempt(self):
        result, output = self.invoke()
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads((output / "federation-assessment.json").read_text())
        self.assertTrue(all(run["run_attempt"] == 1 for run in report["github_runs"]))

    def test_rejects_invalid_run_attempt(self):
        result, output = self.invoke("invalid_run_attempt")
        self.assertEqual(result.returncode, 1, result.stderr)
        self.assertIn("invalid workflow run attempt", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_artifact_created_outside_authenticated_attempt(self):
        for mode in ("artifact_created_before_run", "artifact_created_after_run"):
            with self.subTest(mode=mode):
                result, output = self.invoke(mode, output=self.root / mode)
                self.assertEqual(result.returncode, 1, result.stderr)
                self.assertIn("artifact created outside authenticated workflow window", result.stderr)
                self.assertFalse(output.exists())

    def test_rejects_artifact_one_second_outside_authenticated_run(self):
        # Both timestamps are authenticated GitHub metadata; accepting a
        # pre-start artifact within a clock-skew grace period could accept
        # evidence from a previous attempt of the same workflow run.
        for mode in ("artifact_created_just_before_run", "artifact_created_just_after_run"):
            with self.subTest(mode=mode):
                result, output = self.invoke(mode, output=self.root / mode)
                self.assertEqual(result.returncode, 1, result.stderr)
                self.assertIn("artifact created outside authenticated workflow window", result.stderr)
                self.assertFalse(output.exists())

    def test_rejects_missing_artifact_creation_timestamp(self):
        result, output = self.invoke("artifact_missing_created_at")
        self.assertEqual(result.returncode, 1, result.stderr)
        self.assertIn("invalid authenticated artifact timestamp", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_stale_default_branch(self):
        result, output = self.invoke("stale")
        self.assertEqual(result.returncode, 1)
        self.assertIn("not current", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_branch_advance_during_artifact_download(self):
        result, output = self.invoke("branch_advanced")
        self.assertEqual(result.returncode, 1)
        self.assertIn("default branch advanced during collection", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_repository_identity_change_during_collection(self):
        # A repository can be renamed or transferred while its SHA stays the same.
        # The final metadata recheck must preserve the original numeric identity.
        result, output = self.invoke("repo_id_changed")
        self.assertEqual(result.returncode, 1, result.stderr)
        self.assertIn("repository identity changed during collection", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_wrong_workflow_id(self):
        result, output = self.invoke("wrong_workflow")
        self.assertEqual(result.returncode, 1)
        self.assertIn("workflow identity mismatch", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_wrong_workflow_path(self):
        result, output = self.invoke("wrong_workflow_path")
        self.assertEqual(result.returncode, 1)
        self.assertIn("workflow identity mismatch", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_mismatched_or_missing_workflow_run_id(self):
        for mode in ("wrong_run_id", "missing_run_id"):
            with self.subTest(mode=mode):
                result, output = self.invoke(mode, output=self.root / mode)
                self.assertEqual(result.returncode, 1, result.stderr)
                self.assertIn("run identity mismatch", result.stderr)
                self.assertFalse(output.exists())

    def test_rejects_wrong_run_repository_identity(self):
        for mode in ("wrong_run_repo_id", "wrong_run_head_repo_id",
                     "missing_run_head_repo"):
            with self.subTest(mode=mode):
                result, output = self.invoke(mode, output=self.root / mode)
                self.assertEqual(result.returncode, 1, result.stderr)
                self.assertIn("run metadata does not establish", result.stderr)
                self.assertFalse(output.exists())

    def test_rejects_failed_run(self):
        result, output = self.invoke("failed_run")
        self.assertEqual(result.returncode, 1)
        self.assertIn("failed CI run lacks failed verification evidence", result.stderr)
        self.assertFalse(output.exists())

    def test_preserves_negative_security_evidence_from_failed_ci_run(self):
        # Both supported producers upload the verification JSON even on QA failure.
        result, output = self.invoke("failed_run_with_failed_gate")
        self.assertEqual(result.returncode, 1, result.stderr)
        report = json.loads((output / "federation-assessment.json").read_text())
        self.assertEqual(report["status"], "failed")
        self.assertTrue(all("security" in r["failed_gates"] for r in report["repositories"]))

    def test_preserves_negative_check_evidence_from_failed_ci_run(self):
        result, output = self.invoke("failed_run_with_failed_check")
        self.assertEqual(result.returncode, 1, result.stderr)
        report = json.loads((output / "federation-assessment.json").read_text())
        self.assertEqual(report["status"], "failed")
        self.assertTrue(all(r["status"] == "failed" for r in report["repositories"]))

    def test_rejects_artifact_bound_to_different_run(self):
        result, output = self.invoke("artifact_wrong_run_id")
        self.assertEqual(result.returncode, 1)
        self.assertIn("artifact workflow run identity mismatch", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_artifact_bound_to_different_commit(self):
        result, output = self.invoke("artifact_wrong_run_sha")
        self.assertEqual(result.returncode, 1)
        self.assertIn("artifact workflow run identity mismatch", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_artifact_bound_to_different_branch(self):
        result, output = self.invoke("artifact_wrong_run_branch")
        self.assertEqual(result.returncode, 1)
        self.assertIn("artifact workflow run identity mismatch", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_artifact_from_wrong_repository(self):
        for mode in ("artifact_wrong_repo_id", "artifact_wrong_head_repo_id"):
            with self.subTest(mode=mode):
                result, output = self.invoke(mode, output=self.root / mode)
                self.assertEqual(result.returncode, 1)
                self.assertIn("artifact workflow run identity mismatch", result.stderr)
                self.assertFalse(output.exists())

    def test_rejects_artifact_missing_workflow_run_provenance(self):
        result, output = self.invoke("artifact_missing_run_identity")
        self.assertEqual(result.returncode, 1)
        self.assertIn("artifact workflow run identity mismatch", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_noninteger_artifact_listing_count(self):
        # Python considers True == 1 and 1.0 == 1; neither is a valid
        # GitHub API integer count, even when one artifact is listed.
        for mode in ("artifact_count_boolean", "artifact_count_float"):
            with self.subTest(mode=mode):
                result, output = self.invoke(mode, output=self.root / f"evidence-{mode}")
                self.assertEqual(result.returncode, 1, result.stderr)
                self.assertIn("artifact list incomplete", result.stderr)
                self.assertFalse(output.exists())

    def test_rejects_duplicate_artifact(self):
        result, output = self.invoke("duplicate_artifact")
        self.assertEqual(result.returncode, 1)
        self.assertIn("duplicate", result.stderr)
        self.assertFalse(output.exists())

    def test_collector_rejects_malformed_producer_json_through_strict_consumer(self):
        # Integration regression: the collector must use the strict consumer,
        # not just pass its own API/digest checks. The fake API computes the
        # correct SHA-256 and size for each malformed producer archive.
        for mode in ("nonfinite_producer_json", "deeply_nested_producer_json"):
            with self.subTest(mode=mode):
                result, output = self.invoke(mode, output=self.root / f"evidence-{mode}")
                self.assertEqual(result.returncode, 1, result.stderr)
                self.assertIn("invalid JSON", result.stderr)
                self.assertNotIn("Traceback (most recent call last)", result.stderr)
                self.assertFalse(output.exists())

    def test_rejects_mismatched_artifact_run_url(self):
        result, output = self.invoke("tampered")
        self.assertEqual(result.returncode, 1)
        self.assertIn("artifact source", result.stderr)
        self.assertFalse(output.exists())


    def test_rejects_unlisted_artifact_id(self):
        result, output = self.invoke("wrong_artifact_id")
        self.assertEqual(result.returncode, 1)
        self.assertIn("artifact ID download failed", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_invalid_zip(self):
        result, output = self.invoke("corrupt_archive")
        self.assertEqual(result.returncode, 1)
        self.assertIn("invalid artifact ZIP", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_corrupted_deflate_without_traceback(self):
        # The simulated GitHub API reports the digest and length of the damaged
        # ZIP itself, so only decompression error handling can reject this input.
        result, output = self.invoke("invalid_deflate")
        self.assertEqual(result.returncode, 1, result.stderr)
        self.assertIn("invalid artifact ZIP", result.stderr)
        self.assertNotIn("Traceback (most recent call last)", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_extra_zip_member(self):
        result, output = self.invoke("extra_member")
        self.assertEqual(result.returncode, 1)
        self.assertIn("exactly one file", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_oversized_api_metadata_before_json_parsing(self):
        result, output = self.invoke("oversized_metadata")
        self.assertEqual(result.returncode, 1)
        self.assertIn("GitHub API metadata", result.stderr)
        self.assertIn("too large", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_non_utf8_api_metadata(self):
        result, output = self.invoke("invalid_utf8_metadata")
        self.assertEqual(result.returncode, 1)
        self.assertIn("GitHub API metadata is not UTF-8", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_deeply_nested_api_metadata_without_traceback(self):
        # A 40 KiB GitHub API response can exceed Python JSON parser depth.
        # Fail with a controlled collection error, not an uncaught traceback.
        result, output = self.invoke("deeply_nested_metadata")
        self.assertEqual(result.returncode, 1, result.stderr)
        self.assertIn("invalid GitHub API response", result.stderr)
        self.assertNotIn("Traceback (most recent call last)", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_overflowing_api_exponents_without_traceback(self):
        # Python json.loads otherwise silently converts 1e9999 to infinity.
        # Even unused metadata must be valid finite JSON before provenance checks.
        for mode in ("overflowing_exponent_metadata", "negative_overflowing_exponent_metadata"):
            with self.subTest(mode=mode):
                result, output = self.invoke(mode, output=self.root / f"evidence-{mode}")
                self.assertEqual(result.returncode, 1, result.stderr)
                self.assertIn("non-finite GitHub API number", result.stderr)
                self.assertNotIn("Traceback (most recent call last)", result.stderr)
                self.assertFalse(output.exists())

    def test_accepts_finite_api_exponent(self):
        result, output = self.invoke("finite_exponent_metadata", output=self.root / "evidence-finite")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertTrue(output.is_dir())

    def test_rejects_oversized_artifact_before_download(self):
        result, output = self.invoke("oversized_artifact")
        self.assertEqual(result.returncode, 1)
        self.assertIn("oversized artifact size", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_oversized_actual_download_even_when_metadata_is_small(self):
        # The advertised artifact size is valid, but the subprocess emits >4 MiB.
        # The collector must stop reading before buffering the whole response.
        result, output = self.invoke("oversized_stream")
        self.assertEqual(result.returncode, 1)
        self.assertIn("artifact archive missing or too large", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_inconsistent_artifact_size_even_with_matching_digest(self):
        # An otherwise valid archive must match its authenticated size metadata.
        result, output = self.invoke("wrong_artifact_size")
        self.assertEqual(result.returncode, 1, result.stderr)
        self.assertIn("artifact archive size does not match metadata", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_missing_artifact_size(self):
        result, output = self.invoke("missing_artifact_size")
        self.assertEqual(result.returncode, 1)
        self.assertIn("missing, invalid, or oversized artifact size", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_invalid_artifact_size_type(self):
        result, output = self.invoke("invalid_artifact_size")
        self.assertEqual(result.returncode, 1)
        self.assertIn("missing, invalid, or oversized artifact size", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_artifact_digest_mismatch(self):
        result, output = self.invoke("digest_mismatch")
        self.assertEqual(result.returncode, 1)
        self.assertIn("artifact SHA-256 digest mismatch", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_missing_artifact_digest(self):
        result, output = self.invoke("missing_digest")
        self.assertEqual(result.returncode, 1)
        self.assertIn("missing or invalid artifact SHA-256 digest", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_malformed_artifact_digest(self):
        result, output = self.invoke("invalid_digest")
        self.assertEqual(result.returncode, 1)
        self.assertIn("missing or invalid artifact SHA-256 digest", result.stderr)
        self.assertFalse(output.exists())

    def test_atomic_publication_leaves_no_partial_directory_on_write_error(self):
        # Exercise the publisher with a deterministic mid-publication disk failure.
        sys.path.insert(0, str(ROOT / "scripts"))
        try:
            spec = importlib.util.spec_from_file_location("_collector_publication_test", SCRIPT)
            collector = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(collector)
        finally:
            sys.path.remove(str(ROOT / "scripts"))
        records, provenance = [], []
        for index, repo in enumerate(REPOS):
            directory = self.root / f"source-{index}"
            directory.mkdir()
            source = directory / ("aift-catalog-verification.json" if repo.endswith("/.github")
                                  else "aift-runtime-verification.json")
            source.write_text("{}", encoding="utf-8")
            records.append(source)
            provenance.append({"repository": repo})
        output = self.root / "atomic-assessment"
        original_write = Path.write_bytes

        def fail_second_record(path, data):
            if (path.name == "aift-runtime-verification.json" and
                    ".aift-evidence-stage-" in str(path)):
                raise OSError("simulated disk-full during publication")
            return original_write(path, data)

        with mock.patch.object(Path, "write_bytes", fail_second_record):
            with self.assertRaisesRegex(OSError, "simulated disk-full"):
                collector.publish_records(output, records, provenance, {"status": "partial"})
        self.assertFalse(output.exists(), "incomplete assessment must never be published")
        self.assertEqual(list(self.root.glob(".aift-evidence-stage-*")), [])

        # A subsequent successful publication is complete and leaves no staging debris.
        collector.publish_records(output, records, provenance, {"status": "partial"})
        self.assertEqual({p.name for p in output.iterdir()},
                         {"aift-catalog-verification.json", "aift-runtime-verification.json",
                          "federation-assessment.json"})
        self.assertEqual(list(self.root.glob(".aift-evidence-stage-*")), [])


    def test_publication_rejects_missing_and_duplicate_provenance(self):
        # The publisher must not silently truncate mismatched lists via zip().
        sys.path.insert(0, str(ROOT / "scripts"))
        try:
            spec = importlib.util.spec_from_file_location("_collector_invariant_test", SCRIPT)
            collector = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(collector)
        finally:
            sys.path.remove(str(ROOT / "scripts"))
        sources = []
        for index in range(2):
            source = self.root / f"publisher-source-{index}.json"
            source.write_text("{}", encoding="utf-8")
            sources.append(source)
        repositories = list(REPOS)
        for case, records, provenance in (
            ("missing", sources[:1], [{"repository": repo} for repo in repositories]),
            ("duplicate", sources, [{"repository": repositories[0]}] * 2),
        ):
            with self.subTest(case=case):
                output = self.root / f"publisher-{case}"
                with self.assertRaisesRegex(collector.CollectionError, "incomplete or duplicate"):
                    collector.publish_records(output, records, provenance, {"status": "partial"})
                self.assertFalse(output.exists())
                self.assertEqual(list(self.root.glob(".aift-evidence-stage-*")), [])

    def test_publication_rejects_swapped_source_filenames(self):
        # Equal-sized source/provenance lists are insufficient if their
        # ordering differs; never publish a producer's file under another name.
        sys.path.insert(0, str(ROOT / "scripts"))
        try:
            spec = importlib.util.spec_from_file_location("_collector_mapping_test", SCRIPT)
            collector = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(collector)
        finally:
            sys.path.remove(str(ROOT / "scripts"))
        sources, provenance = [], []
        for index, repo in enumerate(REPOS):
            directory = self.root / f"mapping-source-{index}"
            directory.mkdir()
            source = directory / collector.ARTIFACTS[repo][1]
            source.write_text("{}", encoding="utf-8")
            sources.append(source)
            provenance.append({"repository": repo})
        output = self.root / "swapped-assessment"
        with self.assertRaisesRegex(collector.CollectionError,
                                    "publication source does not match repository"):
            collector.publish_records(output, sources[::-1], provenance,
                                      {"status": "partial"})
        self.assertFalse(output.exists())
        self.assertEqual(list(self.root.glob(".aift-evidence-stage-*")), [])


    def test_rejects_checkout_local_output_before_download(self):
        # A successful collector run must never create files in the source
        # checkout after the workflow's clean-checkout verification.
        with tempfile.TemporaryDirectory(prefix=".aift-collector-boundary-", dir=ROOT) as directory:
            direct = Path(directory) / "evidence"
            alias = self.root / "checkout-alias"
            alias.symlink_to(directory, target_is_directory=True)
            for output in (direct, alias / "evidence"):
                with self.subTest(output=str(output)):
                    result, _ = self.invoke(output=output)
                    self.assertEqual(result.returncode, 1, result.stderr)
                    self.assertIn("output directory must be outside the collector checkout",
                                  result.stderr)
                    self.assertFalse(direct.exists())


    def test_never_overwrites_existing_directory(self):
        target = self.root / "evidence"
        target.mkdir()
        (target / "sentinel").write_text("preserve")
        result, _ = self.invoke(output=target)
        self.assertEqual(result.returncode, 1)
        self.assertEqual((target / "sentinel").read_text(), "preserve")


    def test_publication_race_preserves_concurrently_created_empty_destination(self):
        # Deterministically create the destination after the publisher's final
        # exists() check but before the atomic no-clobber rename.
        sys.path.insert(0, str(ROOT / "scripts"))
        try:
            spec = importlib.util.spec_from_file_location("_collector_race_test", SCRIPT)
            collector = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(collector)
        finally:
            sys.path.remove(str(ROOT / "scripts"))
        records, provenance = [], []
        for index, repo in enumerate(REPOS):
            directory = self.root / f"race-source-{index}"
            directory.mkdir()
            source = directory / collector.ARTIFACTS[repo][1]
            source.write_text("{}", encoding="utf-8")
            records.append(source)
            provenance.append({"repository": repo})
        output = self.root / "raced-assessment"
        actual_rename = collector.rename_directory_no_replace

        def create_destination_before_rename(stage, destination):
            destination.mkdir()
            return actual_rename(stage, destination)

        with mock.patch.object(collector, "rename_directory_no_replace",
                               side_effect=create_destination_before_rename):
            with self.assertRaisesRegex(collector.CollectionError, "already exists"):
                collector.publish_records(output, records, provenance, {"status": "reported_pass"})
        self.assertTrue(output.is_dir())
        self.assertEqual(list(output.iterdir()), [], "existing empty directory was replaced")
        self.assertEqual(list(self.root.glob(".aift-evidence-stage-*")), [])

        # The same publisher still succeeds normally after the collision.
        output.rmdir()
        collector.publish_records(output, records, provenance, {"status": "reported_pass"})
        self.assertEqual({p.name for p in output.iterdir()},
                         {"aift-catalog-verification.json", "aift-runtime-verification.json",
                          "federation-assessment.json"})
        self.assertEqual(list(self.root.glob(".aift-evidence-stage-*")), [])


    def test_live_collector_authenticates_only_on_default_branch_events(self):
        # Same-repository PRs can also modify the collector. No PR-controlled
        # code should execute in a job with GH_TOKEN, even when read-only.
        workflow = (ROOT / ".github/workflows/federation-evidence-collector.yml").read_text(encoding="utf-8")
        self.assertIn("  push:\n    branches: [main]\n", workflow)
        self.assertIn("  pull_request:\n", workflow)
        self.assertIn("  schedule:\n    - cron: '23 * * * *'\n", workflow)
        live_job = workflow.split("\n  live-producer-evidence:\n", 1)[1]
        guards = [line.strip().removeprefix("if: ") for line in live_job.splitlines()
                  if line.startswith("    if: ")]
        self.assertEqual(guards, [
            "github.event_name == 'push' || github.event_name == 'schedule'"
        ])
        self.assertNotIn("github.event.pull_request", live_job)
        self.assertIn("    needs: collector\n", live_job)
        self.assertIn("      actions: read\n", live_job)
        self.assertIn("          GH_TOKEN: ${{ steps.producer-token.outputs.token }}\n", live_job)


    def test_live_collector_retains_verified_negative_producer_evidence(self):
        # A producer may upload valid failed-check evidence from an unsuccessful run.
        # Keep CI red, but retain only records that passed the assessment validation.
        workflow = (ROOT / ".github/workflows/federation-evidence-collector.yml").read_text(encoding="utf-8")
        live_job = workflow.split("\n  live-producer-evidence:\n", 1)[1]
        self.assertIn(r'(.conclusion == \"success\" or .conclusion == \"failure\")', live_job)
        self.assertIn("        id: collect\n", live_job)
        self.assertIn("          collector_status=$?\n", live_job)
        self.assertIn('          echo "validated=true" >> "$GITHUB_OUTPUT"\n', live_job)
        self.assertIn('          exit "$collector_status"\n', live_job)
        self.assertIn("        if: ${{ always() && steps.collect.outputs.validated == 'true' }}\n", live_job)


    def test_live_workflow_shell_retains_failed_producer_assessment(self):
        # Execute the workflow's actual post-collection shell, not just its
        # YAML text. A verified producer failure must remain red while setting
        # the output that permits the validated JSON files to be retained.
        workflow = (ROOT / ".github/workflows/federation-evidence-collector.yml").read_text(
            encoding="utf-8"
        )
        start = "          set +e\n"
        end = '          exit "$collector_status"\n'
        self.assertEqual(workflow.count(start), 1)
        self.assertEqual(workflow.count(end), 1)
        tail = workflow.split(start, 1)[1].split(end, 1)[0]
        shell = textwrap.dedent(start + tail + end)
        args = " ".join(
            f'--run "{repo}={run}" --expect "{repo}={sha}"'
            for repo, (run, sha, _, _, _) in REPOS.items()
        )

        for mode, validated in (
            ("failed_run_with_failed_check", True),
            ("digest_mismatch", False),
        ):
            with self.subTest(mode=mode):
                # Prepare the same fixture files consumed by the fake gh CLI.
                self.invoke(mode, output=self.root / f"preflight-{mode}")
                runner_temp = self.root / mode
                runner_temp.mkdir()
                github_output = runner_temp / "github-output"
                env = dict(
                    os.environ,
                    PATH=f"{self.bin}:{os.environ.get('PATH', '')}",
                    FAKE_GH_MODE=mode,
                    FAKE_GH_FIXTURES=str(self.root / "fixtures"),
                    RUNNER_TEMP=str(runner_temp),
                    GITHUB_OUTPUT=str(github_output),
                )
                result = subprocess.run(
                    ["bash", "-c", f"set -euo pipefail\nargs=({args})\n" + shell],
                    cwd=ROOT, env=env, capture_output=True, text=True, check=False,
                )
                self.assertEqual(result.returncode, 1, result.stderr)
                output = runner_temp / "federation-evidence-live"
                if validated:
                    self.assertEqual(github_output.read_text(), "validated=true\n")
                    self.assertEqual(
                        {p.name for p in output.iterdir()},
                        {"federation-assessment.json",
                         "aift-catalog-verification.json",
                         "aift-runtime-verification.json"},
                    )
                    assessment = json.loads(
                        (output / "federation-assessment.json").read_text()
                    )
                    self.assertEqual(assessment["status"], "failed")
                    self.assertEqual(assessment["errors"], [])
                else:
                    self.assertFalse(github_output.exists())
                    self.assertFalse(output.exists())


    def test_rejects_oversized_numeric_run_id_without_traceback(self):
        # A 5,000-digit ASCII run ID previously raised Python's integer
        # conversion ValueError outside the collector's handled error path.
        output = self.root / "oversized-run-id-evidence"
        repo, (_, sha, _, _, _) = next(iter(REPOS.items()))
        for run_id in ("9" * 5000, "0" * 5000, "0"):
            with self.subTest(run_id_length=len(run_id), first=run_id[0]):
                result = subprocess.run(
                    [sys.executable, "-B", str(SCRIPT),
                     "--run", f"{repo}={run_id}",
                     "--expect", f"{repo}={sha}",
                     "--output-dir", str(output)],
                    capture_output=True, text=True, check=False,
                )
                self.assertEqual(result.returncode, 1, result.stderr)
                self.assertIn("invalid or duplicate run", result.stderr)
                self.assertNotIn("Traceback", result.stderr)
                self.assertFalse(output.exists())


if __name__ == "__main__":
    unittest.main()
