package service

// AI 出题 Prompt 模板（5 套对应 5 种题型）+ Few-shot 示例。
// 核心目标：让 LLM 严格按《系统分析师教程（第 2 版）》考点命制题目，
// 并按统一格式输出 JSON。
//
// 所有模板统一头部（commonPromptHead）注入：角色、考试结构、本次范围、考点清单、
// 难度定义、分析三段式、Few-shot、输出格式约束。

// commonPromptHead 5 套模板共用的头部。
// 注入：角色、考试结构、本次范围（科目+章节+难度+数量）、考点清单、
// 难度定义、分析三段式、Few-shot、输出格式约束。
//
// 参数：
//
//	subject    - 科目名（用于头部「本次范围」展示）
//	chapter    - 章节名（"不限定"表示全章节随机）
//	kpBlock    - 考点清单（formatChapterKPBlock / formatChapterOutlineBlock 格式化后）
//	difficulty - 难度描述（简单 / 中等 / 困难）
//	count      - 数量
func commonPromptHead(subject, chapter, kpBlock, difficulty, count string) string {
	// 章节命中时强调"必须从该章节出题"；不限定时强调"从考点清单中任选"
	chapterRule := "请从下方「考点清单」中任选 1~3 个考点作为本题核心考查点。"
	if chapter != "" && chapter != "不限定" {
		chapterRule = "你必须从「" + chapter + "」这一章的考点中出题，不允许跨章节串题。"
	}
	return `【角色】你是软考《` + subject + `》命题专家，深度熟悉该科目官方考试大纲与全部考点。
【考试结构】上午题=综合知识（75 道选择/判断题，1 分/题）；下午题 I=案例分析（5 道大题，15 分/题）；下午题 II=论文（4~5 选 1，75 分）。
【本次范围】
- 科目：` + subject + `
- 章节：` + chapter + `
- 难度：` + difficulty + `（easy=识记级/概念直接复现；medium=对比与应用/原理推导；hard=综合场景/多步推理/简单计算）
- 数量：` + count + ` 道

【科目硬约束】你出的每一道题都必须落在《` + subject + `》官方考纲范围内。**严禁**出其他科目（如系统分析师、软件设计师、网络工程师等）的题。如果你发现按当前范围无题可出，直接返回空数组 []，不要硬出无关题。

【章节与考点】` + chapterRule + `不允许自创不在清单中的考点。
` + kpBlock + `

【分析与辨析要求】每道题目的 analysis 字段必须包含三段：
- 【考点】说明本题考查的具体细分知识点（必须能在上方考点清单中找到对应条目）。
- 【解析】对正确选项的原理说明，至少 2 句。
- 【辨析】对每个干扰项单独说明为什么错、易混淆点在哪。

【输出格式】严格遵守：
- 只输出一个 JSON 数组，禁止任何其他文字、解释、问候语。
- 禁止使用 markdown 代码块（禁止以 ` + "`" + `` + "`" + `` + "`" + ` 开头或结尾）。
- 禁止截断、禁止用 "..." 省略内容。
- 数组中所有字符串值必须使用合法 JSON 转义（双引号、反斜杠、换行等）。
- 生成 ` + count + ` 道题时，请在一条回复中一次性输出完整的 ` + count + ` 个数组元素，不要分批输出。

【Few-shot 参考示例】以下示例展示了一道标准题目的格式，请严格按此结构输出（仅作格式参考，不要抄题）：
`
}

// fewShotSingle 单选题示例。
const fewShotSingle = `[
  {
    "type": "single",
    "difficulty": "medium",
    "content": "某 CPU 流水线分 4 段 IF/ID/EX/WB，每段耗时 1ns。无冒险情况下执行 100 条指令，总耗时约为（  ）。",
    "options": [
      "100ns",
      "103ns",
      "400ns",
      "401ns"
    ],
    "answer": "B",
    "analysis": "【考点】指令流水线执行时间计算。\n【解析】流水线首条指令建立时间为 (4-1)*1ns = 3ns，之后每条指令 1ns，总时间 = 3 + 100*1 = 103ns。\n【辨析】A 错：未考虑流水线建立时间；C 错：误把每段串行 4ns；D 错：4+100 错算。",
    "knowledge_point": "指令流水线与吞吐率"
  }
]`

// fewShotMulti 多选题示例。
const fewShotMulti = `[
  {
    "type": "multi",
    "difficulty": "medium",
    "content": "下列访问控制模型中，属于强制访问控制与自主访问控制相结合的有（  ）。",
    "options": [
      "ACL 访问控制列表",
      "RBAC 基于角色的访问控制",
      "MAC 强制访问控制",
      "DAC 自主访问控制"
    ],
    "answer": "A,C,D",
    "analysis": "【考点】访问控制模型分类。\n【解析】ACL 属于自主访问控制（DAC）实现，MAC 属于强制访问控制，RBAC 既可基于 DAC 也可基于 MAC 实施。\n【辨析】B 错：RBAC 既不是纯 DAC 也不是纯 MAC；其余均为典型访问控制模型，题目问"强制+自主"组合时 A/C/D 均符合。",
    "knowledge_point": "访问控制模型"
  }
]`

