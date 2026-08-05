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
  options: string
  answer: string
  analysis: string
  knowledge_point: string
}

export interface GenerateQuestionsResp {
  questions: AiGeneratedQuestion[]
}

export interface AnalyzeReq {
  api_config: AiApiConfig
  question_content: string
  question_type: string
  question_answer: string
  user_answer: string
}

export interface AnalyzeResp {
  analysis: string
}

export interface StartAiExamReq {
  subject_id: number
  count: number
  duration: number
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
