# -*- coding: utf-8 -*-
"""将 data/sa_questions_standard.json 生成批量 INSERT SQL（与 import_real 逻辑一致）。

字段格式对齐 backend/cmd/import_real/main.go：
- subject_id=9（系统分析师），sub_subject_id=4（综合知识）
- options: ["A. 文本", "B. 文本", ...]（按键排序）
- type: single/multi -> single
- difficulty: medium, status: 1
- source: "{source} | {paper}"
"""
import io
import json
import os

BASE = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..")
SRC = os.path.join(BASE, "data", "sa_questions_standard.json")
DST = os.path.join(BASE, "data", "sa_questions_real.sql")

SUBJECT_ID = 9
SUB_SUBJECT_ID = 4  # 系统分析师-综合知识


def esc(s):
    return "'" + str(s).replace("\\", "\\\\").replace("'", "''") + "'"


def main():
    recs = json.load(io.open(SRC, encoding="utf-8"))
    lines = [
        "SET NAMES utf8mb4;",
        "INSERT INTO `questions` "
        "(`subject_id`, `sub_subject_id`, `type`, `difficulty`, `content`, `options`, `answer`, `analysis`, `year`, `status`, `source`, `created_at`) VALUES",
    ]
    parts = []
    for r in recs:
        content = r.get("content", "").strip()
        answer = r.get("answer", "").strip()
        if not content or not answer:
            continue
        qtype = r.get("type", "single")
        if qtype not in ("single", "multi"):
            qtype = "single"
        opts = r.get("options") or {}
        keys = sorted(opts.keys())
        arr = [f"{k}. {opts[k]}" for k in keys]
        options_json = json.dumps(arr, ensure_ascii=False)
        paper = r.get("paper", "")
        source = r.get("source", "")
        src = f"{source} | {paper}" if source else paper
        parts.append(
            "({}, {}, '{}', 'medium', {}, {}, {}, {}, {}, 1, {}, NOW())".format(
                SUBJECT_ID, SUB_SUBJECT_ID, qtype,
                esc(content), esc(options_json), esc(answer),
                esc(r.get("analysis", "")), int(r.get("year", 0)), esc(src),
            )
        )
    lines.append(",\n".join(parts) + ";")
    io.open(DST, "w", encoding="utf-8", newline="\n").write("\n".join(lines))
    print(f"写入 {DST}: {len(parts)} 行")


if __name__ == "__main__":
    main()
