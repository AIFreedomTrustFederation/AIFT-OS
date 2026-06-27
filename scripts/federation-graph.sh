#!/usr/bin/env bash
set -euo pipefail

WORKSPACE="${1:-$HOME/AIFT}"
OUT="$WORKSPACE/reports/federation-graph.txt"

mkdir -p "$WORKSPACE/reports"

cat > "$OUT" <<GRAPH
AI Freedom Trust Federation
├── AIFT-Forge
│   ├── Templates
│   ├── Manifests
│   ├── Agents
│   ├── Scripts
│   └── Local-first operating system core
├── www.aifreedomtrust.com
│   └── Public portal
├── VPS
│   └── Infrastructure and node systems
├── booksmith-ai
│   └── Author-first writing and publishing studio
├── AI-Freedom-Trust
│   └── Doctrine, trust research, and alignment papers
├── Aether_Coin_biozonecurrency
│   └── Economic layer and DynastyLink concepts
├── capital-city-provisions
│   └── Applied federation business system
├── TheMindofAll
│   └── Consciousness and theory layer
└── biozone-harmony-boost
    └── Biozone / wellness / harmony systems
GRAPH

echo "Graph written to: $OUT"
cat "$OUT"
