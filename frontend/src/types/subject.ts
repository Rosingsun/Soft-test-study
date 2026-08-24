export interface ExamLevel {
  id: number
  name: string
}

export interface Subject {
  id: number
  level_id: number
  name: string
  short_name: string
  description: string
  icon: string
  sort_order: number
}

export interface SubSubject {
  id: number
  subject_id: number
  name: string
  sort_order: number
}

export interface Chapter {
  id: number
  subject_id: number
  sub_subject_id: number
  parent_id: number
  name: string
  sort_order: number
  question_count: number
  material_id: number
}
