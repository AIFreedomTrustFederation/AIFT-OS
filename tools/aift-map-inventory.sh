#!/data/data/com.termux/files/usr/bin/bash

echo "========================================"
echo " AIFT-OS UNIFIED MAP INVENTORY"
echo "========================================"

OUT="reports/aift-unified-map-$(date +%Y%m%d-%H%M%S).md"

mkdir -p reports

{
echo "# AIFT-OS Unified Architecture Map"
echo
echo "Generated: $(date)"
echo

echo "## Core Graph"
find internal/graph -type f 2>/dev/null

echo
echo "## Federation Nodes"
find internal/federation -type f 2>/dev/null

echo
echo "## Capability Registry"
find . -path "*capabil*" -type f \
  | grep -v node_modules

echo
echo "## Runtime Components"
find internal -maxdepth 2 -type d \
 | grep -E "runtime|execution|scheduler|planner|workflow|supervisor"

echo
echo "## Intelligence Layer"
find internal -maxdepth 2 -type d \
 | grep -E "ai|intelligence|memory|state"

echo
echo "## Security Status"
for x in internal/security internal/auth internal/identity
do
 if [ -e "$x" ]; then
  echo "FOUND $x"
 else
  echo "MISSING $x"
 fi
done

echo
echo "## Hardware Abstraction"
for x in internal/hal internal/device internal/drivers
do
 if [ -e "$x" ]; then
  echo "FOUND $x"
 else
  echo "MISSING $x"
 fi
done

echo
echo "## Quantum Readiness"
for x in internal/quantum internal/quantum/providers
do
 if [ -e "$x" ]; then
  echo "FOUND $x"
 else
  echo "MISSING $x"
 fi
done

echo
echo "## Registries"
find registry -maxdepth 2 -type f \
 | grep -v reports

echo
echo "## Documentation Architecture"
find docs -type f \
 | grep -Ei "architecture|design|model|baseline"

echo
echo "## Future Required Maps"
echo "- Identity Graph"
echo "- Trust Graph"
echo "- Security Threat Graph"
echo "- Hardware Capability Graph"
echo "- Quantum Provider Graph"
echo "- Dependency Supply Chain Graph"
echo "- AI Agent Knowledge Graph"
echo "- Autonomous Governance Graph"

} > "$OUT"

echo
echo "========================================"
echo " MAP COMPLETE"
echo "$OUT"
echo "========================================"
