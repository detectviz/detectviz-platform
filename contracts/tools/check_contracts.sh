#!/usr/bin/env bash
set -euo pipefail

# contracts/tools/check_contracts.sh
# 一鍵執行 contracts 區的 lint / generate / validate 流程

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

have_cmd() { command -v "$1" >/dev/null 2>&1; }

# ---- 0) 先檢查必備檔案 -----------------------------------------------------
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
    echo "[OK]   $f"
  else
    echo "[MISS] $f"; missing=1
  fi
done
if [[ $missing -ne 0 ]]; then
  echo "[FAIL] 缺少必要檔案，請依上方 [MISS] 修補後再執行。" >&2
  exit 1
fi

# ---- 1) 工具與環境檢查 -----------------------------------------------------
if ! have_cmd buf; then
  echo "[FAIL] 未找到 'buf' 指令，請先安裝：https://buf.build/" >&2
  exit 1
fi
if ! have_cmd python3; then
  echo "[WARN] 未找到 'python3'，將跳過部分驗證（config/schema 與 module.card）。" >&2
fi

# ---- 2) buf lint & generate ------------------------------------------------
echo
echo "== buf lint & generate =="
buf lint
buf generate

# 生成碼存在性檢查（路徑須與 buf.gen.yaml 對齊）
GOGO="gen/go/detectviz/contracts/v1"
PYOUT="gen/python/detectviz/contracts/v1"
[[ -d "$GOGO" ]] || { echo "[FAIL] Go 生成碼不存在：$GOGO" >&2; exit 1; }
[[ -d "$PYOUT" ]] || { echo "[FAIL] Python 生成碼不存在：$PYOUT" >&2; exit 1; }

echo "[OK] 生成碼產出：$GOGO, $PYOUT"

# ---- 3) Schema 驗證：samples/config.yaml -----------------------------------
echo
echo "== validate: schemas/config.schema.json vs samples/config.yaml =="
if have_cmd python3; then
  python3 - <<'PY'
import sys, json, pathlib
try:
    import yaml, jsonschema  # type: ignore
except Exception as e:
    print(f"[WARN] 缺少 python 相依 (yaml/jsonschema)，跳過 config.yaml 驗證：{e}")
    sys.exit(0)
root = pathlib.Path(__file__).resolve().parents[1]
schema_path = root / 'schemas' / 'config.schema.json'
sample_path = root / 'samples' / 'config.yaml'

schema = json.loads(schema_path.read_text(encoding='utf-8'))
config = yaml.safe_load(sample_path.read_text(encoding='utf-8')) or {}
jsonschema.Draft202012Validator(schema).validate(config)
print('[OK] config.yaml 通過 JSON Schema 驗證')
PY
else
  echo "[SKIP] 無 python3，略過 config.yaml 驗證"
fi

# ---- 4) Module Cards 驗證 --------------------------------------------------
echo
echo "== validate: module.card.json (recursive) =="
if have_cmd python3; then
  shopt -s nullglob
  mapfile -t cards < <(find "$ROOT" -type f -name 'module.card.json' | sort)
  if ((${#cards[@]}==0)); then
    echo "[INFO] 未找到 module.card.json，可略過"
  else
    fail=0
    for m in "${cards[@]}"; do
      echo "- $m"
      if ! python3 tools/validate_module_card.py "$m"; then
        echo "[FAIL] 驗證失敗：$m"; fail=1
      fi
    done
    if [[ $fail -ne 0 ]]; then
      echo "[FAIL] 有模組卡驗證失敗" >&2
      exit 1
    fi
    echo "[OK] 所有 module.card.json 均通過"
  fi
else
  echo "[SKIP] 無 python3，略過 module.card.json 驗證"
fi

# ---- 5) 總結 ----------------------------------------------------------------
echo
echo "== Summary =="
echo "contracts 目錄 lint/generate/validate 全部完成"
exit 0