// fewShotJudge 判断题示例。
const fewShotJudge = `[
  {
    "type": "judge",
    "difficulty": "easy",
    "content": "在 M/M/1 排队系统中，服务强度 ρ 越大，平均队列长度越短。（  ）",
    "options": [
      "正确",
      "错误"
    ],
    "answer": "错误",
    "analysis": "【考点】M/M/1 排队论模型。\n【解析】M/M/1 的平均队列长度 Lq = ρ²/(1-ρ)，随 ρ 单调递增，ρ 越大 Lq 越长（系统接近饱和时急剧上升）。\n【辨析】易错点：服务强度增大意味着到达率接近服务率，队列堆积而非缩短。",
    "knowledge_point": "排队论 M/M/1 模型"
  }
]`

// fewShotCaseStudy 案例分析题示例。
// 案例题示例展示完整结构：content 字段为子问题列表，
// 实际写入数据库时 case_material 字段单独存储背景材料。
const fewShotCaseStudy = `[
  {
    "type": "case_study",
    "difficulty": "hard",
    "content": "案例背景见 case_material 字段。\n\n请根据背景回答以下问题：\n（1）请简述该企业实施 ERP 项目的可行性和必要性（5 分）。\n（2）下列需求获取方法中，最适合 ERP 调研阶段的有（  ）（多选，2 分）。\n（3）该企业将订单服务与库存服务拆分时，应采用何种分布式事务方案？请说明理由（8 分）。",
    "options": [],
    "answer": "（1）答：可行性-企业已具备信息化基础、领导支持；必要性-解决信息孤岛、提升运营效率（写出任 2 个核心要点即可得分）。\n（2）A,B,D\n（3）采用最终一致 + 消息队列异步方案，理由是订单与库存属于可异步场景，强一致会显著降低吞吐；具体可用本地消息表 + RocketMQ 实现对账。",
    "analysis": "【考点】企业信息化 ERP 实施 + 软件工程需求获取 + 分布式系统事务。\n【解析】(1) ERP 立项需从战略、业务、技术三方面论述可行性与必要性。(2) 常用需求获取方法有访谈、问卷、原型法、观察、焦点小组；纯实验不属需求获取方法。(3) 分布式系统优先选最终一致而非强一致，需结合业务容忍度。\n【辨析】(2) 干扰点：C 实验室对照实验属于研究方法而非需求获取；E 假设-演绎属于推理方法。",
    "knowledge_point": "ERP 实施与分布式事务"
  }
]`

// fewShotEssay 论文题示例。
const fewShotEssay = `[
  {
    "type": "essay",
    "difficulty": "hard",
    "content": "论软件需求变更管理在大型信息系统项目中的实践\n\n要求：\n1. 概要叙述你参与管理和开发的软件项目，以及你在其中承担的工作。\n2. 结合项目实际，详细论述需求变更管理的主要过程、所用工具与方法。\n3. 针对需求变更对项目进度、成本、质量的影响，提出具体的应对措施。",
    "options": [],
    "answer": "",
    "analysis": "",
    "knowledge_point": "软件需求工程与需求变更管理"
  }
]`

// outputFormatSingle 单选题 JSON 格式约束。
const outputFormatSingle = `请严格按照以下 JSON 格式输出（每道题 4 个选项，answer 为单个字母）：
[
  {
    "type": "single",
    "difficulty": "<easy|medium|hard>",
    "content": "题干（支持 Markdown）",
    "options": ["选项 1", "选项 2", "选项 3", "选项 4"],
    "answer": "A",
    "analysis": "【考点】...【解析】...【辨析】...",
    "knowledge_point": "细分考点名（必须与考点清单中一致）"
  }
]`

// outputFormatMulti 多选题 JSON 格式约束。
const outputFormatMulti = `请严格按照以下 JSON 格式输出（每道题 4 个选项，answer 为多个字母用逗号分隔，如 "A,C"）：
[
  {
    "type": "multi",
    "difficulty": "<easy|medium|hard>",
    "content": "题干（支持 Markdown）",
    "options": ["选项 1", "选项 2", "选项 3", "选项 4"],
    "answer": "A,C",
    "analysis": "【考点】...【解析】...【辨析】...",
    "knowledge_point": "细分考点名"
  }
]`

