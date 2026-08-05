# -*- coding: utf-8 -*-
"""抓取 lightsoft.tech 系统分析师（高级）论文历年真题，输出统一 JSON。

论文题为写作题，无客观选项与标准答案，仅记录写作主题。

数据字段：
    {
        "subject": "系统分析师",
        "year": 2025,
        "half": "上半年",          # 上半年 / 下半年
        "paper": "2025年 上半年 论文",
        "type": "essay",
        "content": "题干纯文本",
        "options": {},
        "answer": "",              # 论文题无标准答案
        "analysis": "",
        "source": "lightsoft.tech"
    }
"""

import datetime
import html
import json
import os
import re
import sys
import time
from urllib.parse import quote, unquote
from urllib.request import Request, urlopen

SUBJECT_ID = "057641fa9f9540b1877ab054adc12b65"  # 高级 系统分析师
SUBJECT_NAME = "系统分析师"
BASE = "https://www.lightsoft.tech"
SUBJECT_URL = (
    f"{BASE}/doquestion/subject?doType=0&subjectId={SUBJECT_ID}"
)
PAPER_URL = (
    f"{BASE}/doquestion/doquestion?userId=null&index={{index}}&doType=0"
    f"&id={{pid}}&name={{name}}&subjectId={SUBJECT_ID}"
)
HEADERS = {
    "User-Agent": (
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
        "(KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
    ),
    "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
    "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
    "Referer": SUBJECT_URL,
}

CURRENT_YEAR = datetime.datetime.now().year
YEAR_MIN, YEAR_MAX = CURRENT_YEAR - 4, CURRENT_YEAR
OUT_DIR = os.path.join(os.path.dirname(__file__), "..", "data")


def fetch(url, retries=3):
    for i in range(retries):
        try:
            req = Request(url, headers=HEADERS)
            with urlopen(req, timeout=30) as resp:
                return resp.read().decode("utf-8")
        except Exception as exc:  # noqa: BLE001
            if i == retries - 1:
                raise
            time.sleep(2 * (i + 1))


def clean_html(text):
    """HTML 片段 → 纯文本：去掉标签、还原实体、合并空白。"""
    if not text:
        return ""
    text = re.sub(r"<br\s*/?>", "\n", text, flags=re.I)
    text = re.sub(r"</p>|</div>|</span>", "\n", text, flags=re.I)
    text = re.sub(r"<[^>]+>", "", text)
    text = html.unescape(text)
    text = re.sub(r"[ \t\u3000]+\n", "\n", text)
    text = re.sub(r"\n{2,}", "\n", text)
    text = text.strip()
    return text


def get_papers():
    raw = fetch(SUBJECT_URL)
    links = re.findall(r'href="([^"]*doquestion/doquestion[^"]*)"', raw)
    papers = []
    for u in links:
        u = html.unescape(u)
        m = re.search(r"id=([a-f0-9]{32})", u)
        nm = re.search(r"name=([^&]+)", u)
        if not m or not nm:
            continue
        name = unquote(nm.group(1))
        pid = m.group(1)
        if "论文" not in name or "模拟" in name:
            continue
        m_year = re.search(r"(20\d{2})年", name)
        year = int(m_year.group(1)) if m_year else 0
        half = "上半年" if "上半年" in name else ("下半年" if "下半年" in name else "")
        if not (YEAR_MIN <= year <= YEAR_MAX):
            continue
        papers.append({"pid": pid, "name": name, "year": year, "half": half})
    return papers


def fetch_question(pid, name, index):
    url = PAPER_URL.format(index=index, pid=pid, name=quote(name))
    raw = fetch(url)
    m = re.search(r'<script id="__NEXT_DATA__" type="application/json">(.*?)</script>', raw, re.S)
    if not m:
        return None, None
    data = json.loads(m.group(1))
    pp = data["props"]["pageProps"]
    return pp.get("qIds"), pp.get("question")


def parse_essay_question(q, source_paper, year, half):
    content = clean_html(q.get("name"))
    if not content:
        return None
    return {
        "subject": SUBJECT_NAME,
        "year": year,
        "half": half,
        "paper": source_paper,
        "type": "essay",
        "content": content,
        "options": {},
        "answer": "",
        "analysis": "",
        "source": "lightsoft.tech",
    }


def main():
    papers = get_papers()
    print(f"发现 {len(papers)} 套论文试卷：")
    for p in papers:
        print(f"  {p['name']}")
    if not papers:
        sys.exit(1)

    all_questions = []
    for p in papers:
        idx = 0
        paper_questions = []
        qids = None
        while True:
            try:
                qids_new, q = fetch_question(p["pid"], p["name"], idx)
            except Exception as exc:  # noqa: BLE001
                print(f"  抓取失败 index={idx}: {exc}")
                break
            if qids is None and qids_new:
                qids = qids_new
            if q is None:
                break
            rec = parse_essay_question(q, p["name"], p["year"], p["half"])
            if rec:
                paper_questions.append(rec)
            idx += 1
            if qids and idx >= len(qids):
                break
            time.sleep(0.35)
        all_questions.extend(paper_questions)
        print(f"{p['name']}: {len(paper_questions)} 题")

    os.makedirs(OUT_DIR, exist_ok=True)
    out_path = os.path.join(OUT_DIR, "sa_essay_questions.json")
    with open(out_path, "w", encoding="utf-8") as f:
        json.dump(all_questions, f, ensure_ascii=False, indent=1)
    print(f"\n共 {len(all_questions)} 题，已写入 {out_path}")


if __name__ == "__main__":
    main()
