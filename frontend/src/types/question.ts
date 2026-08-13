export type QuestionType = 'single' | 'multi' | 'judge' | 'fill' | 'short' | 'comprehensive' | 'essay' | 'case_study' | 'multi_blank'
export type Difficulty = 'easy' | 'medium' | 'hard'

// 多空题每空选项结构，与后端 blank_options JSON 格式一致
export interface BlankOption {
  blank_index: number
  options: QuestionOption[]
}

export interface QuestionOption {
  id: string
  content: string
}

export interface Question {
  id: number
  subject_id: number
  sub_subject_id: number
  chapter_id: number
  type: QuestionType
  difficulty: Difficulty
  content: string
  case_material?: string
  options: string
  // 多空题专用：每空独立选项 JSON 字符串
  blank_options?: string
  answer: string
  analysis: string
  year: number
  source: string
  // AI 出题场景下携带的考点信息;非 AI 题目无此字段。
  // PracticeRunner 在 hideExtractWhenHasKnowledge=true 时据此判断是否展示"获取知识点"入口。
  knowledge_point?: string
}

export interface PracticeSubmitReq {
  question_id: number
  mode: string
  answer: string
  duration: number
}

export interface PracticeRecordResp {
  id: number
  question_id: number
  mode: string
  answer: string
  is_correct: number
  duration: number
}

export interface FavoriteFolder {
  id: number
  name: string
}

export interface CreateFolderReq {
  name: string
}

export interface QuestionFavoriteResp {
  id: number
  question_id: number
  folder_id: number
  created_at: string
}

export interface QuestionMarkResp {
  id: number
  question_id: number
  created_at: string
}

export interface WrongQuestionResp {
  id: number
  question_id: number
  wrong_count: number
  correct_count: number
  last_wrong_at: string
  created_at: string
  content: string
  type: QuestionType
  subject_id: number
  source?: string
}

export interface WrongPracticeSubmitReq {
  question_id: number
  answer: string
  duration: number
}
