#!/bin/bash
# no-harness: Replit post-merge workspace hook — restores dependencies after task-agent merges, not an AIFT OS runtime script
set -e
pnpm install --frozen-lockfile --ignore-scripts

# Database schema changes require a separate, explicit operator action. A merge
# hook must not mutate shared or local database state without human approval.
