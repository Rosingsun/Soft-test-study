import { get, post } from './request'
import type { PracticeSubmitReq, PracticeRecordResp, Question } from '@/types/question'

export function submitPractice(data: PracticeSubmitReq) {
  return post<PracticeRecordResp>('/practice/submit', data)
}

export function getRandomQuestions(subjectId: number, count = 10, difficulty?: string) {
  return get<Question[]>('/questions/random', {
    subject_id: subjectId,
    count,
    ...(difficulty ? { difficulty } : {}),
  })
}

export function getSpecialQuestions(subjectId: number, type?: string, difficulty?: string, count = 10) {
  return get<Question[]>('/questions/special', {
    subject_id: subjectId,
    count,
    ...(type ? { type } : {}),
    ...(difficulty ? { difficulty } : {}),
  })
}
