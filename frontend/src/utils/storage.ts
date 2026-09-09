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

// ---------------------------------------------------------------------------
// 「显式记住」场景的 localStorage 封装
//
// 默认敏感数据（如 AI API Key）只存 sessionStorage，关闭标签页即失效。
// 只有用户在 UI 上主动勾选「记住」时，才通过下面这组函数写入 localStorage。
// 统一加前缀，便于与历史无前缀的 localStorage 键（ai_key / ai_api_meta）区分。
// ---------------------------------------------------------------------------

const LOCAL_PREFIX = 'sts_local:'

function safeLocal(): Storage | null {
  try {
    if (typeof window !== 'undefined' && window.localStorage) {
      return window.localStorage
    }
  } catch {
    // ignore
  }
  return null
}

export function getLocalItem(key: string): string | null {
  const s = safeLocal()
  if (!s) return null
  try {
    return s.getItem(LOCAL_PREFIX + key)
  } catch {
    return null
  }
}

export function setLocalItem(key: string, value: string): void {
  const s = safeLocal()
  if (!s) return
  try {
    s.setItem(LOCAL_PREFIX + key, value)
  } catch {
    // ignore quota / disabled
  }
}

export function removeLocalItem(key: string): void {
  const s = safeLocal()
  if (!s) return
  try {
    s.removeItem(LOCAL_PREFIX + key)
  } catch {
    // ignore
  }
}

/** 读取「记住」开关（无前缀的历史键，保持向后兼容） */
export function getRememberFlag(key: string): boolean {
  const s = safeLocal()
  if (!s) return false
  try {
    return s.getItem(key) === '1'
  } catch {
    return false
  }
}

/** 写入「记住」开关 */
export function setRememberFlag(key: string, on: boolean): void {
  const s = safeLocal()
  if (!s) return
  try {
    s.setItem(key, on ? '1' : '0')
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
