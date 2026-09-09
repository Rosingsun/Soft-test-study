/**
 * 站点地址与邀请链接生成。
 *
 * VITE_SITE_URL 用于多环境下生成正确的前台链接：
 * 未配置时降级为当前浏览器 origin（本地开发与同源部署均可直接工作）。
 */
export function getSiteUrl(): string {
  const fromEnv = String(import.meta.env.VITE_SITE_URL || '').trim()
  if (fromEnv) return fromEnv.replace(/\/+$/, '')
  return window.location.origin
}

/** 生成带邀请码的注册邀请链接 */
export function buildInviteUrl(code: string): string {
  return `${getSiteUrl()}/register?invite_code=${encodeURIComponent(code)}`
}
