# -*- coding: utf-8 -*-
"""系统分析师-综合知识题目二次精细分类脚本

背景：cmd/import_real 的关键词评分制已把大部分真题归入考纲 18 章，
但仍有约 150 题不含任何专业术语（数学应用题、英语题、图表描述题等）
堆积在兜底章「软件工程」(chapter_id=214)。

本脚本从数据库拉取该子科目下全部题目，使用更强的多信号规则
（正则模式 + 特征组合）进行二次分类，并直接更新 questions.chapter_id。
仅调整 sub_subject_id=4 范围内的题目，不影响其他科目。

用法：
    python scripts/reclassify_sa_chapters.py            # 试运行（dry-run，仅打印报告）
    python scripts/reclassify_sa_chapters.py --apply    # 实际写入数据库
"""

import os
import re
import subprocess
import sys

MYSQL = r"D:\ProgramFile\ProgramEnvironment\mysql-8.0.24-winx64\bin\mysql.exe"
DB_ARGS = ["-h", "43.157.65.164", "-P", "3306", "-u", "SoftTestStudyT",
           "softteststudyt", "--default-character-set=utf8mb4", "--batch", "--raw"]
APPLY = "--apply" in sys.argv

# 系统分析师-综合知识子科目与各章节 ID（与 backend/database/schema.sql 一致）
SUB_SUBJECT_ID = 4
CHAPTERS = {
    "绪论": 200,
    "法律法规与标准化": 201,
    "数学基础": 202,
    "运筹学基础": 203,
    "数据结构与算法": 204,
    "计算机组成与体系结构": 205,
    "操作系统": 206,
    "程序设计语言与语言处理": 207,
    "嵌入式系统": 208,
    "计算机网络": 209,
    "分布式系统与中间件": 210,
    "多媒体基础": 211,
    "数据库系统": 212,
    "企业信息化": 213,
    "软件工程": 214,
    "面向对象方法与设计模式": 215,
    "项目管理": 216,
    "信息安全": 217,
}
FALLBACK = CHAPTERS["软件工程"]


def run_mysql(sql):
    """执行只读查询并按 TSV 解析返回行列表（每行为字符串列列表）"""
    env = dict(os.environ)
    env["MYSQL_PWD"] = "C5SbSrEzBBLSZBaZ"
    r = subprocess.run([MYSQL] + DB_ARGS + ["-e", sql],
                       capture_output=True, env=env)
    out = r.stdout.decode("utf-8", errors="replace")
    lines = [ln for ln in out.split("\n") if ln.strip()]
    return [ln.split("\t") for ln in lines[1:]]


