const BLOCKED_TAGS = ['script', 'iframe', 'object', 'embed', 'form', 'link', 'meta']

export function sanitizeHtml(html: string): string {
  if (!html) return ''
  const doc = new DOMParser().parseFromString(html, 'text/html')
  BLOCKED_TAGS.forEach(tag => {
    doc.querySelectorAll(tag).forEach(el => el.remove())
  })
  doc.querySelectorAll('*').forEach(el => {
    Array.from(el.attributes).forEach(attr => {
      const name = attr.name.toLowerCase()
      const value = attr.value.trim().toLowerCase()
      if (name.startsWith('on') || (name === 'href' && value.startsWith('javascript:'))) {
        el.removeAttribute(attr.name)
      }
    })
  })
  return doc.body.innerHTML
}
