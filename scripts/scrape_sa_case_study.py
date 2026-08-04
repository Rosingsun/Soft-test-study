#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
系统分析师（SA）案例分析真题爬虫

数据源：lightsoft.tech（与 scrape_sa_questions.py / scrape_sa_essays.py 同源）
题型映射：questionType=4 表示「案例」

题目模型：
  - 每个案例真题为一个大案例材料（case_material），下面通常包含 3~5 个小题
  - 小题以「【问题 1】/【问题 2】...」或「1. / 2.」形式出现
  - 小题类型为简答题（short），录入时 type='case_study'，
    case_material 字段统一存放案例背景材料，content 存放小题题干，analysis 存放解析

使用方法：
  python scrape_sa_case_study.py --year 2023
  python scrape_sa_case_study.py --all          # 抓取所有年份
  python scrape_sa_case_study.py --year 2022 --out case_2022.json

输出：JSON 数组，可直接交给 seed 脚本或转换为 INSERT SQL 入库。
"""

import argparse
import json
import re
import sys
import time

try:
    import requests
    from bs4 import BeautifulSoup
except ImportError:
    sys.exit("缺少依赖，请先安装：pip install requests beautifulsoup4")

BASE_URL = "https://lightsoft.tech"
# 系统分析师真题列表页（案例分析）
LIST_API = f"{BASE_URL}/api/pc/tests/real"
HEADERS = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    "Accept": "application/json",
}

# 题型映射（与后端 model/question.go 保持一致）
QUESTION_TYPE = "case_study"
DIFFICULTY = "medium"
SUBJECT_ID = 2  # 系统分析师
SUB_SUBJECT_ID = 7  # 系统分析师-案例分析

# 抓取配置
YEARS = [2019, 2020, 2021, 2022, 2023, 2024]


def fetch_case_list(year):
    """获取指定年份的案例分析真题列表"""
    resp = requests.get(
        LIST_API,
        params={
            "examId": 2,          # 系统分析师 exam 标识
            "questionType": 4,    # 案例
            "year": year,
            "pageNum": 1,
            "pageSize": 50,
        },
        headers=HEADERS,
        timeout=20,
    )
    resp.raise_for_status()
    data = resp.json()
    # 兼容两种返回结构
    records = data.get("data", {}).get("list") or data.get("list") or data.get("data") or []
    return records


def fetch_case_detail(test_id):
    """获取单道案例真题的详情（含案例材料 + 小题）"""
    resp = requests.get(
        f"{BASE_URL}/api/pc/tests/real/{test_id}",
        headers=HEADERS,
        timeout=20,
    )
    resp.raise_for_status()
    return resp.json().get("data", {})


def split_sub_questions(content):
    """
    将案例材料 + 小题拆分为结构化数据。
    返回 (case_material, [ {content, analysis}, ... ])
    """
    # 常见小题标记：【问题 1】 【问题2】 (1) 1. 等
    pattern = re.compile(
        r"(?:【\s*问题\s*\d+\s*】|【\s*\d+\s*】|\(?\d+\)\s+|\d+\.\s+)"
    )
    matches = list(pattern.finditer(content))
    if not matches:
        return content.strip(), []

    case_material = content[: matches[0].start()].strip()
    subs = []
    for i, m in enumerate(matches):
        start = m.start()
        end = matches[i + 1].start() if i + 1 < len(matches) else len(content)
        block = content[start:end].strip()
        # 题干与解析通常以「解析：」「参考答案：」分隔
        answer_kw = re.search(r"(?:参考答案|解析)[：:]", block)
        if answer_kw:
            q = block[: answer_kw.start()].strip()
            analysis = block[answer_kw.start():].strip()
        else:
            q = block
            analysis = ""
        subs.append({"content": q, "analysis": analysis})
    return case_material, subs


def parse_case(record, year):
    """解析一条案例记录，返回若干小题 question dict"""
    test_id = record.get("id")
    detail = fetch_case_detail(test_id)
    raw_content = detail.get("content") or detail.get("title") or record.get("content") or ""
    if not raw_content:
        return []

    case_material, subs = split_sub_questions(raw_content)
    if not subs:
        # 无法拆分，整段作为单个 case_study
        subs = [{"content": raw_content.strip(), "analysis": detail.get("analysis", "")}]

    questions = []
    for idx, sub in enumerate(subs, start=1):
        questions.append({
            "subject_id": SUBJECT_ID,
            "sub_subject_id": SUB_SUBJECT_ID,
            "chapter_id": 0,
            "type": QUESTION_TYPE,
            "difficulty": DIFFICULTY,
            "content": sub["content"],
            "case_material": case_material,
            "options": {},
            "answer": "",
            "analysis": sub["analysis"],
            "year": year,
        })
    return questions


def main():
    parser = argparse.ArgumentParser(description="系统分析师案例分析真题爬虫")
    parser.add_argument("--year", type=int, help="抓取的年份")
    parser.add_argument("--all", action="store_true", help="抓取所有年份")
    parser.add_argument("--out", type=str, default="sa_case_study.json", help="输出 JSON 路径")
    args = parser.parse_args()

    years = YEARS if args.all else ([args.year] if args.year else [2023])
    all_questions = []

    for year in years:
        print(f"==> 抓取 {year} 年案例分析真题")
        try:
            records = fetch_case_list(year)
        except Exception as e:
            print(f"  列表获取失败：{e}")
            continue
        print(f"  共 {len(records)} 条记录")
        for rec in records:
            try:
                qs = parse_case(rec, year)
                all_questions.extend(qs)
                time.sleep(0.5)  # 礼貌性限速
            except Exception as e:
                print(f"  解析失败（id={rec.get('id')}）：{e}")

    with open(args.out, "w", encoding="utf-8") as f:
        json.dump(all_questions, f, ensure_ascii=False, indent=2)
    print(f"\n完成，共抓取 {len(all_questions)} 道小题，已写入 {args.out}")


if __name__ == "__main__":
    main()
