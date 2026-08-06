export function formatDuration(seconds: number | null | undefined): string {
  const s = Number(seconds) || 0
  if (s < 60) return `${s}秒`
  const m = Math.floor(s / 60)
  if (m < 60) return `${m}分${s % 60}秒`
  const h = Math.floor(m / 60)
  return `${h}小时${m % 60}分`
}

export function formatDate(dateStr: string | null | undefined): string {
  if (!dateStr) return '-'
  return dateStr
}

export function formatPercent(value: number | null | undefined): string {
  const n = Number(value) || 0
  return n.toFixed(1) + '%'
}

export function formatFileSize(bytes: number | null | undefined): string {
  const n = Number(bytes) || 0
  if (n <= 0) return '-'
  if (n < 1024) return `${n}B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)}KB`
  if (n < 1024 * 1024 * 1024) return `${(n / (1024 * 1024)).toFixed(1)}MB`
  return `${(n / (1024 * 1024 * 1024)).toFixed(2)}GB`
}

export function timeAgo(dateStr: string): string {
  const date = new Date(dateStr.replace(' ', 'T'))
  if (isNaN(date.getTime())) return dateStr
  const diff = Date.now() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes} 分钟前`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} 小时前`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days} 天前`
  return dateStr
}
