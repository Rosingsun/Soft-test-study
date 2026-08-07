import { get, post, del } from './request'

export function addMark(questionId: number) {
  return post<{ marked: boolean }>(`/questions/${questionId}/mark`)
}

export function removeMark(questionId: number) {
  return del<{ marked: boolean }>(`/questions/${questionId}/mark`)
}

export function checkMarked(questionId: number) {
  return get<{ marked: boolean }>(`/questions/${questionId}/marked`)
}
