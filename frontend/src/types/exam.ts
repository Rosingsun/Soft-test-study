export interface ExamTemplateResp {
  id: number
  subject_id: number
  name: string
  duration: number
  total_score: number
  year: number
  question_type: string
  question_count: number
}

export interface ExamQuesResp {
  id: number
  type: string
  content: string
  case_material?: string
  options: string
  blank_options?: string
  difficulty: string
  score: number
  sort_order: number
}

export interface StartExamResp {
  record_id: number
  template_id: number
  template_name: string
  duration: number
  total_score: number
  questions: ExamQuesResp[]
  answers: Record<number, string>
  elapsed: number
  status: string
}

export interface SubmitAnswerReq {
  question_id: number
  answer: string
}

export interface AnswerDetailResp {
  question_id: number
  exam_answer_id: number
  type: string
  content: string
  options: string
  blank_options?: string
  your_answer: string
  correct_answer: string
  is_correct: number
  score: number
  analysis: string
}

export interface SectionAccuracyResp {
  type: string
  total_count: number
  correct_count: number
  accuracy: number
}

export interface ExamResultResp {
  record_id: number
  template_name: string
  subject_id: number
  score: number
  total_score: number
  duration: number      // 秒
  correct_count: number
  total_count: number
  accuracy: number      // 本次考试整体正确率 %
  status: string
  started_at: string
  finished_at: string
  section_accuracy: SectionAccuracyResp[]
  answer_details: AnswerDetailResp[]
}

export interface ExamRecordResp {
  id: number
  template_id: number
  template_name: string
  subject_id: number
  score: number
  total_score: number
  duration: number
  correct_count: number
  accuracy: number
  status: string
  started_at: string
  finished_at: string
  created_at: string
}
