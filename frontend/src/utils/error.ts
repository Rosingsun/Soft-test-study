// 统一错误码 → 文案映射
// 后端错误码定义见 backend/internal/config/errors.go
const CODE_MESSAGES: Record<number, string> = {
  10006: '操作过于频繁，请稍后再试',
}

// 默认文案
const DEFAULT_ERROR_MESSAGE = '请求失败'

/**
 * 将后端错误码转换为面向用户的提示文案
 * 命中已知 code 时返回映射文案，否则返回服务端 message 或默认文案
 */
export function getErrorMessage(code: number, serverMessage?: string): string {
  if (code !== 0 && CODE_MESSAGES[code]) {
    return CODE_MESSAGES[code]
  }
  return serverMessage || DEFAULT_ERROR_MESSAGE
}

/**
 * 判断是否为限流错误（用于上层决定是否额外 toast / 走降级逻辑）
 */
export function isRateLimitError(code: number): boolean {
  return code === 10006
}
