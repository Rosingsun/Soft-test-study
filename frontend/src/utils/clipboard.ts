/**
 * 复制文本到剪贴板。
 *
 * 生产站点为 http（非安全上下文），`navigator.clipboard` 不可用，
 * 因此统一：「优先 Clipboard API → 回落 textarea + execCommand」。
 * 任一路径成功即返回 true。
 */
export async function copyText(text: string): Promise<boolean> {
  if (!text) return false

  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text)
      return true
    }
  } catch {
    // 权限被拒或非安全上下文，走下面的回落方案
  }

  return legacyCopy(text)
}

function legacyCopy(text: string): boolean {
  const textarea = document.createElement('textarea')
  textarea.value = text
  // 移出可视区但保留可编辑能力，避免部分浏览器因 display:none 无法选中
  textarea.setAttribute('readonly', '')
  textarea.style.position = 'fixed'
  textarea.style.top = '-9999px'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)

  try {
    textarea.select()
    textarea.setSelectionRange(0, text.length)
    return document.execCommand('copy')
  } catch {
    return false
  } finally {
    document.body.removeChild(textarea)
  }
}
