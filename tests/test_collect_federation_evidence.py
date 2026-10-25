#!/usr/bin/env python3
"""Black-box tests of the authenticated GitHub Actions evidence collector."""

import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts/collect_federation_evidence.py"
REPOS = {
    "AIFreedomTrustFederation/.github": ("40146248552", "b14a458e6e6ef4bbda5cde96ee445e66678fee9c", "aift-catalog-verification", "aift-catalog-verification.json", 12580921836),
    "AIFreedomTrustFederation/AIFT-Runtime": ("40150125926", "d953ddccacdb5700d46e2d79d5af7c7dd309807f", "aift-runtime-verification", "aift-runtime-verification.json", 12582097262),
}

FAKE_GH = r'''#!/usr/bin/env python3
import json,os,sys,shutil
from pathlib import Path
args=sys.argv[1:]
repo_by_id={"40146248552":"AIFreedomTrustFederation/.github","40150125926":"AIFreedomTrustFederation/AIFT-Runtime"}
repos={
"AIFreedomTrustFederation/.github":("40146248552","b14a458e6e6ef4bbda5cde96ee445e66678fee9c","aift-catalog-verification", "aift-catalog-verification.json",12580921836),
"AIFreedomTrustFederation/AIFT-Runtime":("40150125926","d953ddccacdb5700d46e2d79d5af7c7dd309807f","aift-runtime-verification","aift-runtime-verification.json",12582097262),
}
mode=os.environ.get("FAKE_GH_MODE", "ok")
if args[0]=="api":
 path=args[1].removeprefix("repos/")
 repo=next((r for r in repos if path==r or path.startswith(r+"/")),None)
 if not repo: sys.exit(2)
 run,sha,artifact,filename,artifact_id=repos[repo]
 suffix=path[len(repo):]
 if suffix=="": result={"default_branch":"main"}
 elif suffix=="/branches/main": result={"commit":{"sha":"0"*40 if mode=="stale" else sha}}
 elif suffix=="/actions/runs/"+run:
  result={"repository":{"full_name":repo},"head_sha":sha,"head_branch":"main","event":"push","status":"completed","conclusion":"failure" if mode=="failed_run" else "success"}
 elif suffix=="/actions/runs/"+run+"/artifacts?per_page=100":
  items=[{"name":artifact,"id":artifact_id,"expired":False}]
  if mode=="duplicate_artifact":items*=2
  result={"total_count":len(items),"artifacts":items}
 else: sys.exit(3)
 print(json.dumps(result))
elif args[:2]==["run","download"]:
 run=args[2]; repo=args[args.index("--repo")+1]; out=Path(args[args.index("--dir")+1]); filename=repos[repo][3]
 if repo_by_id[run]!=repo:sys.exit(4)
 src=Path(os.environ["FAKE_GH_FIXTURES"])/filename
 if mode=="tampered" and repo.endswith(".github"):
  data=json.loads(src.read_text());data["source"]["run_url"]="https://github.com/AIFreedomTrustFederation/.github/actions/runs/123"
  (out/filename).write_text(json.dumps(data))
 else:shutil.copyfile(src,out/filename)
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

    def invoke(self, mode="ok", output=None):
        fixture_dir = self.root / "fixtures"
        fixture_dir.mkdir(exist_ok=True)
        for repo, (run, sha, _, filename, _) in REPOS.items():
            record = {"schema": "aift.verification.v1", "repository": repo, "commit_sha": sha,
                      "generated_at": "2026-10-25T01:00:00+00:00",
                      "source": {"type": "github-actions", "run_url": f"https://github.com/{repo}/actions/runs/{run}"},
                      "checks": [{"id": "qa", "command": "python -m unittest", "status": "passed", "exit_code": 0}],
                      "overall_status": "passed", "security": {"status": "not_checked"},
                      "dependencies": {"status": "not_checked"}, "deployment": {"status": "not_applicable"},
                      "limitations": ["No live deployment tested."]}
            (fixture_dir / filename).write_text(json.dumps(record), encoding="utf-8")
        output = output or self.root / "evidence"
        cmd = [sys.executable, "-B", str(SCRIPT), "--output-dir", str(output)]
        for repo, (run, sha, _, _, _) in REPOS.items():
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
        self.assertEqual(len(list(output.iterdir())), 3)

    def test_rejects_stale_default_branch(self):
        result, output = self.invoke("stale")
        self.assertEqual(result.returncode, 1)
        self.assertIn("not current", result.stderr)
        self.assertFalse(output.exists())

    def test_rejects_failed_run(self):
        result, output = self.invoke("failed_run")
        self.assertEqual(result.returncode, 1)
        self.assertIn("successful default-branch CI", result.stderr)
        self.assertFalse(output.exists())

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

    def test_never_overwrites_existing_directory(self):
        target = self.root / "evidence"
        target.mkdir()
        (target / "sentinel").write_text("preserve")
        result, _ = self.invoke(output=target)
        self.assertEqual(result.returncode, 1)
        self.assertEqual((target / "sentinel").read_text(), "preserve")


if __name__ == "__main__":
    unittest.main()
