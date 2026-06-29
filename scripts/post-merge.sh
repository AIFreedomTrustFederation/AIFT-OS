#!/bin/bash
# no-harness: Replit post-merge workspace hook — runs pnpm install after task-agent merges, not an AIFT OS runtime script
set -e
pnpm install --frozen-lockfile
pnpm --filter db push
