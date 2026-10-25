#!/usr/bin/env python3
"""Black-box tests of the authenticated GitHub Actions evidence collector."""

import json
import importlib.util
import os
import subprocess
import sys
import tempfile
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
 data=(Path(os.environ["FAKE_GH_FIXTURES"])/filename).read_bytes()
 if mode=="tampered" and repo.endswith("/.github"):
  record=json.loads(data);record["source"]["run_url"]="https://github.com/AIFreedomTrustFederation/.github/actions/runs/123"
  data=json.dumps(record).encode()
 buffer=io.BytesIO()
 with zipfile.ZipFile(buffer,"w",compression=zipfile.ZIP_DEFLATED) as archive:
  info=zipfile.ZipInfo(filename,date_time=(2020,1,1,0,0,0));info.compress_type=zipfile.ZIP_DEFLATED
  archive.writestr(info,data)
  if mode=="extra_member":
   extra=zipfile.ZipInfo("unexpected.txt",date_time=(2020,1,1,0,0,0));extra.compress_type=zipfile.ZIP_DEFLATED
   archive.writestr(extra,b"untrusted")
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
 if suffix=="": result={"default_branch":"main"}
 elif suffix=="/branches/main":
  advanced=mode=="branch_advanced" and (Path(os.environ["FAKE_GH_FIXTURES"])/".branch-advanced").exists()
  result={"commit":{"sha":"0"*40 if mode=="stale" or advanced else sha}}
 elif suffix=="/actions/runs/"+run:
  result={"repository":{"full_name":repo},"head_sha":sha,"head_branch":"main","event":"push","status":"completed","conclusion":"failure" if mode in ("failed_run", "failed_run_with_failed_gate", "failed_run_with_failed_check") else "success",
          "workflow_id":362136764 if repo.endswith("/.github") else 361044013,
          "path":".github/workflows/catalog-integrity.yml" if repo.endswith("/.github") else ".github/workflows/qa.yml"}
  if mode=="wrong_workflow":result["workflow_id"]=999
  if mode=="wrong_workflow_path":result["path"]=".github/workflows/untrusted.yml"
 elif suffix=="/actions/runs/"+run+"/artifacts?per_page=100":
  digest="sha256:"+hashlib.sha256(archive_bytes(repo)).hexdigest()
  if mode=="digest_mismatch":digest="sha256:"+"0"*64
  if mode=="invalid_digest":digest="sha256:not-a-digest"
  item={"name":artifact,"id":artifact_id,"expired":False,"digest":digest,"size_in_bytes":len(archive_bytes(repo))}
  if mode=="oversized_artifact":item["size_in_bytes"]=5*1024*1024
  if mode=="missing_artifact_size":item.pop("size_in_bytes")
  if mode=="invalid_artifact_size":item["size_in_bytes"]=True
  if mode=="missing_digest":item.pop("digest")
  items=[item]
  if mode=="duplicate_artifact":items*=2
  if mode=="wrong_artifact_id":items[0]["id"]+=1
  result={"total_count":len(items),"artifacts":items}
 elif suffix=="/actions/artifacts/"+str(artifact_id)+"/zip":
  if mode=="oversized_stream":
   sys.stdout.buffer.write(b"x"*(5*1024*1024))
   sys.exit(0)
  sys.stdout.buffer.write(archive_bytes(repo))
  if mode=="branch_advanced" and repo.endswith("/AIFT-Runtime"): (Path(os.environ["FAKE_GH_FIXTURES"])/".branch-advanced").write_text("advanced")
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
                      "generated_at": "2026-10-25T01:00:00+00:00",
                      "source": {"type": "github-actions", "run_url": f"https://github.com/{repo}/actions/runs/{run}"},
                      "checks": [{"id": "qa", "command": "python -m unittest", "status": "passed", "exit_code": 0}],
                      "overall_status": "passed", "security": {"status": "passed" if mode == "all_gates_pass" else "not_checked"},
                      "dependencies": {"status": "passed" if mode == "all_gates_pass" else "not_checked"},
                      "deployment": {"status": "passed" if mode == "all_gates_pass" else "not_applicable"},
                      "limitations": ["No live deployment tested."]}
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
        expected_workflows = {
            "AIFreedomTrustFederation/.github": (362136764, ".github/workflows/catalog-integrity.yml"),
            "AIFreedomTrustFederation/AIFT-Runtime": (361044013, ".github/workflows/qa.yml"),
        }
        for run in report["github_runs"]:
            self.assertEqual((run["workflow_id"], run["workflow_path"]),
                             expected_workflows[run["repository"]])
        self.assertEqual(len(list(output.iterdir())), 3)

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

    def test_rejects_duplicate_artifact(self):
        result, output = self.invoke("duplicate_artifact")
        self.assertEqual(result.returncode, 1)
        self.assertIn("duplicate", result.stderr)
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
            source = self.root / f"source-{index}.json"
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


    def test_never_overwrites_existing_directory(self):
        target = self.root / "evidence"
        target.mkdir()
        (target / "sentinel").write_text("preserve")
        result, _ = self.invoke(output=target)
        self.assertEqual(result.returncode, 1)
        self.assertEqual((target / "sentinel").read_text(), "preserve")


if __name__ == "__main__":
    unittest.main()
