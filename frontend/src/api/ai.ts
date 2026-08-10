import { get, post } from './request'
import type {
  AiProvider,
  GenerateQuestionsReq,
  GenerateQuestionsResp,
  StartAiExamReq,
  EssayScoreReq,
  EssayScoreResp,
  EssayScoreCheckResp,
  AsyncGenerateTask,
  AsyncSubmitResp,
} from '@/types/ai'
import type { StartExamResp } from '@/types/exam'

export function getAiProviders() {
  return get<{ providers: AiProvider[] }>('/ai/providers')
}

export function generateQuestions(data: GenerateQuestionsReq) {
  return post<GenerateQuestionsResp>('/ai/generate', data)
}

// 异步 AI 出题：立即返回 task_id，后台生成完成后通过通知告知
export function submitGenerateAsync(data: GenerateQuestionsReq) {
  return post<AsyncSubmitResp>('/ai/generate/async', data)
}

export function getGenerateTask(id: string) {
  return get<AsyncGenerateTask>(`/ai/tasks/${id}`)
}

export function listGenerateTasks() {
  return get<{ list: AsyncGenerateTask[] }>('/ai/tasks')
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
