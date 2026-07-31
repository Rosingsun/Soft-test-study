import { get } from './request'
import type { ExamLevel, Subject, SubSubject, Chapter, KnowledgePoint } from '@/types/subject'

export function getExamLevels() {
  return get<ExamLevel[]>('/exam-levels')
}

export function getSubjectsByLevel(levelId: number) {
  return get<Subject[]>('/subjects', { level_id: levelId })
}

export function getSubject(id: number) {
  return get<Subject>(`/subjects/${id}`)
}

export function getSubSubjects(subjectId: number) {
  return get<SubSubject[]>(`/subjects/${subjectId}/sub-subjects`)
}

export function getChapters(subSubjectId: number) {
  return get<Chapter[]>(`/sub-subjects/${subSubjectId}/chapters`)
}

export function getKnowledgePoints(chapterId: number) {
  return get<KnowledgePoint[]>(`/chapters/${chapterId}/knowledge-points`)
}
