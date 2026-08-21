/**
 * 会话级存储工具（sessionStorage 封装）
 *
 * OPT-05 集中管理 sessionStorage 读写：
 * - 浏览器关闭后自动失效，避免 XSS 持久化窃取
 * - 后续其他敏感字段（如未来的 user_token_2fa 等）也走此模块
 */

const SESSION_PREFIX = 'sts_session:'

function safeSession(): Storage | null {
  try {
    if (typeof window !== 'undefined' && window.sessionStorage) {
      return window.sessionStorage
    }
  } catch {
    // ignore
  }
  return null
}

export function getSessionItem(key: string): string | null {
  const s = safeSession()
  if (!s) return null
  try {
    return s.getItem(SESSION_PREFIX + key)
  } catch {
    return null
  }
}

export function setSessionItem(key: string, value: string): void {
  const s = safeSession()
  if (!s) return
  try {
    s.setItem(SESSION_PREFIX + key, value)
  } catch {
    // ignore quota / disabled
  }
}

export function removeSessionItem(key: string): void {
  const s = safeSession()
  if (!s) return
  try {
    s.removeItem(SESSION_PREFIX + key)
  } catch {
    // ignore
  }
}

/**
 * 兼容旧 localStorage 中的 ai_key（OPT-05 迁移逻辑）：
 * 若发现 localStorage 中存在 ai_key，则迁移到 sessionStorage 后删除 localStorage 项。
 * 返回迁移后的 key 值；若不存在则返回 null。
 */
export function migrateAiKeyFromLocalStorage(): string | null {
  try {
    if (typeof window === 'undefined' || !window.localStorage) return null
    const legacy = window.localStorage.getItem('ai_key')
    if (!legacy) return null
    setSessionItem('ai_key', legacy)
    window.localStorage.removeItem('ai_key')
    return legacy
  } catch {
    return null
  }
}
