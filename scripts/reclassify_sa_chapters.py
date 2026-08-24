# -*- coding: utf-8 -*-
"""系统分析师-综合知识题目二次精细分类脚本

背景：cmd/import_real 的关键词评分制已把大部分真题归入考纲 22 章，
但仍有约 150 题不含任何专业术语（数学应用题、英语题、图表描述题等）
堆积在兜底章「软件工程」(chapter_id=206)。

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
# 按《系统分析师教程（第 2 版 · 2024）》清华大学出版社目录，22 章
SUB_SUBJECT_ID = 4
CHAPTERS = {
    "绪论": 200,
    "数学与工程基础": 201,
    "计算机系统": 202,
    "计算机网络与分布式系统": 203,
    "数据库系统": 204,
    "企业信息化": 205,
    "软件工程": 206,
    "项目管理": 207,
    "信息安全": 208,
    "系统规划与分析": 209,
    "软件需求工程": 210,
    "软件架构设计": 211,
    "系统设计": 212,
    "软件实现与测试": 213,
    "系统运行与维护": 214,
    "Web 应用系统": 215,
    "嵌入式系统": 216,
    "移动应用系统": 217,
    "大数据处理系统": 218,
    "微服务系统": 219,
    "信息物理系统": 220,
    "系统分析师论文写作要点": 221,
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
    "数学与工程基础": [
        # 运筹学/规划类
        (r"(最大|最小)[值化流]", 2), (r"线性规划|单纯形|对偶", 4),
        (r"(运输|分配|指派|派遣).{0,12}(问题|方案)", 3),
        (r"(工人|机器|任务|应聘者?).{2,20}分派|定岗", 3),
        (r"盈亏平衡|期望收益|后悔值|乐观.{0,3}悲观", 3),
        (r"(总效益|总费用|总运量|最大流量|最低成本)", 3),
        (r"博弈|纳什均衡", 4), (r"排队论|M/M/1", 4),
        (r"决策(树|方案|分析)", 2), (r"整数规划|0-1规划|分支限界", 3),
        (r"(如下表|下表).{0,40}(效益|得分|成本|费用|利润)", 2),
        # 概率/统计/离散数学
        (r"概率|随机变量|数学期望|方差|正态分布", 4),
        (r"排列|组合数|鸽巢|抽屉原理|容斥", 4),
        (r"命题|谓词|真值表|推理(不)?正确", 3),
        (r"(集合|关系){0,2}(运算|矩阵|性质)|偏序|等价关系", 2),
        (r"插值|拟合|数值积分|迭代求解|牛顿迭代|误差", 3),
        (r"矩阵|行列式|线性方程组", 3),
        (r"(成绩|分数).{0,20}(变换|转换|统计)", 2),
        # 图论
        (r"最小生成树|最短路径|Prim|Kruskal|Dijkstra|Floyd|拓扑排序", 3),
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
    "计算机系统": [
        # 数据结构与算法
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
        # 组成与体系结构
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
        # 操作系统
        (r"(进程|线程).{0,10}(同步|互斥|调度|切换|通信|状态)", 3),
        (r"(信号量|PV操作|pv操作|P、V操作|管程|临界区|死锁|银行家)", 4),
        (r"(页面置换|LRU|FIFO|缺页|虚拟存储|分页|分段|段页)", 3),
        (r"(磁盘调度|SCAN|SSTF|索引节点|位示图|文件目录|空闲空间)", 3),
        (r"(前趋图|前驱图)", 2),
        (r"(作业调度|时间片轮转|响应比|优先级调度|多级反馈)", 3),
        (r"(Spooling|缓冲|设备管理|I/O管理)", 2),
        # 程序设计语言
        (r"文法|正规式|正则表达式|有限自动机", 4),
        (r"(词法|语法|语义)分析|中间代码|目标代码", 4),
        (r"(传值调用|引用调用|传地址|值传递|地址传递)", 4),
        (r"(编译|解释)(器|程序|执行|方式)", 3),
        (r"(静态|动态)(绑定|作用域)|闭包", 4),
        (r"(后缀|前缀|中缀)(表达式|式)", 4),
        (r"(标识符|关键字|保留字)", 2),
        (r"(C|Java|Python|Fortran|Prolog|Lisp|SQL)\s*(语言|程序)?\s*(中|是|属于)", 2),
        # 多媒体
        (r"多媒体|图像(存储)?容量|分辨率|像素深度|色深", 4),
        (r"采样(频率|定理)|量化位数|帧率|刷新率", 4),
        (r"MPEG|JPEG|H\.26\d|MP3|AVI|MOV", 4),
        (r"(有损|无损)压缩|行程编码|预测编码|变换编码", 4),
        (r"(音频|视频|声音|图像)(编码|格式|文件|压缩)", 3),
        (r"色彩空间|RGB|YUV|CMYK|PAL|NTSC", 3),
    ],
    "计算机网络与分布式系统": [
        (r"CAP定理|CAP 定理|BASE理论|最终一致", 4),
        (r"Paxos|Raft|ZooKeeper|共识算法", 4),
        (r"(消息队列|发布订阅|Kafka|RabbitMQ|RocketMQ)", 4),
        (r"(负载均衡|缓存穿透|Redis|Memcached|分库分表|读写分离)", 3),
        (r"(分布式事务|TCC补偿|Saga|柔性事务|Seata)", 3),
        (r"SOA|ESB|RPC|CORBA|EJB", 3),
        (r"(Kubernetes|Docker|容器编排|云原生)", 2),
        (r"一致性哈希|数据分片|哈希环", 4),
        # 网络
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
    "嵌入式系统": [
        (r"嵌入式|RTOS|实时操作系统|片上系统|SoC", 4),
        (r"(交叉编译|JTAG|看门狗|Bootloader|在线仿真)", 4),
        (r"(单片机|微控制器|ARM|DSP)芯片?", 3),
        (r"(板级|片级|系统级)初始化", 4),
        (r"固件|实时任务|优先级反转", 3),
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
    "软件工程": [
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
    "绪论": [
        (r"系统分析师(的)?(角色|职责|定位)", 4),
        (r"(信息系统)?(生存|生命)周期", 3),
        (r"职业道德|职业规范|行为准则", 4),
        (r"(需求获取|原型法|焦点小组|访谈|问卷).{0,15}(方法|阶段)?", 2),
        (r"软件危机|系统工程|基于模型的系统工程", 2),
    ],
    "系统规划与分析": [
        (r"(系统)?可行性(研究|分析|报告)", 4),
        (r"(技术|经济|社会|运行|操作)可行性", 4),
        (r"(总体)?(信息)?(化)?(战略|规划)", 2),
        (r"业务流程(分析|图|建模|重组|再造|优化)", 3),
        (r"(系统|数据)(流图|DFD|字典)", 3),
        (r"(系统分析|说明书|分析报告)", 2),
    ],
    "软件需求工程": [
        (r"(软件|系统)?需求(规格|说明)(书|SRS)?", 3),
        (r"需求(获取|建模|分析|验证|评审|确认|变更|管理|跟踪)", 3),
        (r"(联合需求计划|JRP|焦点小组|原型法|访谈|问卷)", 3),
        (r"(功能|非功能|业务|用户)需求", 2),
        (r"(用例|活动|状态)图.{0,6}(建模|分析)?", 2),
    ],
    "软件架构设计": [
        (r"(软件|系统)架构(设计|风格|评估)?", 3),
        (r"(分层|管道-?过滤器|微内核|事件驱动|客户端-?服务器|C/S|B/S|MVC|MVP|MVVM)", 3),
        (r"4\+1视图|4+1 视图|架构视图|ADL|架构描述", 4),
        (r"(ATAM|SAAM|CBAM)(方法|评估)?", 4),
        (r"(质量属性|性能|可用性|可维护性|可扩展性|安全性).{0,4}(权衡|架构)?", 2),
        (r"(EJB|Servlet|Spring( Boot| Cloud)?|J2EE|\.NET)", 2),
    ],
    "系统设计": [
        (r"(概要|详细)设计(说明书|报告|评审)?", 4),
        (r"(模块|接口|数据)(结构)?设计", 3),
        (r"(程序流程图|N-S图|N-S 图|PAD图|PAD 图|PDL|伪代码)", 4),
        (r"(耦合|内聚|模块独立性|信息隐藏|结构化设计)", 3),
    ],
    "软件实现与测试": [
        (r"(编码|编程)(规范|风格|审查)?", 3),
        (r"(持续集成|持续交付|灰度发布|CI/?CD)", 4),
        (r"(蛮力法|回溯法|原因排除法|演绎法|归纳法).{0,4}调试", 4),
        (r"(测试用例|测试计划|测试报告|缺陷(管理|跟踪)|回归测试)", 3),
        (r"(覆盖率|缺陷密度|MTTF|MTBF)", 3),
    ],
    "系统运行与维护": [
        (r"(系统|软件)运行(管理|制度|记录|值班)", 3),
        (r"(ITIL|ITSS|事件管理|问题管理|变更管理).{0,6}(流程|实践)?", 3),
        (r"(改正性|适应性|完善性|预防性)维护", 4),
        (r"(维护申请|维护分析|维护实施|维护评审)", 3),
        (r"(系统|性能|经济)评价|用户满意度|技术债务|再工程", 2),
    ],
    "Web 应用系统": [
        (r"(RESTful|REST|GraphQL|前后端分离)", 4),
        (r"(Vue|React|Angular|Node\.js|Spring Boot|Django|Express)", 3),
        (r"(PWA|Service Worker|渐进式|Web 应用)", 4),
        (r"(CDN|懒加载|Web 性能|Web 缓存|浏览器缓存)", 3),
        (r"(Web 安全|WAF|SQL 注入|XSS|CSRF|会话劫持)", 3),
    ],
    "移动应用系统": [
        (r"(Android|iOS|Flutter|React Native|移动(端|应用|APP))", 4),
        (r"(Activity|Service|BroadcastReceiver|ContentProvider|Cocoa)", 4),
        (r"(原生|混合|跨平台)(开发|应用)", 3),
        (r"(WebSocket|推送|心跳|移动网络)", 3),
        (r"(Material Design|响应式|移动 UX|无障碍)", 2),
    ],
    "大数据处理系统": [
        (r"(大数据|Hadoop|HDFS|MapReduce|YARN|Hive|HBase|Sqoop|Flume)", 4),
        (r"(Spark|Flink|Storm|流处理|批处理|Spark SQL|Spark Streaming|Kafka Streams)", 4),
        (r"(数据湖|数据中台|湖仓一体|元数据管理)", 4),
        (r"(数据挖掘|分类|聚类|关联规则|协同过滤|机器学习)", 3),
    ],
    "微服务系统": [
        (r"微服务(架构|拆分|治理)?", 4),
        (r"(服务(注册|发现|治理|网关|熔断|降级|限流))", 4),
        (r"(Eureka|Consul|Nacos|ZooKeeper|Spring Cloud)", 4),
        (r"(Service Mesh|Istio|Linkerd|Envoy)", 4),
        (r"(Docker|Kubernetes|容器编排|云原生)", 3),
        (r"(Saga|TCC|两阶段提交|最大努力通知|Seata|分布式事务)", 3),
    ],
    "信息物理系统": [
        (r"(信息物理系统|CPS|赛博物理系统)", 4),
        (r"(工业(互联网|4\.0|云|PaaS|APP)|智能制造|数字孪生|智能工厂)", 4),
        (r"(OPC UA|MQTT|CoAP|时间敏感网络|TSN)", 4),
        (r"(边缘计算|工业 APP|柔性制造|感知层|网络层)", 3),
    ],
    "系统分析师论文写作要点": [
        (r"(论文摘要|摘要.{0,6}(正文|论题|论据))", 4),
        (r"(论文写作|论题|论点|论据|经验总结)", 4),
        (r"(项目背景|技术难点|解决方案|效果评价).{0,8}论文", 3),
    ],
}

EN_RULES = {
    # 英语题按英文考点词归类
    "软件工程": [(r"UML|object[- ]oriented|use ?case|polymorphi|inheritan", 3),
                 (r"requirement|systems? analysis|software (design|testing|maintenance|process)|waterfall|SDLC", 3)],
    "计算机网络与分布式系统": [(r"network|protocol|router|TCP|firewall|distributed|microservice", 3)],
    "数据库系统": [(r"database|SQL|normaliz", 3)],
    "Web 应用系统": [(r"web|HTML|CSS|JavaScript|frontend|backend|API", 3)],
    "微服务系统": [(r"microservice|service mesh|container|orchestrat", 3)],
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
            best_name, best_score = score, best_name if best_name is None else best_name
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
