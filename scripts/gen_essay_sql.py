# -*- coding: utf-8 -*-
"""将 data/sa_essay_questions.json 生成批量 INSERT SQL（论文题，子科目=论文）。

字段格式：
- subject_id=9（系统分析师），sub_subject_id=6（论文），chapter_id=22（论文写作专题）
- options: []（论文题无选项）
- type: essay
- difficulty: hard, status: 1
- source: "{source} | {paper}"
"""
import io
import json
import os

BASE = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..")
SRC = os.path.join(BASE, "data", "sa_essay_questions.json")
DST = os.path.join(BASE, "data", "sa_essay_questions.sql")

SUBJECT_ID = 9
SUB_SUBJECT_ID = 6  # 系统分析师-论文
CHAPTER_ID = 22     # 论文写作专题


def esc(s):
    return "'" + str(s).replace("\\", "\\\\").replace("'", "''") + "'"


def main():
    recs = json.load(io.open(SRC, encoding="utf-8"))
    lines = [
        "SET NAMES utf8mb4;",
        "INSERT INTO `questions` "
        "(`subject_id`, `sub_subject_id`, `chapter_id`, `type`, `difficulty`, `content`, `options`, `answer`, `analysis`, `year`, `status`, `source`, `created_at`) VALUES",
    ]
    parts = []
    for r in recs:
        content = r.get("content", "").strip()
        if not content:
            continue
        qtype = "essay"
        paper = r.get("paper", "")
        source = r.get("source", "")
        src = f"{source} | {paper}" if source else paper
        parts.append(
            "({}, {}, {}, '{}', 'hard', {}, '[]', '', '', {}, 1, {}, NOW())".format(
                SUBJECT_ID, SUB_SUBJECT_ID, CHAPTER_ID, qtype,
                esc(content), int(r.get("year", 0)), esc(src),
            )
        )
    lines.append(",\n".join(parts) + ";")
    io.open(DST, "w", encoding="utf-8", newline="\n").write("\n".join(lines))
    print(f"写入 {DST}: {len(parts)} 行")


if __name__ == "__main__":
    main()
