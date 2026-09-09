import type { ExamRecordResp } from '@/types/exam'
import type { Difficulty, QuestionType } from '@/types/question'
import type {
  ChapterProgressResp,
  DailyStatsResp,
  StatsOverviewResp,
  SubjectProgressResp,
} from '@/types/stats'

// =============================================================
// 管理端（/admin）相关类型
// 字段命名与后端 dto/admin.go 保持一致（snake_case）
// =============================================================

/** 邀请码状态：未使用 / 已使用 / 已停用 */
export type InvitationCodeState = 'unused' | 'used' | 'disabled'

export interface InvitationCodeResp {
  id: number
  code: string
  note: string
  status: number
  /** 由后端统一判定，前端不重复推演 */
  state: InvitationCodeState
  created_by: number
  created_by_name: string
  used_by: number | null
  used_by_name: string
  used_at: string
  created_at: string
}

export interface AdminUserProfile {
  id: number
  username: string
  nickname: string
  email: string
  email_verified: boolean
  avatar: string
  role: string
  status: number
  level_id: number
  level_name: string
  subject_id: number
  subject_name: string
  difficulty: string
  created_at: string
}

export interface AdminUserStudyResp {
  overview: StatsOverviewResp
  recent: DailyStatsResp[]
  subjects: SubjectProgressResp[]
  chapters: ChapterProgressResp[]
}

export interface AdminUserAnswerResp {
  total_exams: number
  avg_exam_score: number
  records: ExamRecordResp[]
}

export interface AdminUserDetailResp {
  user: AdminUserProfile
  /** 管理员白名单注册的用户不消耗邀请码，可能为 null */
  invite_code: InvitationCodeResp | null
  study: AdminUserStudyResp
  answer: AdminUserAnswerResp
}

export interface ListAdminInvitationCodesParams {
  state?: string
  keyword?: string
  created_by?: number
  page?: number
  page_size?: number
}

export interface CreateAdminInvitationCodeReq {
  count?: number
  note?: string
}

// =============================================================
// 科目管理
// =============================================================

export interface AdminSubjectResp {
  id: number
  level_id: number
  level_name: string
  name: string
  short_name: string
  description: string
  icon: string
  sort_order: number
  /** 1=启用 0=停用 */
  status: number
  sub_subject_count: number
  question_count: number
}

export interface ListAdminSubjectsParams {
  level_id?: number
  keyword?: string
}

export interface AdminSubjectUpsertReq {
  level_id: number
  name: string
  short_name?: string
  description?: string
  icon?: string
  sort_order?: number
  /** 省略或 1=启用，0=停用 */
  status?: number
}

// =============================================================
// 题库管理（/admin/questions）
// 字段命名与后端 dto/admin.go 保持一致
// =============================================================

/** 题目列表查询参数 */
export interface ListAdminQuestionsParams {
  subject_id?: number
  sub_subject_id?: number
  chapter_id?: number
  type?: string
  difficulty?: string
  /** 空串=全部 / '1'=启用 / '0'=停用 */
  status?: string
  year?: number
  keyword?: string
  page?: number
  page_size?: number
}

/** 管理端题目行（含科目/章节名称） */
export interface AdminQuestionResp {
  id: number
  subject_id: number
  subject_name: string
  sub_subject_id: number
  sub_subject_name: string
  chapter_id: number
  chapter_name: string
  parent_id: number
  type: QuestionType
  difficulty: Difficulty
  content: string
  case_material: string
  /** 选项 JSON 字符串（前端解析后编辑/渲染） */
  options: string
  blank_options: string
  answer: string
  analysis: string
  year: number
  source: string
  status: number
  created_at: string
  updated_at: string
  /** 案例分析大题目下的小题目（仅详情接口填充） */
  children?: AdminQuestionChildResp[]
}

/** 案例分析大题目下的小题目（详情用） */
export interface AdminQuestionChildResp {
  id: number
  type: QuestionType
  content: string
  answer: string
  analysis: string
}

/** 新增 / 更新题目请求体 */
export interface AdminQuestionUpsertReq {
  subject_id: number
  sub_subject_id?: number
  chapter_id?: number
  type: QuestionType
  difficulty: Difficulty
  content: string
  case_material?: string
  options?: string
  blank_options?: string
  answer: string
  analysis?: string
  year?: number
  source?: string
  /** 省略时默认启用（1）；显式传 0 表示停用 */
  status?: number
}

/** 批量启用 / 停用请求体 */
export interface AdminBatchQuestionStatusReq {
  ids: number[]
  /** 1=启用 / 0=停用 */
  status: number
}

/** 各题型题量统计 */
export interface AdminQuestionTypeStat {
  type: QuestionType
  count: number
}

/** 题库概览 */
export interface AdminQuestionStatsResp {
  total: number
  enabled: number
  disabled: number
  by_type: AdminQuestionTypeStat[]
}

// =============================================================
// 用户管理（/admin/users）
// 字段命名与后端 dto/admin.go 保持一致
// =============================================================

/** 管理端用户列表行 */
export interface AdminUserListItem {
  id: number
  username: string
  nickname: string
  email: string
  email_verified: boolean
  avatar: string
  role: string
  status: number
  level_id: number
  level_name: string
  subject_id: number
  subject_name: string
  created_at: string
}

/** 用户列表查询参数 */
export interface ListAdminUsersParams {
  keyword?: string
  role?: string
  status?: number
  page?: number
  page_size?: number
}

/** 管理员新建用户请求 */
export interface CreateAdminUserReq {
  username: string
  email: string
  password: string
  nickname?: string
  /** 省略默认 student */
  role?: string
  /** 省略默认 1（正常） */
  status?: number
  level_id?: number
  subject_id?: number
}

/** 管理员编辑用户请求（全字段可选，指针型在 JSON 里按缺省省略） */
export interface UpdateAdminUserReq {
  nickname?: string
  email?: string
  role?: string
  status?: number
  level_id?: number
  subject_id?: number
}

/** 启/禁用请求体 */
export interface UpdateAdminUserStatusReq {
  status: number
}
