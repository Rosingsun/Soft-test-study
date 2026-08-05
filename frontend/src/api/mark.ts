import { get, post, del } from './request'
import type { QuestionMarkResp } from '@/types/question'

export function addMark(questionId: number) {
  return post<{ marked: boolean }>(`/questions/${questionId}/mark`)
}

export function removeMark(questionId: number) {
  return del<{ marked: boolean }>(`/questions/${questionId}/mark`)
}

export function checkMarked(questionId: number) {
  return get<{ marked: boolean }>(`/questions/${questionId}/marked`)
}

export function listMarks() {
  return get<QuestionMarkResp[]>('/marks')
}
