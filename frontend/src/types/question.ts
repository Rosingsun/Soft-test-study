export type QuestionType = 'single' | 'multi' | 'judge' | 'fill' | 'short' | 'comprehensive' | 'essay' | 'case_study'
export type Difficulty = 'easy' | 'medium' | 'hard'

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
  answer: string
  analysis: string
  year: number
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
}

export interface WrongPracticeSubmitReq {
  question_id: number
  answer: string
  duration: number
}
