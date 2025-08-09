

#!/usr/bin/env python3
"""
Detectviz module.card.json validator

功能：
- 以 contracts/schemas/module.card.schema.json 驗證結構（JSON Schema Draft 2020-12）。
- 進一步檢查平台規範：
  * agent：觀測基線（agent.run）、工具/記憶體對應 span、A2A timeouts 與容量、capability_tags。
  * plugin：依 category 檢查 I/O 與觀測命名（collector/process/aggregate/write）。
  * aggregate.aggregator：責任說明需包含窗口語義（TUMBLING/SLIDING/WINDOW）。
  * specVersion >= 1.1.0。
- 支援檔案或目錄掃描（遞迴尋找 **/module.card.json）。
- 產出人類可讀摘要與選擇性的 JSON 報告。

退出碼：0=全通過（或僅警告且未要求 fail-on-warn）；1=有錯誤；2=僅警告且啟用 --fail-on-warn。
"""
from __future__ import annotations

import argparse
import json
import os
import re
import sys
from pathlib import Path
from typing import Any, Dict, List, Tuple

# ---- jsonschema 依賴檢查 ----
try:
    import jsonschema
    from jsonschema import Draft202012Validator
except Exception as e:  # pragma: no cover
    print("[ERROR] 需要 jsonschema 套件：pip install jsonschema>=4", file=sys.stderr)
    raise

SCHEMA_MIN_VER = (1, 1, 0)


def semver_tuple(s: str) -> Tuple[int, int, int]:
    m = re.match(r"^(\d+)\.(\d+)\.(\d+)", s)
    if not m:
        return (0, 0, 0)
    return tuple(int(x) for x in m.groups())  # type: ignore


def load_schema(schema_path: Path) -> Dict[str, Any]:
    with schema_path.open("r", encoding="utf-8") as f:
        return json.load(f)


def iter_cards(root: Path) -> List[Path]:
    if root.is_file():
        return [root]
    out: List[Path] = []
    for p in root.rglob("module.card.json"):
        if p.is_file():
            out.append(p)
    return out


class Issue:
    def __init__(self, level: str, msg: str, path: List[str] | None = None):
        self.level = level  # "ERROR" | "WARN"
        self.msg = msg
        self.path = path or []

    def to_dict(self) -> Dict[str, Any]:
        return {"level": self.level, "msg": self.msg, "path": self.path}


def validate_against_schema(card: Dict[str, Any], schema: Dict[str, Any]) -> List[Issue]:
    issues: List[Issue] = []
    validator = Draft202012Validator(schema)
    for err in sorted(validator.iter_errors(card), key=lambda e: e.path):
        issues.append(Issue("ERROR", f"schema: {err.message}", [str(x) for x in err.path]))
    return issues


def ensure_observability(card: Dict[str, Any], issues: List[Issue]) -> None:
    spans = set(card.get("observability", {}).get("spans", []) or [])
    metrics = set(card.get("observability", {}).get("metrics", []) or [])
    kind = card.get("kind")
    category = card.get("category")

    if kind == "agent":
        if "agent.run" not in spans:
            issues.append(Issue("ERROR", "agent 必須包含 span: agent.run", ["observability", "spans"]))
        # 若使用 tools，需含 tool.exec
        if card.get("tools"):
            if "tool.exec" not in spans:
                issues.append(Issue("WARN", "agent 使用 tools 時建議包含 span: tool.exec", ["observability", "spans"]))
        # 若宣告 memory read/write，需含 memory.*
        mem = card.get("memory", {}) or {}
        if mem.get("read") or mem.get("write"):
            if not ({"memory.search", "memory.read", "memory.write"} & spans):
                issues.append(Issue("WARN", "agent 使用記憶體時建議包含 span: memory.*", ["observability", "spans"]))
        if "agent_runs_total" not in metrics:
            issues.append(Issue("WARN", "建議提供 metrics: agent_runs_total", ["observability", "metrics"]))

    if kind == "plugin":
        expected = {
            "collector.input": "plugin.collect",
            "transform.processor": "plugin.process",
            "aggregate.aggregator": "plugin.aggregate",
            "sink.output": "plugin.write",
        }.get(category)
        if expected and expected not in spans:
            issues.append(Issue("ERROR", f"plugin 類別 {category} 必須包含 span: {expected}", ["observability", "spans"]))
        # 基線 metrics
        if "plugin_runs_total" not in metrics:
            issues.append(Issue("WARN", "建議提供 metrics: plugin_runs_total", ["observability", "metrics"]))
        if "plugin_duration_seconds_bucket" not in metrics:
            issues.append(Issue("WARN", "建議提供 metrics: plugin_duration_seconds_bucket", ["observability", "metrics"]))


