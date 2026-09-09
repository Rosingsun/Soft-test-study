import { del, get, patch, post, put } from './request'
import type { PageData } from '@/types/common'
import type {
  AdminBatchQuestionStatusReq,
  AdminQuestionResp,
  AdminQuestionStatsResp,
  AdminQuestionUpsertReq,
  AdminSubjectResp,
  AdminSubjectUpsertReq,
  AdminUserDetailResp,
  AdminUserListItem,
  CreateAdminInvitationCodeReq,
  CreateAdminUserReq,
  InvitationCodeResp,
  ListAdminInvitationCodesParams,
  ListAdminQuestionsParams,
  ListAdminSubjectsParams,
  ListAdminUsersParams,
  UpdateAdminUserReq,
} from '@/types/admin'

/** 分页查询邀请码（仅管理员） */
export function listAdminInvitationCodes(params: ListAdminInvitationCodesParams = {}) {
  const query: Record<string, string | number> = {}
  if (params.state) query.state = params.state
  if (params.keyword) query.keyword = params.keyword
  if (params.created_by) query.created_by = params.created_by
  query.page = params.page ?? 1
  query.page_size = params.page_size ?? 20
  return get<PageData<InvitationCodeResp>>('/admin/invitation-codes', query)
}

/** 生成邀请码（仅管理员），返回新生成的码列表 */
export function createAdminInvitationCode(body: CreateAdminInvitationCodeReq) {
  return post<InvitationCodeResp[]>('/admin/invitation-codes', body)
}

/** 查询任意用户的详情档案（仅管理员） */
export function getAdminUserDetail(userId: number) {
  return get<AdminUserDetailResp>(`/admin/users/${userId}`)
}

/** 管理端科目列表（含停用科目，可按键名 / 等级筛选，仅管理员） */
export function listAdminSubjects(params: ListAdminSubjectsParams = {}) {
  const query: Record<string, string | number> = {}
  if (params.level_id) query.level_id = params.level_id
  if (params.keyword) query.keyword = params.keyword
  return get<AdminSubjectResp[]>('/admin/subjects', query)
}

/** 新增科目（仅管理员） */
export function createAdminSubject(body: AdminSubjectUpsertReq) {
  return post<AdminSubjectResp>('/admin/subjects', body)
}

/** 更新科目（仅管理员） */
export function updateAdminSubject(id: number, body: AdminSubjectUpsertReq) {
  return put<AdminSubjectResp>(`/admin/subjects/${id}`, body)
}

/** 启用 / 停用科目（仅管理员） */
export function setAdminSubjectStatus(id: number, status: 0 | 1) {
  return patch<AdminSubjectResp>(`/admin/subjects/${id}/status`, { status })
}

/** 删除科目（仅管理员） */
export function deleteAdminSubject(id: number) {
  return del<AdminSubjectResp>(`/admin/subjects/${id}`)
}

// ============ 题库管理 ============

/** 分页查询题目（支持 科目/子科目/章节/题型/难度/状态/年份/关键词 筛选） */
export function listAdminQuestions(params: ListAdminQuestionsParams = {}) {
  const query: Record<string, string | number> = {}
  if (params.subject_id) query.subject_id = params.subject_id
  if (params.sub_subject_id) query.sub_subject_id = params.sub_subject_id
  if (params.chapter_id) query.chapter_id = params.chapter_id
  if (params.type) query.type = params.type
  if (params.difficulty) query.difficulty = params.difficulty
  // status 为 '0' 是合法筛选值，需单独判断避免被 truthy 短路吞掉
  if (params.status !== undefined && params.status !== '') query.status = params.status
  if (params.year) query.year = params.year
  if (params.keyword) query.keyword = params.keyword
  query.page = params.page ?? 1
  query.page_size = params.page_size ?? 20
  return get<PageData<AdminQuestionResp>>('/admin/questions', query)
}

/** 查询题目详情（案例分析大题目会带回小题目） */
export function getAdminQuestion(id: number) {
  return get<AdminQuestionResp>(`/admin/questions/${id}`)
}

/** 新增题目 */
export function createAdminQuestion(body: AdminQuestionUpsertReq) {
  return post<AdminQuestionResp>('/admin/questions', body)
}

/** 更新题目 */
export function updateAdminQuestion(id: number, body: AdminQuestionUpsertReq) {
  return put<AdminQuestionResp>(`/admin/questions/${id}`, body)
}

/** 删除题目（软删除：置为停用） */
export function deleteAdminQuestion(id: number) {
  return del<{ id: number }>(`/admin/questions/${id}`)
}

/** 批量启用 / 停用题目 */
export function batchSetAdminQuestionStatus(body: AdminBatchQuestionStatusReq) {
  return patch<{ updated: number }>('/admin/questions/status', body)
}

/** 题库概览：总量 / 启用 / 停用 / 各题型分布 */
export function getAdminQuestionStats() {
  return get<AdminQuestionStatsResp>('/admin/stats/questions')
}

// ============ 用户管理 ============

/** 分页查询用户（关键词 / 角色 / 状态 筛选） */
export function listAdminUsers(params: ListAdminUsersParams = {}) {
  const query: Record<string, string | number> = {}
  if (params.keyword) query.keyword = params.keyword
  if (params.role) query.role = params.role
  if (params.status !== undefined) query.status = params.status
  query.page = params.page ?? 1
  query.page_size = params.page_size ?? 20
  return get<PageData<AdminUserListItem>>('/admin/users', query)
}

/** 管理员直接新建用户 */
export function createAdminUser(body: CreateAdminUserReq) {
  return post<AdminUserListItem>('/admin/users', body)
}

/** 管理员编辑用户资料 */
export function updateAdminUser(id: number, body: UpdateAdminUserReq) {
  return put<AdminUserListItem>(`/admin/users/${id}`, body)
}

/** 启用 / 禁用用户 */
export function setAdminUserStatus(id: number, status: number) {
  return patch<AdminUserListItem>(`/admin/users/${id}/status`, { status })
}

/** 管理员重置用户密码 */
export function resetAdminUserPassword(id: number, password: string) {
  return post<{ ok: boolean }>(`/admin/users/${id}/reset-password`, { password })
}

/** 删除用户 */
export function deleteAdminUser(id: number) {
  return del<{ ok: boolean }>(`/admin/users/${id}`)
}
