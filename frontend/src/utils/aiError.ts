/**
 * 将 AI 相关接口的原始错误信息格式化为用户友好的中文提示。
 * 通用入口：AIPractice、ChapterAiGenerateModal 等所有调 LLM 的场景共用。
 */
export function formatSubmitError(msg: string): string {
  if (!msg) return '提交失败，请稍后重试'
  if (msg.includes('429') || msg.includes('rate')) return '请求过于频繁，请稍候 1 分钟再试'
  if (msg.includes('timeout') || msg.includes('Timeout')) return 'AI 响应超时，请稍后重试'
  if (msg.includes('API Key') || msg.includes('api_key')) return 'API Key 无效，请前往「AI 配置」检查'
  if (msg.includes('401') || msg.includes('403')) return 'AI 鉴权失败，请检查 API Key 是否正确'
  return msg
}
