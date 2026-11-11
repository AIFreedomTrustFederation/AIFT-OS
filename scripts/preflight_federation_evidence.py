#!/usr/bin/env python3
"""Read-only AIFT evidence reader bootstrap diagnostics (never reads secret values)."""
import argparse
import json
import os
import subprocess
import sys

REPO = "AIFreedomTrustFederation/AIFT-OS"
PRODUCERS = ("AIFreedomTrustFederation/.github", "AIFreedomTrustFederation/AIFT-Runtime")
VARIABLE = "AIFT_FEDERATION_EVIDENCE_APP_CLIENT_ID"
SECRET = "AIFT_FEDERATION_EVIDENCE_APP_PRIVATE_KEY"

def request(path):
    env = dict(os.environ, GH_HOST="github.com", GH_PROMPT_DISABLED="1")
    try:
        result = subprocess.run(["gh", "api", "--hostname", "github.com", path],
                                capture_output=True, timeout=30, env=env, check=False)
    except (OSError, subprocess.TimeoutExpired):
        return None
    if result.returncode:
        return None
    try:
        return json.loads(result.stdout)
    except (ValueError, UnicodeError):
        return None

def diagnose(fetch=request):
    # Metadata endpoints reveal names only, not secret values.
    variable = fetch(f"repos/{REPO}/actions/variables/{VARIABLE}")
    secret = fetch(f"repos/{REPO}/actions/secrets/{SECRET}")
    checks = {
        "client_id_variable": isinstance(variable, dict) and variable.get("name") == VARIABLE and bool(variable.get("value")),
        "private_key_secret_exists": isinstance(secret, dict) and secret.get("name") == SECRET,
    }
    # Access through the operator's CLI is only a visibility check, not proof
    # that the GitHub App installation token can read these producers.
    for producer in PRODUCERS:
        result = fetch(f"repos/{producer}")
        checks["operator_can_see_" + producer.split("/")[-1]] = (
            isinstance(result, dict) and result.get("full_name") == producer
        )
    return checks

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--json", action="store_true", help="Machine-readable results")
    args = parser.parse_args()
    checks = diagnose()
    if args.json:
        print(json.dumps({"checks": checks, "ready_for_live_attempt": all(checks.values()),
                          "note": "Metadata checks do not verify App installation or token permissions."}, sort_keys=True))
    else:
        for name, ok in checks.items():
            print(f"{'OK' if ok else 'BLOCKED'}: {name}")
        print("This preflight never reads private key values or mints an App token.")
        print("App installation permissions and current-SHA evidence must be verified by the live workflow.")
        if not all(checks.values()):
            print("Register/install the read-only App, then run scripts/setup_federation_evidence.py --help.")
    return 0 if all(checks.values()) else 1

if __name__ == "__main__":
    sys.exit(main())
