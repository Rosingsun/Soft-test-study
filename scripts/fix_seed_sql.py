# -*- coding: utf-8 -*-
"""解析 seed_questions.sql / seed_questions_batch2.sql，生成带正确 sub_subject_id 的插入脚本。

旧库的题库 SQL 中 sub_subject_id 全部硬编码为 1，需按科目重映射到新库：
    subject 1(程序员)  -> 13 基础知识
    subject 4(软件设计师) -> 19 基础知识
    subject 5(网络工程师) -> 21 基础知识
    subject 6(数据库系统工程师) -> 23 基础知识
    subject 8(信息系统项目管理师) -> 1 综合知识
"""
import io
import os

BASE = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..")
SUB_SUBJECT_MAP = {1: 13, 4: 19, 5: 21, 6: 23, 8: 1}
SUBJECT_SUBNAME = {
    1: "程序员-基础知识",
    4: "软件设计师-基础知识",
    5: "网络工程师-基础知识",
    6: "数据库系统工程师-基础知识",
    8: "信息系统项目管理师-综合知识",
}


def find_string_ranges(text):
    """返回 (start, end) 列表，表示所有单引号字符串的区间（含引号）。
    支持 SQL 的 \\' 与 '' 转义。"""
    ranges = []
    i = 0
    n = len(text)
    while i < n:
        if text[i] == "'":
            bc = 0
            j = i - 1
            while j >= 0 and text[j] == "\\":
                bc += 1
                j -= 1
            if bc % 2 == 1:
                i += 1
                continue
            j = i + 1
            while j < n:
                if text[j] == "\\":
                    j += 2
                    continue
                if text[j] == "'":
                    if j + 1 < n and text[j + 1] == "'":
                        j += 2
                        continue
                    break
                j += 1
            ranges.append((i, j + 1))
            i = j + 1
        else:
            i += 1
    return ranges


def in_range(pos, ranges):
    for s, e in ranges:
        if s <= pos < e:
            return True
        if pos < s:
            return False
    return False


def split_statements(text):
    """按顶层分号切分，返回每个语句字符串。"""
    ranges = find_string_ranges(text)
    stmts = []
    start = 0
    for i, ch in enumerate(text):
        if ch == ";" and not in_range(i, ranges):
            stmts.append(text[start:i + 1])
            start = i + 1
    if start < len(text) and text[start:].strip():
        stmts.append(text[start:])
    return stmts


def parse_statement(stmt):
    """从 INSERT 语句中解析出行元组列表（含外层括号）。"""
    vals_idx = stmt.find("VALUES")
    if vals_idx < 0:
        return []
    body = stmt[vals_idx + len("VALUES"):]
    # body 应从第一个 '(' 开始
    first = body.find("(")
    if first < 0:
        return []
    # 找对应的最后一个 ')'
    ranges = find_string_ranges(body)
    depth = 0
    end = -1
    for i in range(first, len(body)):
        if body[i] == "(" and not in_range(i, ranges):
            depth += 1
        elif body[i] == ")" and not in_range(i, ranges):
            depth -= 1
            if depth == 0:
                end = i
    return split_rows(body[first:end + 1])


def split_rows(block):
    """把连续的 (row),(row),... 拆成行列表。"""
    ranges = find_string_ranges(block)
    rows = []
    depth = 0
    start = -1
    for i, ch in enumerate(block):
        if ch == "(" and not in_range(i, ranges):
            if depth == 0:
                start = i
            depth += 1
        elif ch == ")" and not in_range(i, ranges):
            depth -= 1
            if depth == 0 and start >= 0:
                rows.append(block[start:i + 1])
                start = -1
    return rows


def parse_row(row):
    """解析单个 (v1, v2, ...) 行，返回值列表（含引号）。"""
    inner = row.strip()
    if inner.startswith("(") and inner.endswith(")"):
        inner = inner[1:-1]
    ranges = find_string_ranges(inner)
    vals = []
    depth = 0
    buf = ""
    for i, ch in enumerate(inner):
        if ch == "(" and not in_range(i, ranges):
            depth += 1
            buf += ch
        elif ch == ")" and not in_range(i, ranges):
            depth -= 1
            buf += ch
        elif ch == "," and depth == 0 and not in_range(i, ranges):
            vals.append(buf.strip())
            buf = ""
        else:
            buf += ch
    if buf.strip():
        vals.append(buf.strip())
    return vals


def unquote(v):
    v = v.strip()
    if not (v.startswith("'") and v.endswith("'")):
        return v
    v = v[1:-1]
    out = []
    i = 0
    n = len(v)
    while i < n:
        if v[i] == "\\" and i + 1 < n:
            out.append(v[i + 1])
            i += 2
        elif v[i] == "'" and i + 1 < n and v[i + 1] == "'":
            out.append("'")
            i += 2
        else:
            out.append(v[i])
            i += 1
    return "".join(out)


def esc(s):
    return "'" + s.replace("\\", "\\\\").replace("'", "''") + "'"


def process(src, dst):
    text = io.open(src, encoding="utf-8", errors="replace").read()
    rows = []
    for stmt in split_statements(text):
        if "INSERT INTO `questions`" not in stmt:
            continue
        rows.extend(parse_statement(stmt))
    print(f"{os.path.basename(src)}: 解析出 {len(rows)} 行")

    lines = [
        "SET NAMES utf8mb4;",
        "INSERT INTO `questions` "
        "(`subject_id`, `sub_subject_id`, `type`, `difficulty`, `content`, `options`, `answer`, `analysis`, `year`, `status`, `source`, `created_at`) VALUES",
    ]
    parts = []
    for r in rows:
        vals = parse_row(r)
        if len(vals) != 11:
            print(f"  跳过异常行（列数 {len(vals)}）: {r[:100]}")
            continue
        subject_id = int(vals[0])
        ssub = SUB_SUBJECT_MAP.get(subject_id)
        if ssub is None:
            print(f"  跳过未知科目 subject_id={subject_id}")
            continue
        content = unquote(vals[4])
        options = unquote(vals[5]) if vals[5].upper() != "NULL" else "[]"
        answer = unquote(vals[6])
        analysis = unquote(vals[7]) if vals[7].upper() != "NULL" else ""
        year = vals[8]
        status = vals[9]
        source = f"seed:{SUBJECT_SUBNAME[subject_id]}"
        parts.append(
            "({}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, NOW())".format(
                subject_id, ssub, vals[2], vals[3],
                esc(content), esc(options), esc(answer), esc(analysis),
                year, status, esc(source),
            )
        )
    lines.append(",\n".join(parts) + ";")
    io.open(dst, "w", encoding="utf-8", newline="\n").write("\n".join(lines))
    print(f"  已写入 {dst}（{len(parts)} 行）")
    return len(parts)


if __name__ == "__main__":
    files = [
        (os.path.join(BASE, "seed_questions.sql"), os.path.join(BASE, "data", "seed_questions_fixed.sql")),
        (os.path.join(BASE, "seed_questions_batch2.sql"), os.path.join(BASE, "data", "seed_questions_batch2_fixed.sql")),
    ]
    total = 0
    for src, dst in files:
        total += process(src, dst)
    print("总计:", total)