# ---------------------------------------------------------------------------
# 二级分类规则：每章一组 (正则模式, 权重)。命中即累加权重，取最高分章节。
# 规则侧重 import_real 关键词法覆盖不到的特征：
#   - 数学/运筹类应用计算题的叙述特征
#   - 英语题的英文考点词
#   - 图表类题干（前趋图 / 表格 / 路线图）所属知识域推断
# ---------------------------------------------------------------------------
RULES = {
    "运筹学基础": [
        (r"(最大|最小)[值化流]", 2), (r"线性规划|单纯形|对偶", 4),
        (r"(运输|分配|指派|派遣).{0,12}(问题|方案)", 3),
        (r"(工人|机器|任务|应聘者?).{2,20}分派|定岗", 3),
        (r"盈亏平衡|期望收益|后悔值|乐观.{0,3}悲观", 3),
        (r"(总效益|总费用|总运量|最大流量|最低成本)", 3),
        (r"博弈|纳什均衡", 4), (r"排队论|M/M/1", 4),
        (r"决策(树|方案|分析)", 2), (r"整数规划|0-1规划|分支限界", 3),
        (r"(如下表|下表).{0,40}(效益|得分|成本|费用|利润)", 2),
    ],
    "项目管理": [
        (r"关键路径|总时差|自由时差", 4),
        (r"(最早|最迟)(开始|完成)(时间)?", 3),
        (r"(总)?工期|赶进度|赶工", 3),
        (r"挣值|成本偏差|进度偏差|\bCV\b|\bSV\b|\bCPI\b|\bSPI\b", 4),
        (r"(项目|工程).{0,6}(管理|监控|收尾|立项|验收)", 2),
        (r"甘特图|里程碑|WBS|工作分解", 4),
        (r"(风险识别|风险应对|风险管理)", 3),
        (r"沟通渠道|干系人", 3),
        (r"(招标|投标|评标|中标|索赔|合同类型)", 2),
        (r"\d+\s*[个道]\s*(作业|工序|活动).{0,30}(紧前|衔接|先后)", 3),
    ],
    "数据结构与算法": [
        (r"二叉树|哈夫曼|平衡二叉树|红黑树", 4),
        (r"(时间|空间)复杂度", 4),
        (r"排序算法|(冒泡|快速|归并|堆|基数|希尔)排序|稳定排序", 4),
        (r"散列|哈希表|冲突", 2),
        (r"邻接矩阵|邻接表|拓扑排序|AOV|AOE", 4),
        (r"(栈|队列).{0,10}(特性|应用|操作|先进后出|先进先出|LIFO|FIFO)", 3),
        (r"(折半|二分)查找|顺序查找", 4),
        (r"(贪心|动态规划|分治|回溯|分支限界)法?", 3),
        (r"(循环|链式|单链|双链)表|顺序存储结构", 2),
        (r"(表达式求值|逆波兰|后缀表达式|中缀)", 4),
    ],
    "数学基础": [
        (r"概率|随机变量|数学期望|方差|正态分布", 4),
        (r"排列|组合数|鸽巢|抽屉原理|容斥", 4),
        (r"命题|谓词|真值表|推理(不)?正确", 3),
        (r"(集合|关系){0,2}(运算|矩阵|性质)|偏序|等价关系", 2),
        (r"插值|拟合|数值积分|迭代求解|牛顿迭代|误差", 3),
        (r"矩阵|行列式|线性方程组", 3),
        (r"(成绩|分数).{0,20}(变换|转换|统计)", 2),
    ],
    "程序设计语言与语言处理": [
        (r"文法|正规式|正则表达式|有限自动机", 4),
        (r"(词法|语法|语义)分析|中间代码|目标代码", 4),
        (r"(传值调用|引用调用|传地址|值传递|地址传递)", 4),
        (r"(编译|解释)(器|程序|执行|方式)", 3),
        (r"(静态|动态)(绑定|作用域)|闭包", 4),
        (r"(后缀|前缀|中缀)(表达式|式)", 4),
        (r"(标识符|关键字|保留字)", 2),
        (r"(C|Java|Python|Fortran|Prolog|Lisp|SQL)\s*(语言|程序)?\s*(中|是|属于)", 2),
    ],
    "分布式系统与中间件": [
        (r"CAP定理|CAP 定理|BASE理论|最终一致", 4),
        (r"Paxos|Raft|ZooKeeper|共识算法", 4),
        (r"(微服务|消息队列|发布订阅|服务治理|服务网关|注册中心)", 4),
        (r"(负载均衡|缓存穿透|Redis|Memcached|分库分表|读写分离)", 3),
        (r"(分布式事务|TCC补偿|Saga|柔性事务)", 3),
        (r"SOA|ESB|RPC|CORBA|EJB", 3),
        (r"(Kubernetes|Docker|容器编排|云原生)", 2),
        (r"一致性哈希|数据分片|哈希环", 4),
    ],
    "多媒体基础": [
        (r"多媒体|图像(存储)?容量|分辨率|像素深度|色深", 4),
        (r"采样(频率|定理)|量化位数|帧率|刷新率", 4),
        (r"MPEG|JPEG|H\.26\d|MP3|AVI|MOV", 4),
        (r"(有损|无损)压缩|行程编码|预测编码|变换编码", 4),
        (r"(音频|视频|声音|图像)(编码|格式|文件|压缩)", 3),
        (r"色彩空间|RGB|YUV|CMYK|PAL|NTSC", 3),
    ],
    "嵌入式系统": [
        (r"嵌入式|RTOS|实时操作系统|片上系统|SoC", 4),
        (r"(交叉编译|JTAG|看门狗|Bootloader|在线仿真)", 4),
        (r"(单片机|微控制器|ARM|DSP)芯片?", 3),
        (r"(板级|片级|系统级)初始化", 4),
        (r"固件|实时任务|优先级反转", 3),
    ],
    "计算机网络": [
        (r"(子网掩码|子网划分|无类域间|CIDR|网段|广播域)", 4),
        (r"(路由器|路由协议|RIP|OSPF|BGP|下一跳)", 3),
        (r"IP(报文|地址|首部|数据报)|(网络)?地址解析", 3),
        (r"(TCP|UDP|三次握手|四次挥手|滑动窗口|拥塞)", 3),
        (r"(HTTP|HTTPS|FTP|SMTP|POP3|DNS|DHCP|SNMP|Telnet)", 3),
        (r"(以太网|CSMA|令牌|交换机|VLAN|网桥|集线器)", 3),
        (r"(曼彻斯特|奈奎斯特|香农|波特率|比特率|多路复用|信道)", 3),
        (r"(QoS|IntServ|DiffServ|RSVP|资源预留|MPLS|SDN)", 4),
        (r"(无线局域网|Wi-?Fi|蓝牙|ZigBee|5G|4G|IPv6)", 3),
        (r"(拓扑|星型|总线型|环型|网状)结构", 2),
    ],
    "计算机组成与体系结构": [
        (r"(流水线|吞吐率|加速比|指令周期|指令系统|寻址方式)", 3),
        (r"(Cache|高速缓存|命中率|组相联|直接映射|全相联)", 4),
        (r"(RAID|冗余阵列|磁盘阵列|磁盘存取)", 4),
        (r"(原码|反码|补码|移码|浮点数|规格化|阶码|尾数)", 4),
        (r"(海明码|CRC|校验码|奇偶校验)", 3),
        (r"(RISC|CISC|Flynn|SIMD|MISD|MIMD|多核|并行)", 3),
        (r"(中断|DMA|通道|总线|I/O接口|输入输出)", 2),
        (r"(可靠性|可用性|MTBF|MTTR|串联系统|并联系统|冗余度)", 2),
        (r"(运算器|控制器|寄存器|存储容量|字节编址|按字编址)", 2),
        (r"(相联存储器|虚拟存储|主存|辅存|光盘|硬盘)", 1),
    ],
    "信息安全": [
        (r"(加密|解密|密钥|明文|密文)", 3),
        (r"(RSA|DES|AES|ECC|ElGamal|国密|SM[234])算法?", 4),
        (r"(数字签名|数字证书|PKI|CA认证|X\.509)", 4),
        (r"(防火墙|入侵检测|IDS|IPS|蜜罐|WAF)", 4),
        (r"(病毒|木马|蠕虫|恶意代码|钓鱼|DDoS|拒绝服务|洪泛)", 3),
        (r"(访问控制|RBAC|自主访问|强制访问|最小权限)", 4),
        (r"(SSL|TLS|IPSec|VPN|Kerberos|SET协议)", 4),
        (r"(报文摘要|MD5|SHA-?\d*|Hash|哈希函数)", 3),
        (r"(安全审计|等级保护|灾备|容灾|RPO|RTO|渗透测试)", 3),
        (r"(弱口令|口令扫描|端口扫描|缓冲区溢出|SQL注入|XSS|CSRF)", 4),
    ],
    "法律法规与标准化": [
        (r"(著作权|专利权|专利法|商标|署名权|发表权|保护期限)", 4),
        (r"(软件保护条例|著作权登记|版权|侵权)", 3),
        (r"(招投标法|招标投标|政府采购法|合同法|网络安全法)", 4),
        (r"(国家标准|行业标准|国际标准|GB/T|强制性标准|推荐性标准|标准化法)", 3),
        (r"(知识产权|职务作品|委托开发|软件著作权)", 3),
        (r"(许可|转让).{0,6}(使用权|著作权)", 2),
    ],
    "绪论": [
        (r"系统分析师(的)?(角色|职责|定位)", 4),
        (r"(信息系统)?(生存|生命)周期", 3),
        (r"职业道德|职业规范|行为准则", 4),
        (r"(需求获取|原型法|焦点小组|访谈|问卷).{0,15}(方法|阶段)?", 2),
        (r"软件危机|系统工程|基于模型的系统工程", 2),
    ],
    "面向对象方法与设计模式": [
        (r"UML|(用例图|类图|对象图|时序图|顺序图|通信图|状态图|活动图|构件图|部署图|包图)", 4),
        (r"(设计模式|单例|工厂方法|抽象工厂|观察者|策略|适配器|装饰器|代理|外观|桥接|享元|责任链|中介者|备忘录|迭代器|访问者|命令模式|状态模式|组合模式)", 4),
        (r"(封装|继承|多态|重载|重写|泛型)", 2),
        (r"(开闭原则|里氏替换|依赖倒置|接口隔离|单一职责|SOLID)", 4),
        (r"(对象|类).{0,6}(属性|方法|实例化|消息)", 2),
        (r"(include|extend|包含|扩展).{0,8}(关系)?", 1),
    ],
    "企业信息化": [
        (r"(ERP|CRM|SCM|MRP|PLM)(系统)?", 4),
        (r"(客户关系管理|供应链管理|企业资源规划|产品生命周期管理)", 4),
        (r"(电子商务|电子政务|B2B|B2C|C2C|O2O|G2B|G2C)", 3),
        (r"(信息化|数字化转型|智能制造|工业互联网|两化融合)", 3),
        (r"(业务流程重组|业务流程再造|BPR|流程建模|工作流)", 3),
        (r"(决策支持|商业智能|数据仓库|数据挖掘|BI)", 2),
        (r"(IT服务|ITIL|ITSS|服务级别协议|SLA|运维管理|服务运营)", 3),
        (r"(知识管理|显性知识|隐性知识|知识库)", 3),
        (r"(战略|规划).{0,8}(信息|数字化)", 2),
    ],
    "操作系统": [
        (r"(进程|线程).{0,10}(同步|互斥|调度|切换|通信|状态)", 3),
        (r"(信号量|PV操作|pv操作|P、V操作|管程|临界区|死锁|银行家)", 4),
        (r"(页面置换|LRU|FIFO|缺页|虚拟存储|分页|分段|段页)", 3),
        (r"(磁盘调度|SCAN|SSTF|索引节点|位示图|文件目录|空闲空间)", 3),
        (r"(前趋图|前驱图)", 2),
        (r"(作业调度|时间片轮转|响应比|优先级调度|多级反馈)", 3),
        (r"(Spooling|缓冲|设备管理|I/O管理)", 2),
    ],
    "数据库系统": [
        (r"(关系模式|候选码|候选键|主键|外键|函数依赖|无损连接|保持依赖)", 4),
        (r"(范式|1NF|2NF|3NF|BCNF|模式分解)", 4),
        (r"(事务|并发控制|封锁|两段锁|隔离级别|脏读|幻读)", 3),
        (r"(视图|触发器|存储过程|游标|授权|GRANT)", 3),
        (r"(E-R图|ER图|实体联系|联系类型|1:n|m:n)", 3),
        (r"(数据仓库|OLAP|OLTP|数据挖掘|NoSQL|反规范化)", 3),
        (r"(分布式数据库|分片|透明性|两阶段提交)", 3),
        (r"(SELECT|CREATE TABLE|GRANT|REVOKE)", 2),
    ],
}

