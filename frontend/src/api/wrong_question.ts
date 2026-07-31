import { get, post, del } from './request'
import type { WrongQuestionResp, WrongPracticeSubmitReq, PracticeRecordResp } from '@/types/question'

export function listWrongQuestions(subjectId?: number) {
  const query = subjectId ? `?subject_id=${subjectId}` : ''
  return get<WrongQuestionResp[]>(`/wrong-questions${query}`)
}

export function countWrongQuestions() {
  return get<{ count: number }>('/wrong-questions/count')
}

export function wrongPracticeSubmit(data: WrongPracticeSubmitReq) {
  return post<PracticeRecordResp>('/wrong-questions/practice', data)
}

export function removeWrongQuestion(questionId: number) {
  return del<void>(`/wrong-questions/${questionId}`)
}
