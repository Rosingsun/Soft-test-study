import { get, post, patch, del } from './request'
import type {
  ExtractKnowledgePointsReq,
  ExtractKnowledgePointsResp,
  AddKnowledgePointReq,
  BatchAddKnowledgePointReq,
  BatchAddKnowledgePointResp,
  ResolveDuplicateReq,
  UpdateUserKnowledgePointReq,
  KnowledgePointStatsResp,
  ListKnowledgePointParams,
  UserKnowledgePointResp,
  AddKnowledgePointItemResp,
  KnowledgePointItem,
} from '@/types/knowledge'

export async function extractKnowledgePoints(data: ExtractKnowledgePointsReq) {
  return post<ExtractKnowledgePointsResp>('/ai/knowledge-points/extract', data)
}

export async function addKnowledgePoint(data: AddKnowledgePointReq) {
  return post<{ status: string; user_kp_id: number; knowledge: { id: number; name: string } }>(
    '/knowledge-points',
    data,
  )
}

export async function batchAddKnowledgePoints(data: BatchAddKnowledgePointReq) {
  return post<BatchAddKnowledgePointResp>('/knowledge-points/batch', data)
}

export async function resolveDuplicate(id: number, data: ResolveDuplicateReq) {
  return post<AddKnowledgePointItemResp>(`/knowledge-points/${id}/resolve-duplicate`, data)
}

export async function listKnowledgePoints(params: ListKnowledgePointParams = {}) {
  const query: Record<string, string | number> = {}
  if (params.filter) query.filter = params.filter
  if (params.subject_id) query.subject_id = params.subject_id
  if (params.keyword) query.keyword = params.keyword
  query.page = params.page || 1
  query.page_size = params.page_size || 20
  return get<{ list: UserKnowledgePointResp[]; total: number; page: number; page_size: number }>(
    '/knowledge-points',
    query,
  )
}

export async function getKnowledgeStats() {
  return get<KnowledgePointStatsResp>('/knowledge-points/stats')
}

export async function updateKnowledgePoint(id: number, data: UpdateUserKnowledgePointReq) {
  return patch<{ id: number }>(`/knowledge-points/${id}`, data)
}

export async function removeKnowledgePoint(id: number) {
  return del<{ id: number }>(`/knowledge-points/${id}`)
}

export type { KnowledgePointItem }
