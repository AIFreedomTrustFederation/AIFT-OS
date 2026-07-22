#!/data/data/com.termux/files/usr/bin/bash

echo "========================================"
echo " AIFT-OS FUTURE ARCHITECTURE AUDIT"
echo "========================================"
echo

ROOT="$(pwd)"

check() {
    if [ -e "$1" ]; then
        echo "✓ FOUND   $1"
    else
        echo "✗ MISSING $1"
    fi
}

echo "=== CORE CONTROL PLANE ==="
check "cmd/aift/main.go"
check "internal/cli"
check "registry/cli"
echo

echo "=== CAPABILITY SYSTEM ==="
check ".aift/capabilities.json"
check "internal/capabilities"
check "internal/capability"
check "internal/discoveryengine"
echo

echo "=== RUNTIME ENGINE ==="
check "internal/runtime"
check "internal/execution"
check "internal/supervisor"
check "internal/scheduler"
check "internal/planner"
check "internal/workflow"
echo

echo "=== AI / COGNITIVE SYSTEM ==="
check "internal/intelligence"
check "internal/graph"
check "internal/ai"
echo

echo "=== MEMORY / KNOWLEDGE ==="
check "internal/state"
check "internal/events"
check "internal/eventmesh"
check "internal/eventbus"
echo

echo "=== PLUGIN ARCHITECTURE ==="
check "internal/plugins"
check "internal/providers"
check "internal/providerregistry"
echo

echo "=== FEDERATION ==="
check "internal/federation"
check "FEDERATION.md"
check "aift.repo.json"
echo

echo "=== SELF MANAGEMENT ==="
check "tools/cli-self-repair"
check "tools/cli-self-verifier"
check "internal/doctor"
check "internal/readiness"
echo

echo "=== SECURITY READINESS ==="
check "internal/security"
check "internal/auth"
check "internal/identity"
echo

echo "=== HARDWARE ABSTRACTION ==="
check "internal/hal"
check "internal/device"
check "internal/drivers"
echo

echo "=== QUANTUM READINESS ==="
check "internal/quantum"
check "internal/quantum/providers"
echo

echo "=== SUPPLY CHAIN SECURITY ==="
check "SBOM"
check "security"
check "DEPENDENCIES.md"
check "go.sum"
check "package.json"
echo

echo "=== IMMUTABLE UPDATE MODEL ==="
check "internal/version"
check "internal/lifecycle"
check "internal/rollback"
echo

echo "=== PLATFORM DETECTION ==="

if grep -R "runtime.GOOS" -n . >/dev/null 2>&1; then
    echo "✓ Go OS detection exists"
else
    echo "✗ No Go OS detection found"
fi

if grep -R "capabilities" -n .aift internal >/dev/null 2>&1; then
    echo "✓ Capability references detected"
else
    echo "✗ Capability references missing"
fi

echo
echo "=== DEPENDENCY SUMMARY ==="

if command -v go >/dev/null; then
    echo "Go:"
    go version
fi

if command -v node >/dev/null; then
    echo "Node:"
    node --version
fi

if command -v pnpm >/dev/null; then
    echo "PNPM:"
    pnpm --version
fi

echo
echo "========================================"
echo " AUDIT COMPLETE"
echo "========================================"
