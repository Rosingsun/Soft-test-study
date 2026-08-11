import { get, post, del } from './request'
import type { WrongQuestionResp, WrongPracticeSubmitReq, PracticeRecordResp } from '@/types/question'

export function listWrongQuestions(subjectId?: number, source?: string, sort?: string) {
  const params = new URLSearchParams()
  if (subjectId) params.set('subject_id', String(subjectId))
  if (source && source !== 'all') params.set('source', source)
  if (sort) params.set('sort', sort)
  const query = params.toString()
  return get<WrongQuestionResp[]>(`/wrong-questions${query ? `?${query}` : ''}`)
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
