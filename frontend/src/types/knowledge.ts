export interface KnowledgePointItem {
  name: string
  description: string
}

export interface ExtractKnowledgePointsReq {
  api_config: {
    provider: string
    api_key: string
    base_url: string
    model: string
  }
  question_content: string
  question_type: string
  /** 保留字段，后端不再读取 */
  question_answer?: string
  /** 保留字段，后端不再读取 */
  question_analysis?: string
  subject_id?: number
}

export interface ExtractKnowledgePointsResp {
  points: KnowledgePointItem[]
}

export interface AddKnowledgePointReq {
  name: string
  description?: string
  subject_id?: number
  source_question_id?: number
}

export interface ExistingKnowledgePoint {
  user_kp_id: number
  name: string
  description: string
  is_important: boolean
  is_mastered: boolean
  created_at: string
}

export interface KnowledgePointResp {
  id: number
  name: string
  description: string
  subject_id: number
  source: string
  created_at: string
  updated_at: string
}

export interface UserKnowledgePointResp {
  id: number
  user_id: number
  knowledge_point_id: number
  source_question_id: number
  is_important: boolean
  is_mastered: boolean
  mastered_at: string
  note: string
  created_at: string
  updated_at: string
  knowledge: KnowledgePointResp
}

export interface AddKnowledgePointItemResp {
  name: string
  status: string
  user_kp_id: number
  knowledge?: KnowledgePointResp
  existing?: ExistingKnowledgePoint
}

export interface BatchAddKnowledgePointReq {
  points: AddKnowledgePointReq[]
  source_question_id?: number
}

export interface BatchAddKnowledgePointResp {
  added: AddKnowledgePointItemResp[]
  duplicates: AddKnowledgePointItemResp[]
}

export type ResolveDuplicateAction = 'keep_old' | 'replace' | 'add_new'

export interface ResolveDuplicateReq {
  action: ResolveDuplicateAction
  new_name?: string
  new_description?: string
  subject_id?: number
}

export interface UpdateUserKnowledgePointReq {
  is_important?: boolean
  is_mastered?: boolean
}

export interface KnowledgePointStatsResp {
  total: number
  important: number
  mastered: number
  unmastered: number
  recent_week: number
}

export type KnowledgeFilter = 'all' | 'important' | 'mastered' | 'unmastered'

export interface ListKnowledgePointParams {
  filter?: KnowledgeFilter
  subject_id?: number
  keyword?: string
  page?: number
  page_size?: number
}
