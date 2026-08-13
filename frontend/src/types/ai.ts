export interface AiApiConfig {
  provider: string
  api_key: string
  base_url: string
  model: string
}

export interface AiProvider {
  provider: string
  name: string
  base_url: string
  models: string[]
}

export interface GenerateQuestionsReq {
  api_config: AiApiConfig
  subject_id: number
  chapter_id: number
  types: string[]
  difficulty: string
  count: number
}

export interface AiGeneratedQuestion {
  id: number
  type: string
  difficulty: string
  content: string
  // 案例分析题专用：800~1500 字背景材料
  case_material?: string
  options: string
  blank_options?: string
  answer: string
  analysis: string
  knowledge_point: string
  source?: string
}

export interface GenerateQuestionsResp {
  questions: AiGeneratedQuestion[]
}

export interface StartAiExamReq {
  subject_id: number
  count: number
  duration: number
}

// AI 解析题目:对单题做整体解析(知识点/正确答案分析/巩固建议)
// 当 ExtractKnowledgeModal 提取不到有效知识点时,用户可点击"AI 分析"走此接口。
export interface AnalyzeQuestionReq {
  api_config: AiApiConfig
  question_content: string
  question_type: string
  question_answer: string
  // 兜底占位:用户尚未作答时,后端 prompt 仍可正常返回(忽略对比部分)
  user_answer: string
}

export interface AnalyzeQuestionResp {
  analysis: string
}

// 论文 AI 评分
export interface EssayScoreReq {
  api_config: AiApiConfig
  question_id: number
  user_answer: string
  record_type: string  // practice / exam
  record_id: number    // 练习记录ID或考试记录ID
  exam_answer_id?: number  // 考试答题详情ID（仅考试模式）
}

export interface EssayScoreResp {
  id: number
  question_id: number
  question_type?: string
  total_score: number
  argument_score: number
  structure_score: number
  language_score: number
  depth_score: number
  comment: string
  created_at: string
}

export interface EssayScoreCheckResp {
  has_score: boolean
  score?: EssayScoreResp
}

// AI 异步出题任务
// Status: pending / running / success / failed
export interface AsyncGenerateTask {
  id: string
  status: 'pending' | 'running' | 'success' | 'failed' | string
  question_type: string
  chapter_id: number
  chapter_name: string
  difficulty: string
  count: number
  created_at: string
  updated_at: string
  finished_at?: string
  error?: string
  result?: GenerateQuestionsResp
}

export interface AsyncSubmitResp {
  task_id: string
  status: string
}

// AI 生成题历史记录（一次生成 = 一个批次）
// status 字段由后端在 ListGenerateHistory 中并入进行中任务：
//   - success（默认，缺失视为成功）
//   - pending（已提交，排队中）
//   - running（生成中）
export interface AiBatchHistoryItem {
  batch_id: string
  subject_id: number
  subject_name: string
  chapter_id: number
  chapter_name: string
  type: string
  type_label: string
  difficulty: string
  count: number
  created_at: string
  answered_count: number
  correct_count: number
  knowledge_points?: string[]
  status?: 'pending' | 'running' | 'success' | 'failed' | string
}

export interface AiBatchHistoryResp {
  list: AiBatchHistoryItem[]
  total: number
}