EN_RULES = {
    # 英语题按英文考点词归类
    "面向对象方法与设计模式": [(r"UML|object[- ]oriented|use ?case|polymorphi|inheritan", 3)],
    "软件工程": [(r"requirement|systems? analysis|software (design|testing|maintenance|process)|waterfall|SDLC", 3)],
    "计算机网络": [(r"network|protocol|router|TCP|firewall", 3)],
    "数据库系统": [(r"database|SQL|normaliz", 3)],
}


def classify(text):
    """返回得分最高的章节 ID"""
    ascii_ratio = sum(1 for c in text if c.isascii() and c.isalpha()) / max(len(text), 1)
    rules = EN_RULES if ascii_ratio > 0.55 else RULES
    best_name, best_score = None, 0
    total_hits = 0
    for name, patterns in rules.items():
        score = sum(w for pat, w in patterns if (hits := len(re.findall(pat, text))) and (total_hits := total_hits + hits))
        if score > best_score:
            best_name, best_score = name, score
    return (CHAPTERS[best_name], best_score) if best_name else (None, 0)


def esc(s):
    return s.replace("\\", "\\\\").replace("'", "''")


def main():
    print("拉取题目 ...")
    rows = [
        (int(cols[0]), " ".join(cols[1:]))
        for cols in run_mysql(
            "SELECT id, CONCAT_WS(' ', content, options, analysis, answer) "
            f"FROM questions WHERE subject_id=9 AND sub_subject_id={SUB_SUBJECT_ID} AND status=1")
    ]
    print(f"共 {len(rows)} 题")

    current_chapter = {int(c[0]): int(c[1]) for c in run_mysql(
        f"SELECT id, chapter_id FROM questions WHERE sub_subject_id={SUB_SUBJECT_ID}")}

    moves = {}
    from collections import Counter
    dist = Counter()
    for qid, text in rows:
        target, score = classify(text.replace("\\N", " "))
        if target is None or target == FALLBACK:
            dist["保持软件工程(兜底)"] += 1
            continue
        if current_chapter.get(qid) != target:
            moves[qid] = target
            label = next(n for n, i in CHAPTERS.items() if i == target)
            dist[label] += 1
        else:
            dist["无需变更"] += 1

    print("\n=== 分类报告 ===")
    for k, v in sorted(dist.items(), key=lambda x: -x[1]):
        print(f"  {k}: {v}")
    print(f"待迁移: {len(moves)} 题")

    if APPLY:
        ids_by_chapter = Counter(moves.values())
        for target, cnt in ids_by_chapter.items():
            id_list = ",".join(str(q) for q, t in moves.items() if t == target)
            run_mysql(f"UPDATE questions SET chapter_id={target} WHERE id IN ({id_list})")
            print(f"已将 {cnt} 题迁入 chapter_id={target}")
        print("写入完成")
    else:
        print("(dry-run 模式，未写入。加 --apply 执行写入)")


if __name__ == "__main__":
    main()
