# Evidence credential bootstrap

Run this locally from a reviewed checkout with Python 3 and an administrator-authenticated GitHub CLI session for github.com. Do not supply keys through chat, PRs, issues, workflow inputs, or command-line arguments.

## One-time administrator setup

Create or select a GitHub App, install it on AIFreedomTrustFederation/.github and AIFreedomTrustFederation/AIFT-Runtime, and grant only Contents read and Actions read (plus GitHub's implicit metadata access). Record its Client ID and generate a private-key PEM file. App creation, installation approval, and private-key issuance remain administrator actions; this script does not automate them or use the ChatGPT connector's App identity.

## Preview

```sh
python3 scripts/setup_federation_evidence.py
```

This checks administrator access and prints the fixed target settings. Preview does not read a key or mutate GitHub.

## Apply

Protect the PEM file with owner-only permissions, then run:

```sh
chmod 600 /secure/path/evidence-app.pem
python3 scripts/setup_federation_evidence.py --apply \
  --client-id YOUR_APP_CLIENT_ID \
  --private-key /secure/path/evidence-app.pem
```

This writes AIFT_FEDERATION_EVIDENCE_APP_PRIVATE_KEY through gh stdin, then sets AIFT_FEDERATION_EVIDENCE_APP_CLIENT_ID. The private key is never included in child arguments or printed. Secret values cannot be read back. The PEM header check is not cryptographic validation; successful token creation and collection remain necessary.

Writes are sequential, not transactional. If the variable write or later verification fails, the earlier secret write remains. Reapplying replaces the same named settings and supports key rotation. Retain the previous valid key until the replacement has been verified, then revoke the old key through GitHub App settings.

## Verify

Optionally add --rerun FAILED_MAIN_RUN_ID to the apply command. The script verifies that the ID belongs to a completed failed collector push run on the current AIFT-OS main revision before writes and again before rerunning failed jobs. A PR, stale SHA, other workflow, or other repository is rejected. A branch can still advance after validation; inspect the run revision.

```sh
gh run watch FAILED_MAIN_RUN_ID --repo AIFreedomTrustFederation/AIFT-OS --exit-status
```

Token minting, producer API access, live collection, and artifact upload must all succeed. Inspect federation-assessment.json for exact repository/commit/run/attempt/artifact/digest binding and explicit unsigned/partial semantics. A requested rerun is not a successful collection. The existing hourly trusted-default-branch workflow performs recurring collection; no new scheduler is needed.

The bootstrap does not modify workflows, producer allowlists, App permissions, deployments, or branch protections. VPS ingestion remains a separate reviewed integration. Offline tests use synthetic credentials and never configure live settings.
