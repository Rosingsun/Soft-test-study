# -*- coding: utf-8 -*-
"""将 data/sa_questions_standard.json 生成批量 INSERT SQL（与 import_real 逻辑一致）。

字段格式对齐 backend/cmd/import_real/main.go：
- subject_id=9（系统分析师），sub_subject_id=4（综合知识）
- options: ["A. 文本", "B. 文本", ...]（按键排序）
- type: single/multi -> single
- difficulty: medium, status: 1
- source: "{source} | {paper} | {章节}"

注意：系统分析师-综合知识的章节分类现按《系统分析师教程（第2版）》第一篇
「基础知识」9 章（绪论 / 数学与工程基础 / 计算机系统 / 计算机网络与分布式系统 /
数据库系统 / 企业信息化 / 软件工程 / 项目管理 / 信息安全）自动归类，最终由
backend/cmd/import_real/main.go 在导入时写入正确的 chapter_id。本脚本生成的 SQL
仅作历史/离线用途，chapter_id 留空交由 seed 处理，但综合知识题目不再随机分配章节。
"""
import io
import json
import os

BASE = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..")
SRC = os.path.join(BASE, "data", "sa_questions_standard.json")
DST = os.path.join(BASE, "data", "sa_questions_real.sql")

SUBJECT_ID = 9
SUB_SUBJECT_ID = 4  # 系统分析师-综合知识

# 第二版教材综合知识 9 章关键词（顺序即匹配优先级）
CHAPTER_KEYWORDS = [
    ("绪论", ["绪论", "系统分析师", "系统分析", "角色", "职业道德", "法律法规", "标准化", "知识产权", "标准"]),
    ("数学与工程基础", ["图论", "运筹", "线性规划", "排队论", "博弈论", "组合数学", "概率论", "数理统计", "矩阵", "离散", "数值", "误差", "最优化", "数学建模", "微积分", "集合", "逻辑代数", "布尔"]),
    ("计算机系统", ["CPU", "处理器", "指令系统", "冯·诺依曼", "存储系统", "Cache", "内存", "寄存器", "总线", "体系结构", "RISC", "CISC", "流水线", "并行处理", "多核", "嵌入式", "固件", "芯片", "操作系统", "进程", "线程", "死锁", "虚拟内存", "文件系统", "编译", "汇编"]),
    ("计算机网络与分布式系统", ["网络", "TCP", "IP", "HTTP", "DNS", "路由", "交换机", "OSI", "协议", "子网", "网关", "分布式", "微服务", "集群", "负载均衡", "CDN", "区块链", "P2P", "通信", "带宽", "加密传输"]),
    ("数据库系统", ["数据库", "关系模型", "SQL", "ER", "范式", "事务", "并发控制", "索引", "数据仓库", "数据挖掘", "OLAP", "OLTP", "NoSQL", "分布式数据库", "视图", "触发器", "游标"]),
    ("企业信息化", ["企业信息化", "ERP", "CRM", "SCM", "电子商务", "电子政务", "智能制造", "工业互联网", "数字化转型", "业务流程", "BPR", "知识管理", "决策支持", "DSS", "BI", "大数据", "数据治理"]),
    ("软件工程", ["软件工程", "需求", "用例", "UML", "面向对象", "设计模式", "软件测试", "单元测试", "集成测试", "维护", "重构", "敏捷", "瀑布", "CMMI", "质量", "软件度量", "配置管理", "版本", "建模", "状态图", "活动图", "类图"]),
    ("项目管理", ["项目管理", "进度", "关键路径", "CPM", "PERT", "甘特图", "成本", "估算", "挣值", " Risk", "风险", "WBS", "范围", "里程碑", "资源", "招投标", "合同", "质量保证"]),
    ("信息安全", ["信息安全", "加密", "对称", "非对称", "RSA", "DES", "AES", "哈希", "数字签名", "认证", "防火墙", "入侵", "病毒", "木马", "安全", "PKI", "证书", "访问控", "权限", "审计", "隐私"]),
]
FALLBACK_CHAPTER = "软件工程"


def esc(s):
    return "'" + str(s).replace("\\", "\\\\").replace("'", "''") + "'"


def classify_chapter(rec):
    text = rec.get("content", "")
    opts = rec.get("options") or {}
    for k, v in opts.items():
        text += " " + str(k) + " " + str(v)
    text += " " + rec.get("analysis", "")
    for name, kws in CHAPTER_KEYWORDS:
        for kw in kws:
            if kw in text:
                return name
    return FALLBACK_CHAPTER


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
        chapter = classify_chapter(r)
        src = f"{source} | {paper} | {chapter}" if source else f"{paper} | {chapter}"
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
