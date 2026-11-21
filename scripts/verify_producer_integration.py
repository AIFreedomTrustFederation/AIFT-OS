#!/usr/bin/env python3
"""Run real local producer QA and consume its unsigned evidence; never deploy."""
import argparse
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import tempfile

CONSUMER = Path(__file__).resolve().with_name("check_federation_evidence.py")


def execute(args, cwd, env):
    return subprocess.run(args, cwd=cwd, env=env, capture_output=True,
                          text=True, timeout=600, check=False)


def checkout(root, expected, env):
    if not re.fullmatch(r"[0-9a-f]{40}", expected):
        raise ValueError("expected commit must be a full lowercase SHA")
    top = execute(["git", "rev-parse", "--show-toplevel"], root, env)
    head = execute(["git", "rev-parse", "HEAD"], root, env)
    status = execute(["git", "status", "--porcelain", "--untracked-files=all"], root, env)
    if top.returncode or Path(top.stdout.strip()).resolve() != root:
        raise ValueError("input must be an exact Git repository root")
    if head.returncode or head.stdout.strip() != expected:
        raise ValueError("input checkout does not match expected commit")
    if status.returncode or status.stdout:
        raise ValueError("input checkout must be clean")


def run(runtime, forge, runtime_sha, forge_sha):
    # These are local test executions, never fabricated Actions attestations.
    env = {k: v for k, v in os.environ.items() if not k.startswith("GITHUB_")}
    env["PYTHONDONTWRITEBYTECODE"] = "1"
    forbidden = {"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_COMMON_DIR",
                 "GIT_OBJECT_DIRECTORY", "GIT_ALTERNATE_OBJECT_DIRECTORIES",
                 "GIT_CONFIG_COUNT", "GIT_CONFIG_PARAMETERS", "GIT_CONFIG_GLOBAL",
                 "GIT_CONFIG_SYSTEM", "GIT_REPLACE_REF_BASE", "GIT_NAMESPACE", "GIT_GRAFT_FILE"}
    if any(k in forbidden or k.startswith(("GIT_CONFIG_KEY_", "GIT_CONFIG_VALUE_")) for k in env):
        raise ValueError("Git source overrides are unsupported")
    roots = [("AIFT-Runtime", runtime.resolve(), runtime_sha, "run_runtime_verification.py"),
             ("AIFT-Forge", forge.resolve(), forge_sha, "run_forge_verification.py")]
    if roots[0][1] == roots[1][1]:
        raise ValueError("producer roots must be distinct")
    # Validate both inputs before either producer can execute.
    for _, root, sha, script in roots:
        checkout(root, sha, env)
        source = root / "scripts" / script
        if not source.is_file() or source.is_symlink() or not source.resolve().is_relative_to(root):
            raise ValueError("producer entrypoint missing or outside checkout")
    with tempfile.TemporaryDirectory(prefix="aift-producer-integration-") as tmp:
        command = [sys.executable, "-B", str(CONSUMER)]
        codes = {}
        for name, root, sha, script in roots:
            output = Path(tmp) / (name + ".json")
            result = execute([sys.executable, "-B", str(root / "scripts" / script),
                              "--output", str(output)], root, env)
            print(result.stdout, file=sys.stderr, end="")
            print(result.stderr, file=sys.stderr, end="")
            codes[name] = result.returncode
            if result.returncode not in (0, 1) or not output.is_file():
                raise ValueError(name + " did not publish consumable evidence")
            command += ["--expect", "AIFreedomTrustFederation/" + name + "=" + sha,
                        "--input", str(output)]
        result = execute(command, CONSUMER.parent.parent, env)
        if result.returncode not in (0, 1):
            raise ValueError("consumer execution failed: " + result.stderr)
        report = json.loads(result.stdout)
        # Independently recheck identities and checkout state after both runs.
        for _, root, sha, _ in roots:
            checkout(root, sha, env)
        report["integration_scope"] = "real_local_producers_and_offline_consumer"
        report["producer_exit_codes"] = codes
        report["integration_limitations"] = [
            "Local unsigned evidence; not authenticated GitHub artifacts.",
            "No model inference, live services, security certification or deployment.",
            "Ignored files and external host state are outside clean-tree checks."]
        print(json.dumps(report, indent=2, sort_keys=True))
        return 1 if result.returncode or any(codes.values()) else 0


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--runtime-root", type=Path, required=True)
    parser.add_argument("--runtime-sha", required=True)
    parser.add_argument("--forge-root", type=Path, required=True)
    parser.add_argument("--forge-sha", required=True)
    args = parser.parse_args()
    try:
        return run(args.runtime_root, args.forge_root, args.runtime_sha, args.forge_sha)
    except (OSError, ValueError, subprocess.SubprocessError) as error:
        print("Producer integration blocked: " + str(error), file=sys.stderr)
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
