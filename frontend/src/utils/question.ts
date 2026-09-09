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
  multi_blank: '多空题',
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

export function difficultyLabel(d: string): string {
  return DIFFICULTY_LABELS[d as Difficulty] || d
}

/** 是否 AI 出题 */
export function isAiSource(source: string | undefined | null): boolean {
  return !!source && source.trim().toLowerCase() === 'ai'
}

/** 题源展示名：AI / 真题；空值返回空串（调用方按需隐藏） */
export function sourceLabel(source: string | undefined | null): string {
  if (!source) return ''
  return isAiSource(source) ? 'AI' : '真题'
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

// 解析多空题答案：JSON 数组字符串（如 '["A","C"]'）或已分割数组
export function parseBlankAnswer(answer: string | null | undefined): string[] {
  if (!answer) return []
  const s = answer.trim()
  if (s.startsWith('[')) {
    try {
      const arr = JSON.parse(s)
      if (Array.isArray(arr)) return arr.map(String)
    } catch {
      /* 忽略解析失败 */
    }
  }
  return s.split(',').map(x => x.trim()).filter(Boolean)
}

// 解析多空题每题选项 JSON：格式 [{"blank_index":1,"options":[{"id":"A","content":"..."},...]},...]
export function parseBlankOptions(blankOptionsStr?: string): { blank_index: number; options: { id: string; content: string }[] }[] {
  if (!blankOptionsStr) return []
  try {
    const arr = JSON.parse(blankOptionsStr) as unknown
    if (!Array.isArray(arr)) return []
    return (arr as any[])
      .filter(b => b && typeof b === 'object')
      .map(b => {
        const blank = b as { blank_index?: unknown; options?: unknown }
        return {
          blank_index: Number(blank.blank_index),
          options: Array.isArray(blank.options)
            ? (blank.options as any[])
                .filter(o => o && typeof o === 'object')
                .map(o => {
                  const opt = o as { id?: unknown; content?: unknown }
                  return { id: String(opt.id), content: String(opt.content) }
                })
            : [],
        }
      })
      .filter(b => b.options.length > 0)
  } catch {
    return []
  }
}

export function isCorrectAnswer(questionType: QuestionType | string, userAnswer: string, correctAnswer: string): boolean {
  const type = questionType as QuestionType
  if (type === 'multi_blank') {
    const a = parseBlankAnswer(userAnswer)
    const b = parseBlankAnswer(correctAnswer)
    if (a.length !== b.length) return false
    return a.every((v, i) => v === b[i])
  }
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
