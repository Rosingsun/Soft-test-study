import { get, post, put, del } from './request'
import type { StudyPlanResp, StudyPlanCreateReq, StudyPlanUpdateReq } from '@/types/studyPlan'

export function listStudyPlans() {
  return get<StudyPlanResp[]>('/study-plans')
}

export function createStudyPlan(data: StudyPlanCreateReq) {
  return post<StudyPlanResp>('/study-plans', data)
}

export function updateStudyPlan(id: number, data: StudyPlanUpdateReq) {
  return put<StudyPlanResp>(`/study-plans/${id}`, data)
}

export function deleteStudyPlan(id: number) {
  return del<void>(`/study-plans/${id}`)
}
