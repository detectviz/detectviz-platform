#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

req=(
  "buf.yaml"
  "buf.gen.yaml"
  "proto/detectviz/contracts/v1/adk_bridge.proto"
  "schemas/module.card.schema.json"
  "schemas/config.schema.json"
  "samples/config.yaml"
  "tools/validate_module_card.py"
)

echo "== Required files =="
missing=0
for f in "${req[@]}"; do
  if [[ -f "$ROOT/$f" ]]; then
    echo "[OK] $f"
  else
    echo "[MISS] $f"; missing=1
  fi
done

echo
echo "== buf lint & generate =="
cd "$ROOT"
buf lint
buf generate

echo
if [[ $missing -eq 0 ]]; then
  echo "Summary: contracts 目錄必要檔案齊全。"
else
  echo "Summary: 有必要檔案缺漏，請依上方 [MISS] 修補。"; exit 1
fi
