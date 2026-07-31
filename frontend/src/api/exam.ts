import { get, post } from './request'
import type { ExamTemplateResp, StartExamResp, SubmitAnswerReq, ExamResultResp, ExamRecordResp } from '@/types/exam'

export function listExamTemplates() {
  return get<ExamTemplateResp[]>('/exam-templates')
}

export function startExam(templateId: number) {
  return post<StartExamResp>(`/exam-templates/${templateId}/start`)
}

export function submitAnswer(recordId: number, data: SubmitAnswerReq) {
  return post<void>(`/exam-records/${recordId}/submit-answer`, data)
}

export function submitExam(recordId: number) {
  return post<ExamResultResp>(`/exam-records/${recordId}/submit`)
}

export function getExamResult(recordId: number) {
  return get<ExamResultResp>(`/exam-records/${recordId}/result`)
}

export function listExamRecords() {
  return get<ExamRecordResp[]>('/exam-records')
}
