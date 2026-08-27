import { get, post } from './request'
import { getSessionItem } from '@/utils/storage'
import type {
  AiProvider,
  GenerateQuestionsReq,
  GenerateQuestionsResp,
  StartAiExamReq,
  AnalyzeQuestionReq,
  AnalyzeQuestionResp,
  EssayScoreReq,
  EssayScoreResp,
  EssayScoreCheckResp,
  AsyncGenerateTask,
  AsyncSubmitResp,
  AiBatchHistoryResp,
} from '@/types/ai'
import type { StartExamResp } from '@/types/exam'

/**
 * OPT-05: 每次请求前从 sessionStorage 实时取 AI Key，
 * 不再依赖 store state，避免 store 与存储脱节时仍发送过期的 key。
 * 约定请求体中 api_config.api_key 字段被覆盖为 sessionStorage 中的最新值。
 */
export function withLiveApiKey<T>(body: T): T {
  const liveKey = getSessionItem('ai_key') || ''
  const b = body as unknown as Record<string, unknown>
  if (b.api_config && typeof b.api_config === 'object') {
    const ac = { ...(b.api_config as Record<string, unknown>), api_key: liveKey }
    return { ...b, api_config: ac } as unknown as T
  }
  return body
}

export function getAiProviders() {
  return get<{ providers: AiProvider[] }>('/ai/providers')
}

export function generateQuestions(data: GenerateQuestionsReq) {
  return post<GenerateQuestionsResp>('/ai/generate', withLiveApiKey(data))
}

// 异步 AI 出题：立即返回 task_id，后台生成完成后通过通知告知
export function submitGenerateAsync(data: GenerateQuestionsReq) {
  return post<AsyncSubmitResp>('/ai/generate/async', withLiveApiKey(data))
}

export function getGenerateTask(id: string) {
  return get<AsyncGenerateTask>(`/ai/tasks/${id}`)
}

export function listGenerateTasks() {
  return get<{ list: AsyncGenerateTask[] }>('/ai/tasks')
}

// 列出当前用户所有进行中的 AI 出题任务（pending/running）。
// 与 /ai/history 独立：用于在历史列表中显示「生成中」条目；专用端点保证
// 在用户刷新页面后仍能从后端持久化存储里恢复出来。
export function listInflightTasks() {
  return get<AiBatchHistoryResp>('/ai/tasks/inflight')
}

// AI 生成题历史记录列表（按批次）
export function listGenerateHistory() {
  return get<AiBatchHistoryResp>('/ai/history')
}

// 获取某个历史批次生成的全部题目（用于重新答题）
export function getGenerateBatch(batchId: string) {
  return get<GenerateQuestionsResp>(`/ai/history/${batchId}`)
}

export function startAiExam(data: StartAiExamReq) {
  return post<StartExamResp>('/ai/exam/start', withLiveApiKey(data))
}

// AI 解析题目:对题目做整体解析,用于知识库提取失败时的兜底入口
export function analyzeQuestion(data: AnalyzeQuestionReq) {
  return post<AnalyzeQuestionResp>('/ai/analyze', withLiveApiKey(data))
}

export function essayScore(data: EssayScoreReq) {
  return post<EssayScoreResp>('/ai/essay-score', withLiveApiKey(data))
}

export function checkEssayScore(params: { record_type: string; record_id?: number; exam_answer_id?: number }) {
  const query = new URLSearchParams()
  query.set('record_type', params.record_type)
  if (params.record_id) query.set('record_id', String(params.record_id))
  if (params.exam_answer_id) query.set('exam_answer_id', String(params.exam_answer_id))
  return get<EssayScoreCheckResp>(`/ai/essay-score/check?${query.toString()}`)
}
