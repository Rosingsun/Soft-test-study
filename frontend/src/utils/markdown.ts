// 极简 Markdown 渲染：仅支持基础语法（**bold**、`code`、换行、列表、# 标题）
// 输出经 sanitizeHtml 清洗，应用于 AI 描述等可信文本。

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function renderInline(text: string): string {
  let out = escapeHtml(text)
  out = out.replace(/`([^`\n]+)`/g, '<code class="rounded bg-gray-100 px-1 py-0.5 text-[0.85em] text-indigo-600">$1</code>')
  out = out.replace(/\*\*([^*\n]+)\*\*/g, '<strong class="font-semibold text-gray-900">$1</strong>')
  out = out.replace(/(^|[^*])\*([^*\n]+)\*(?!\*)/g, '$1<em class="italic">$2</em>')
  return out
}

export function renderMarkdown(src: string): string {
  if (!src) return ''
  const lines = src.replace(/\r\n/g, '\n').split('\n')
  const out: string[] = []
  let inList = false
  let paraBuf: string[] = []

  function flushPara() {
    if (paraBuf.length) {
      const html = renderInline(paraBuf.join(' ')).replace(/\n/g, '<br>')
      out.push(`<p class="mb-2 leading-6 last:mb-0">${html}</p>`)
      paraBuf = []
    }
  }

  function closeList() {
    if (inList) {
      out.push('</ul>')
      inList = false
    }
  }

  for (const raw of lines) {
    const line = raw.trimEnd()
    if (!line.trim()) {
      flushPara()
      closeList()
      continue
    }
    const h = line.match(/^(#{1,3})\s+(.+)$/)
    if (h) {
      flushPara()
      closeList()
      const level = h[1].length
      const cls =
        level === 1
          ? 'mb-2 mt-3 text-base font-semibold text-gray-900 first:mt-0'
          : level === 2
          ? 'mb-1.5 mt-2 text-sm font-semibold text-gray-900 first:mt-0'
          : 'mb-1 mt-2 text-xs font-semibold text-gray-700 first:mt-0'
      out.push(`<h${level} class="${cls}">${renderInline(h[2])}</h${level}>`)
      continue
    }
    const li = line.match(/^[-*]\s+(.+)$/)
    if (li) {
      flushPara()
      if (!inList) {
        out.push('<ul class="mb-2 list-disc space-y-0.5 pl-5 last:mb-0">')
        inList = true
      }
      out.push(`<li class="leading-6">${renderInline(li[1])}</li>`)
      continue
    }
    closeList()
    paraBuf.push(line)
  }
  flushPara()
  closeList()
  return out.join('')
}
