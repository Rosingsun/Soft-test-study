import { get, post } from './request'
import type { AiProvider, GenerateQuestionsReq, GenerateQuestionsResp, StartAiExamReq, EssayScoreReq, EssayScoreResp, EssayScoreCheckResp } from '@/types/ai'
import type { StartExamResp } from '@/types/exam'

export function getAiProviders() {
  return get<{ providers: AiProvider[] }>('/ai/providers')
}

export function generateQuestions(data: GenerateQuestionsReq) {
  return post<GenerateQuestionsResp>('/ai/generate', data)
}

export function startAiExam(data: StartAiExamReq) {
  return post<StartExamResp>('/ai/exam/start', data)
}

export function essayScore(data: EssayScoreReq) {
  return post<EssayScoreResp>('/ai/essay-score', data)
}

export function checkEssayScore(params: { record_type: string; record_id?: number; exam_answer_id?: number }) {
  const query = new URLSearchParams()
  query.set('record_type', params.record_type)
  if (params.record_id) query.set('record_id', String(params.record_id))
  if (params.exam_answer_id) query.set('exam_answer_id', String(params.exam_answer_id))
  return get<EssayScoreCheckResp>(`/ai/essay-score/check?${query.toString()}`)
}
