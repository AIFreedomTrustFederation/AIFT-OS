# Federation required-check rulesets

This proposal creates **repository-level** rulesets for the 17 listed public repositories. They are grouped by expected check name in the controller, but GitHub receives one ruleset per repository. This avoids depending on organization-level ruleset availability.

## Safety model

- Workflow dispatch only. Audit is the default; applying requires the explicit `apply` input and the protected `federation-policy-apply` GitHub Actions environment with required reviewers.
- Create a fine-grained credential with repository **Administration: read/write** and **Contents: read** for only the 17 target repositories. Store it as the Actions secret `AIFT_RULESET_ADMIN_TOKEN` in the relevant environments. Never commit credentials.
- The token must also be able to read PRs and check runs/statuses. Grant **Pull requests: read** and **Checks: read** as needed.
- Before applying, the controller confirms that the exact required check has appeared on a non-draft PR head targeting the default branch, and that auto-merge and squash merging are enabled.
- No automatic changes to existing rulesets. Existing rules with the managed name must match the declared policy; otherwise execution stops for manual review.
- Existing protections are not removed or relaxed. No bypass actors. No auto-merge is enabled on PRs by this script.
- PR review threads must be resolved. This does not substitute for security/governance-specific reviewer approval. Configure separate code-owner protections where appropriate.
- The controller intentionally fails closed if there is no qualifying PR-head evidence, including repositories with no PR history. Validate those manually before extending the controller.
- Exact check names are hypotheses from prior observation, not verified facts. The preflight must pass before application.
- Private repositories and chktex are deliberately excluded. Paid-plan ruleset eligibility should be checked with GitHub before expanding scope.

## Setup

1. Merge this PR after reviewing the script and workflow.
2. Create the two GitHub Actions environments `federation-policy-audit` and `federation-policy-apply`. Require trusted reviewers for the apply environment.
3. Provision the restricted credential in both environments as `AIFT_RULESET_ADMIN_TOKEN`.
4. Enable **Allow auto-merge** and **Allow squash merging** in AetherianGovernance and capital-city-provisions.
5. Dispatch the workflow with `apply=false`, inspect the audit, and correct any mismatched check names in a reviewed PR.
6. Dispatch with `apply=true` only after the preflight passes and the protected environment reviewer approves.

GitHub ruleset creation is not executed by merging this PR alone. The credential, environments and approved dispatch are required.