// outputFormatJudge 判断题 JSON 格式约束。
const outputFormatJudge = `请严格按照以下 JSON 格式输出（2 个选项固定为「正确」「错误」，answer 为 "正确" 或 "错误"）：
[
  {
    "type": "judge",
    "difficulty": "<easy|medium|hard>",
    "content": "题干为一段陈述句（支持 Markdown）",
    "options": ["正确", "错误"],
    "answer": "正确",
    "analysis": "【考点】...【解析】...【辨析】...",
    "knowledge_point": "细分考点名"
  }
]`

// outputFormatCaseStudy 案例分析题 JSON 格式约束。
// 案例题 prompt 比客观题多一段「案例背景要求」。
const outputFormatCaseStudy = `【案例背景要求】必须先输出一段 800~1500 字的真实业务场景背景材料，作为本题的情境引入。背景应：
- 贴合本章核心考点（如软件工程章节就以软件项目为背景，信息安全章节就以信息系统安全事件为背景）。
- 给出具体业务背景、关键参与者、已实施/拟实施系统的技术栈、面临的关键问题。
- 文风严谨，贴近系统分析师考试下午题 I 案例题的真实风格。

【子问题要求】基于上述背景，命制 3~5 个子问题，混合以下类型：
- 选择 / 判断（每道 1~2 分）
- 填空（每空 1 分）
- 简答（每道 3~8 分，需用文字作答）
- 论述（每道 8~15 分）

请严格按照以下 JSON 格式输出（options 为空数组，answer 用 \\n 分隔各子问题答案）：
[
  {
    "type": "case_study",
    "difficulty": "hard",
    "content": "（此处为子问题列表，每行一个子问题，编号 (1)(2)(3) ...）",
    "case_material": "（此处为 800~1500 字案例背景材料）",
    "options": [],
    "answer": "（1）...\\n（2）A,B\\n（3）...",
    "analysis": "【考点】...\\n【解析】...\\n【辨析】...",
    "knowledge_point": "细分考点名（复合型）"
  }
]`

// outputFormatEssay 论文题 JSON 格式约束。
const outputFormatEssay = `【论文命制要求】本任务为系统分析师下午题 II 论文的命制，结构如下：
- 给出 2~3 个论题方向供 LLM 选其一命制一道完整论文题。
- 最终在 content 字段输出论文题目正文（含"论 ……"主题句 + 3 段写作要求）。
- 论文正文应 200~350 字，包含：背景（项目概述）→ 主题论点（围绕考点清单中的某 1 个细分知识点）→ 3 条写作要求。
- 不需要给出答案或解析（答题是用户的事，AI 评分由 EssayScore 流程独立处理）。

请严格按照以下 JSON 格式输出（options 为空数组，answer / analysis 留空字符串）：
[
  {
    "type": "essay",
    "difficulty": "hard",
    "content": "论 …（论文题正文 200~350 字，含项目背景 + 主题句 + 3 条写作要求）",
    "options": [],
    "answer": "",
    "analysis": "",
    "knowledge_point": "细分考点名"
  }
]`

// extractKnowledgePointsPrompt 从一道题目中识别其考查的核心考点（仅 1 个）的 prompt。
// 输出严格 JSON 数组（仅含 1 个元素），name（<=30字名词短语）+ description（<=150字，Markdown）。
const extractKnowledgePointsPrompt = `【角色】你是软考辅导专家，擅长识别题目所考查的考点。

【任务】根据给定的题目信息，识别这道题在考查的**核心考点**（仅 1 个）。该知识点会被用户收录到「我的知识点」中作为长期复习资料。

【要求】
- 只依据「题干」判断考点，识别题目本身在考查什么。
- name：30 字以内的名词短语，使用软考官方考纲用语，例如「指令流水线与吞吐率」「访问控制模型」。
- description：150 字以内，聚焦该考点的"定义"与"关键要点"，可使用 Markdown 加粗、列表等简单格式。
- **严禁**从选项内容、正确答案、解析中提取知识点。
- **严禁**输出「选项 A 的含义」「B 为什么错」这类拆解选项的子概念。
- 严禁拆解题干中的具体案例/情景，只提取其背后对应的考纲考点。
- 不要直接复述题目本身。

【输出格式】严格遵守：
- 只输出一个 JSON 数组（仅 1 个元素），禁止任何其他文字、解释、问候语。
- 禁止使用 markdown 代码块。
- 禁止截断、禁止用 "..." 省略内容。
- 数组中所有字符串值必须使用合法 JSON 转义。

【Few-shot 参考示例】
[
  {
    "name": "访问控制模型",
    "description": "**DAC**（自主）：资源所有者决定访问权限，如 ACL。\n\n**MAC**（强制）：依据主体与客体的安全级别决定访问。\n\n**RBAC**（角色）：通过角色间接授权用户权限。"
  }
]`
