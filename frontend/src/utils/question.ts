import type { QuestionType, Difficulty } from '@/types/question'

export const TYPE_LABELS: Record<string, string> = {
  single: '单选题',
  multi: '多选题',
  judge: '判断题',
  fill: '填空题',
  short: '简答题',
  comprehensive: '综合题',
  essay: '论文题',
  case_study: '案例分析',
}

export const DIFFICULTY_LABELS: Record<Difficulty, string> = {
  easy: '简单',
  medium: '中等',
  hard: '困难',
}

export const DIFFICULTY_COLORS: Record<Difficulty, string> = {
  easy: 'bg-emerald-50 text-emerald-600',
  medium: 'bg-yellow-50 text-yellow-600',
  hard: 'bg-red-50 text-red-600',
}

export function typeLabel(type: string): string {
  return TYPE_LABELS[type] || type
}

// ===== AI 评分维度（论文 / 案例分析）=====

export interface AiScoreDimension {
  key: 'argument' | 'structure' | 'language' | 'depth'
  label: string
  max: number
}

export interface AiScoreSpec {
  /** 评分标题，如「论文 AI 评分」 */
  title: string
  /** 已有评分提示，如「已对此论文评过分」 */
  scoredHint: string
  /** 总分分母 */
  totalMax: number
  /** 四个维度的标签与满分（argument/structure/language/depth） */
  dimensions: AiScoreDimension[]
}

/** 按题型返回 AI 评分展示规格 */
export function aiScoreSpec(type: string): AiScoreSpec {
  if (type === 'case_study') {
    return {
      title: '案例分析 AI 评分',
      scoredHint: '已对此案例分析评过分',
      totalMax: 80,
      dimensions: [
        { key: 'argument', label: '要点完整性', max: 20 },
        { key: 'structure', label: '分析逻辑', max: 20 },
        { key: 'language', label: '专业术语', max: 20 },
        { key: 'depth', label: '方案可行性', max: 20 },
      ],
    }
  }
  return {
    title: '论文 AI 评分',
    scoredHint: '已对此论文评过分',
    totalMax: 75,
    dimensions: [
      { key: 'argument', label: '论点与立意', max: 20 },
      { key: 'structure', label: '结构与逻辑', max: 20 },
      { key: 'language', label: '语言表达', max: 20 },
      { key: 'depth', label: '深度与广度', max: 15 },
    ],
  }
}

/** 取某维度的得分（argument/structure/language/depth），避免以 string 索引对象 */
export function aiDimensionScore(score: { argument_score: number; structure_score: number; language_score: number; depth_score: number }, key: AiScoreDimension['key']): number {
  switch (key) {
    case 'argument':
      return score.argument_score
    case 'structure':
      return score.structure_score
    case 'language':
      return score.language_score
    case 'depth':
      return score.depth_score
  }
}

export function parseOptions(optionsStr: string): { label: string; text: string }[] {
  try {
    const obj = JSON.parse(optionsStr)
    if (Array.isArray(obj)) {
      return obj.map((item, i) => ({
        label: String.fromCharCode(65 + i),
        text: String(item),
      }))
    }
    return Object.entries(obj).map(([k, v]) => ({ label: k, text: v as string }))
  } catch {
    return []
  }
}

export function normalizeAnswer(answer: string | null | undefined): string {
  return (answer || '').trim().toUpperCase()
}

export function isCorrectAnswer(questionType: QuestionType | string, userAnswer: string, correctAnswer: string): boolean {
  const type = questionType as QuestionType
  if (type === 'multi') {
    const a = (userAnswer || '').split(',').map(s => s.trim()).filter(Boolean).sort().join(',')
    const b = (correctAnswer || '').split(',').map(s => s.trim()).filter(Boolean).sort().join(',')
    return a === b
  }
  if (type === 'judge') {
    const u = userAnswer?.trim() || ''
    const c = correctAnswer?.trim() || ''
    return (u === '正确' && c === 'A') || (u === '错误' && c === 'B') || u === c || normalizeAnswer(u) === normalizeAnswer(c)
  }
  return normalizeAnswer(userAnswer) === normalizeAnswer(correctAnswer)
}
