export interface StudyPlanCreateReq {
  title: string
  subject_id?: number
  daily_goal: number
  start_date: string
  end_date: string
}

export interface StudyPlanUpdateReq {
  title?: string
  subject_id?: number
  daily_goal?: number
  start_date?: string
  end_date?: string
}

export interface StudyPlanResp {
  id: number
  subject_id: number
  title: string
  daily_goal: number
  start_date: string
  end_date: string
  status: number
  today_done: number
  today_goal: number
  overall_done: number
  overall_goal: number
  completion_pct: number
  remaining_days: number
  created_at: string
}
