import type { QuestionType, Difficulty } from '@/types/question'

export const TYPE_LABELS: Record<string, string> = {
  single: '单选题',
  multi: '多选题',
  judge: '判断题',
  fill: '填空题',
  short: '简答题',
  comprehensive: '综合题',
  essay: '论文题',
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
