import { get } from './request'
import type { Question } from '@/types/question'

export function getQuestionsByChapter(chapterId: number, difficulty?: string) {
  const params: Record<string, string | number> = { chapter_id: chapterId }
  if (difficulty) params.difficulty = difficulty
  return get<Question[]>('/questions', params)
}

export function getQuestion(id: number) {
  return get<Question>(`/questions/${id}`)
}
