#!/usr/bin/env bash
# scripts/gen-proto.sh — generates Go protobuf + Connect handler stubs
# Requires: buf (https://buf.build/docs/installation)
set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT"

echo "→ Checking buf..."
which buf >/dev/null 2>&1 || { echo "Install buf: https://buf.build/docs/installation"; exit 1; }

echo "→ Generating proto stubs..."
buf generate

echo "✓ Done. Generated files in gen/"
echo ""
echo "Packages generated:"
find gen -name "*.go" | sed 's|gen/||' | sort
