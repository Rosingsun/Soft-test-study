export type RankingCategory = 'checkin' | 'exam' | 'practice'
export type RankingMetric = 'streak' | 'accuracy' | 'score'

export interface RankingScopeResp {
  level_id: number
  subject_id: number
  total_users: number
}

export interface RankingMyResp {
  rank: number
  value: number
  total_participants: number
  percentile: number
}

export interface RankingReferenceResp {
  avg: number
  top10_threshold: number
}

export interface DistributionItem {
  label: string
  count: number
  ratio: number
  is_mine: boolean
}

export interface RankingResp {
  category: RankingCategory
  metric: RankingMetric
  scope: RankingScopeResp
  my?: RankingMyResp | null
  reference: RankingReferenceResp
  distribution: DistributionItem[]
}

export interface SectionChapterItem {
  chapter_id: number
  chapter_name: string
  weight: number
  attempted: number
  correct: number
  accuracy: number
  est_score: number
}

export interface SectionBreakdown {
  section: 'choice' | 'case_study' | 'essay'
  section_name: string
  max_score: number
  attempted: number
  est_score: number
  accuracy: number
  sample_size: number
  sufficient: boolean
  chapters: SectionChapterItem[]
}

export interface EstimatedScoreResp {
  subject_id: number
  total_est: number
  total_max: number
  sections: SectionBreakdown[]
  generated_at: string
}
