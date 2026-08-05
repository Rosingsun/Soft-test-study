import { get, post } from './request'
import type { PracticeSubmitReq, PracticeRecordResp, Question } from '@/types/question'

export function submitPractice(data: PracticeSubmitReq) {
  return post<PracticeRecordResp>('/practice/submit', data)
}

export function getRandomQuestions(subjectId: number, count = 10, difficulty?: string) {
  return get<Question[]>('/questions/random', {
    subject_id: subjectId,
    count,
    ...(difficulty ? { difficulty } : {}),
  })
}

export function getSpecialQuestions(subjectId: number, type?: string, difficulty?: string, count = 10) {
  return get<Question[]>('/questions/special', {
    subject_id: subjectId,
    count,
    ...(type ? { type } : {}),
    ...(difficulty ? { difficulty } : {}),
  })
}

export function getEssayQuestions(subjectId: number, years = 5) {
  return get<Question[]>('/questions/essays', {
    subject_id: subjectId,
    years,
  })
}

export function getQuestionById(id: number) {
  return get<Question>(`/questions/${id}`)
}

export function generateQuestions(data: any) {
  return post('/ai/generate', data).catch((err) => {
    // 开发环境下后端可能未启动，返回 mock 数据以便前端调试
    // 仅在 Vite 开发模式下启用
    // 注意：生产环境不会触发此分支
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    if (import.meta && import.meta.env && import.meta.env.DEV) {
      return {
        questions: [
          {
            id: 0,
            type: 'essay',
            difficulty: data.difficulty || 'medium',
            content: `（模拟生成）请结合${data.subject_id}学科，论述软件需求分析中的关键步骤及其风险防范措施。`,
            options: '[]',
            answer: '',
            analysis: '',
            knowledge_point: '需求分析',
            year: new Date().getFullYear(),
          },
        ],
      }
    }
    throw err
  })
}

