# -*- coding: utf-8 -*-
"""校验与清理 sa_questions_lightsoft.json，输出标准化版本 sa_questions_standard.json"""
import io
import json
import os
import re

SRC = os.path.join(os.path.dirname(__file__), "..", "data", "sa_questions_lightsoft.json")
OUT = os.path.join(os.path.dirname(__file__), "..", "data", "sa_questions_standard.json")

data = json.load(io.open(SRC, encoding="utf-8"))
print("读取:", len(data), "题")

# 修复残留 HTML
def clean_more(text):
    text = text.replace("&nbsp;", " ")
    text = re.sub(r"<br\s*/?>", " ", text, flags=re.I)
    text = re.sub(r"<[^>]+>", "", text)
    text = re.sub(r"\s+", " ", text)
    return text.strip()

for q in data:
    q["content"] = clean_more(q["content"])
    q["options"] = {k: clean_more(v) for k, v in q["options"].items()}
    q["analysis"] = clean_more(q["analysis"])

# 检查字段一致性
assert all(q["subject"] == "系统分析师" for q in data)
assert all(q["type"] == "single" for q in data)
assert all(q["answer"] in "ABCDE" for q in data)
assert all(4 <= len(q["options"]) <= 5 for q in data)

# 题干去重（去掉首尾空白、压缩空白的归一化）
def norm(s):
    return re.sub(r"\s+", "", s)

seen = {}
dups = 0
kept = []
for q in data:
    key = norm(q["content"])
    if key in seen:
        dups += 1
        continue
    seen[key] = q
    kept.append(q)

print("去重后:", len(kept), "（去掉", dups, "重复）")

# 统计
from collections import Counter
print("年份分布:", dict(Counter(q["year"] for q in kept)))
print("有无解析:", Counter(bool(q["analysis"]) for q in kept))
print("选项数:", dict(Counter(len(q["options"]) for q in kept)))

# 答案分布抽查
ans_dist = Counter(q["answer"] for q in kept)
print("答案字母分布:", dict(ans_dist))

with io.open(OUT, "w", encoding="utf-8") as f:
    json.dump(kept, f, ensure_ascii=False, indent=1)
print("写入:", OUT, len(kept), "题")
