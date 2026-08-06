#!/usr/bin/env bash
# no-harness: long-running local service launcher; performs no repository mutation.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

export AIFT_ROOT="${AIFT_ROOT:-$(cd "$REPO_ROOT/.." && pwd)}"
export AIFT_UXI_HOME="${AIFT_UXI_HOME:-$HOME/.aift/uxi}"
export AIFT_UXI_ADDR="${AIFT_UXI_ADDR:-127.0.0.1:8787}"
export AIFT_MODEL_URL="${AIFT_MODEL_URL:-http://127.0.0.1:8080/v1}"
export AIFT_MODEL_NAME="${AIFT_MODEL_NAME:-local}"

cd "$REPO_ROOT"
exec go run ./cmd/aiftd "$@"