def ensure_business_rules(card: Dict[str, Any], issues: List[Issue]) -> None:
    kind = card.get("kind")
    spec_version = card.get("specVersion", "0.0.0")
    if semver_tuple(spec_version) < SCHEMA_MIN_VER:
        issues.append(Issue("ERROR", f"specVersion 必須 >= 1.1.0，當前: {spec_version}", ["specVersion"]))

    if kind == "agent":
        a2a = card.get("a2a", {}) or {}
        # capability_tags 建議存在
        if not a2a.get("capability_tags"):
            issues.append(Issue("WARN", "建議設定 a2a.capability_tags 以利路由與檢索", ["a2a", "capability_tags"]))
        # timeouts 關係：total >= task >= tool
        t = (a2a.get("timeouts_ms") or {})
        tool = t.get("tool")
        task = t.get("task")
        total = t.get("total")
        if any(x is None for x in (tool, task, total)):
            issues.append(Issue("ERROR", "a2a.timeouts_ms 需同時包含 tool/task/total", ["a2a", "timeouts_ms"]))
        else:
            if not (total >= task >= tool):
                issues.append(Issue("ERROR", f"a2a.timeouts_ms 順序需滿足 total ≥ task ≥ tool（目前: total={total}, task={task}, tool={tool})", ["a2a", "timeouts_ms"]))
        # 並發與佇列
        if (a2a.get("concurrency_limit") or 0) < 1:
            issues.append(Issue("ERROR", "a2a.concurrency_limit 必須 ≥ 1", ["a2a", "concurrency_limit"]))
        if (a2a.get("queue_depth") or 0) < 0:
            issues.append(Issue("ERROR", "a2a.queue_depth 不可為負數", ["a2a", "queue_depth"]))

    if kind == "plugin":
        category = card.get("category")
        inputs = card.get("inputs") or []
        outputs = card.get("outputs") or []
        if category == "transform.processor":
            if not inputs or not outputs:
                issues.append(Issue("ERROR", "transform.processor 必須同時定義 inputs 與 outputs", []))
        if category == "aggregate.aggregator":
            resp = (card.get("responsibility") or "").upper()
            if not any(k in resp for k in ["TUMBLING", "SLIDING", "WINDOW"]):
                issues.append(Issue("WARN", "aggregate.aggregator 建議在 responsibility 說明窗口語義（TUMBLING/SLIDING/WINDOW）", ["responsibility"]))


def validate_card(card_path: Path, schema: Dict[str, Any]) -> Dict[str, Any]:
    with card_path.open("r", encoding="utf-8") as f:
        card = json.load(f)

    issues: List[Issue] = []
    issues += validate_against_schema(card, schema)
    ensure_observability(card, issues)
    ensure_business_rules(card, issues)

    has_error = any(i.level == "ERROR" for i in issues)
    has_warn = any(i.level == "WARN" for i in issues)

    return {
        "file": str(card_path),
        "ok": not has_error,
        "has_warnings": has_warn,
        "issues": [i.to_dict() for i in issues],
        "kind": card.get("kind"),
        "id": card.get("id"),
        "role": card.get("role"),
        "category": card.get("category"),
        "specVersion": card.get("specVersion"),
    }


def main() -> int:
    parser = argparse.ArgumentParser(description="Validate Detectviz module.card.json files")
    parser.add_argument("path", nargs="?", default=".", help="module.card.json 檔案或根目錄（遞迴掃描）")
    parser.add_argument("--schema", default=None, help="覆寫 Schema 路徑（預設為 contracts/schemas/module.card.schema.json）")
    parser.add_argument("--json-out", default=None, help="輸出 JSON 報告檔路徑")
    parser.add_argument("--fail-on-warn", action="store_true", help="有警告時亦回傳非零代碼（2）")
    args = parser.parse_args()

    root = Path(args.path).resolve()

    # 自動推導 schema 預設位置
    if args.schema:
        schema_path = Path(args.schema).resolve()
    else:
        # 本檔位於 contracts/tools/validate_module_card.py
        schema_path = (Path(__file__).resolve().parents[1] / "schemas" / "module.card.schema.json")

    if not schema_path.exists():
        print(f"[ERROR] 找不到 Schema: {schema_path}", file=sys.stderr)
        return 1

    schema = load_schema(schema_path)

    cards = iter_cards(root)
    if not cards:
        print(f"[WARN] 未找到任何 module.card.json（根目錄：{root}）")
        return 0

    results: List[Dict[str, Any]] = []
    total_errors = 0
    total_warnings = 0

    for card_path in sorted(cards):
        res = validate_card(card_path, schema)
        results.append(res)
        errs = [i for i in res["issues"] if i["level"] == "ERROR"]
        warns = [i for i in res["issues"] if i["level"] == "WARN"]
        total_errors += len(errs)
        total_warnings += len(warns)

        status = "OK" if res["ok"] else "ERROR"
        warn_tag = " (warn)" if res["has_warnings"] else ""
        print(f"[{status}{warn_tag}] {card_path}")
        for i in errs + warns:
            loc = "/".join(i.get("path") or [])
            print(f"  - {i['level']}: {i['msg']}{' @'+loc if loc else ''}")

    if args.json_out:
        out_path = Path(args.json_out).resolve()
        out_path.parent.mkdir(parents=True, exist_ok=True)
        with out_path.open("w", encoding="utf-8") as f:
            json.dump({
                "schema": str(schema_path),
                "root": str(root),
                "summary": {"errors": total_errors, "warnings": total_warnings, "files": len(results)},
                "results": results,
            }, f, ensure_ascii=False, indent=2)
        print(f"[INFO] JSON 報告已輸出：{out_path}")

    if total_errors > 0:
        return 1
    if total_warnings > 0 and args.fail_on_warn:
        return 2
    return 0


if __name__ == "__main__":
    sys.exit(main())