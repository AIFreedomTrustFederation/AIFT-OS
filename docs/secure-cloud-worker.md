# AIFT Secure Cloud AI Documentation Worker

This is a scheduled cloud job, not a permanently running daemon. Once merged into AIFT-OS main and explicitly enabled, GitHub Actions attempts to check for approved documentation tasks every 30 minutes whether or not the operator's phone is online. GitHub may delay or skip scheduled starts.

The worker is intentionally a small, verifiable first AI capability. It **cannot** autonomously execute code, merge pull requests, deploy, change trust/identity/security/governance rules, alter workflows, or create or rotate credentials. Broader development automation requires later, separately reviewed permissions and workflows.

## Setup checklist (operator-only configuration)

1. Review this PR and PR #73; both must have passing exact-head CI and approval. Merge PR #73 first, then reconcile and merge this stacked PR. A scheduled workflow is not active while it is only in a PR branch.
2. Register a dedicated GitHub App and install it **only on AIFT-OS**, granting Repository **Contents: Read and write** and **Pull requests: Read and write**, Metadata read (implicit). No Workflow/Administration/Deployments/Secrets permissions. The GitHub App is an identity that GitHub Actions can mint short-lived tokens from; it does not itself run a server.
3. AIFT-OS Settings -> Secrets and variables -> Actions:
   - Variable **AIFT_AUTODEV_APP_CLIENT_ID** = dedicated GitHub App Client ID (shared with controller PR #73).
   - Secret **AIFT_AUTODEV_APP_PRIVATE_KEY** = PEM downloaded for that App. Store the entire PEM, never paste it in chat, commit it, or upload it as a public Actions artifact.
   - Secret **AIFT_AUTODEV_OPENAI_API_KEY** = a dedicated, scoped OpenAI Platform project API key with spend limits set in the Platform project. ChatGPT subscription is not an OpenAI API billing credential.
   - Optional variable **AIFT_AUTODEV_MODEL** = gpt-6-luna (default); select an API model your project can call.
   - Variable **AIFT_AUTODEV_AI_ENABLED** = false initially. After a manual smoke test and policy review, change to **true** to enable model calls.
4. In AIFT-OS Issues -> Labels, add label **aift-autodev-docs**. Apply it only to issues whose instructions and public documentation target you approve. A labeller with write permission is authorising an AI draft proposal, not a merge.
5. Write a labelled open GitHub issue like:

        Target: docs/validation.md
        Instruction: Clarify the existing validation example without inventing features or test results.

   The target must be an existing tracked Markdown file under docs/, outside protected security, trust, identity, doctrine, governance, financial, release and credential paths. Issue body max 3500 characters and instruction max 1800. Avoid personal or secret data.
6. After the workflow is merged, run **AIFT Secure Cloud AI Documentation Worker** manually on main with the **run_ai_once** input set to true. This permits one controlled test while AIFT_AUTODEV_AI_ENABLED remains false. Inspect the logs and resulting draft PR. Without an eligible labelled issue there is deliberately no OpenAI API call or charge. After checking security, spend and the draft, set AIFT_AUTODEV_AI_ENABLED=true to enable scheduled runs.

## How one cloud run works

- **Offline contract:** tests run with minimal GitHub permissions; on PR events, this is the only job executed.
- **Plan:** a read-only GitHub token selects one labelled issue not already represented by an open worker PR. The model API key is available only here. It sends the approved instruction and the existing selected Markdown document to the Responses API with Structured Outputs, no model tools, max 3500 output tokens, and store disabled. If nothing is eligible, no model request occurs.
- **Validation:** the planner checks document size, changed line count, exact source content, path restrictions and output shape. It writes a single JSON plan artifact retained for one day. Never put secrets or private data in issues or documentation: Action artifacts are not a safe private-data vault, especially in public repositories.
- **Publish:** a separate GitHub Actions job has no OpenAI API secret. It mints a short-lived App token scoped only to AIFT-OS, checks the exact main SHA, original document hash, issue approval label and unchanged instruction, rejects duplicate branch names, and pushes exactly one doc file to a new branch. It creates a draft PR with evidence and the source issue reference. This step does not execute AI-generated code or shell commands.
- **Review:** exact-head hosted CI, source accuracy, privacy, dependency compatibility and branch protection remain required before any merge. The worker never deploys or merges.
- **Bounds:** one issue and one model request per successful run, 20 labelled issues scanned, 100 open PRs checked, at most 120 changed lines and 16 KB revised Markdown content, maximum 9-minute planner and 9-minute publisher jobs. GitHub concurrency queues overlapping runs by branch.

## Known limitations

- The setup and credential entry cannot be done through the current ChatGPT GitHub integration. The operator must register the App, create the OpenAI Platform key, add secrets/variables and approve the merge in GitHub.
- Running the schedule is best-effort; it does not provide 100% availability or arbitrary long-running AI coding.
- The worker's issue label is an operator-controlled task authorization. A separate allowlist for more repositories, code files, or model tools requires a new approved PR; this version is restricted to AIFT-OS documentation.
- The API project billing cap and GitHub Actions usage settings are controlled outside this repository. Set limits before enabling the job.
- The worker intentionally does not access your Android phone, Termux SSH keys, GitHub App PEM outside encrypted Actions secrets, or any chat transcript.
- The existing hourly AIFT monitoring and VPS auto-merge tasks are independent. Do not treat this workflow as a replacement or claim it has performed a task until GitHub logs confirm it.

## Validation

Run python3 -B -m unittest discover -s tests -p 'test_aift_secure_cloud_worker.py' -v.
The AIFT-OS canonical gate remains pnpm install --frozen-lockfile --ignore-scripts, pnpm run typecheck, make verify, and a clean tree check. Exact-head CI must pass before activation.

The Responses API and Structured Outputs contract are documented at https://developers.openai.com/api/reference/resources/responses/methods/create and https://developers.openai.com/api/docs/guides/structured-outputs. GitHub installation tokens and workflow permissions are described at https://docs.github.com/en/actions/concepts/security/github_token.
