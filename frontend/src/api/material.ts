import { get } from './request'
import { showToast } from '@/utils/toast'
import type { StudyMaterialResp } from '@/types/material'

const BASE_URL = '/api/v1'

export function listMaterials(subjectId?: number) {
  return get<StudyMaterialResp[]>('/materials', subjectId ? { subject_id: subjectId } : undefined)
}

export function getMaterial(id: number) {
  return get<StudyMaterialResp>(`/materials/${id}`)
}

// 下载 xmind：携带 Bearer token 拉取二进制流，并解析文件名触发浏览器下载
export async function downloadMaterial(id: number): Promise<void> {
  const token = localStorage.getItem('access_token')
  const res = await fetch(`${BASE_URL}/materials/${id}/download`, {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })

  if (!res.ok) {
    showToast(`下载失败 (${res.status})`, 'error')
    throw new Error(`下载失败: HTTP ${res.status}`)
  }

  // 后端统一错误响应为 HTTP 200 + JSON，识别后走统一错误提示
  const contentType = res.headers.get('Content-Type') || ''
  if (contentType.includes('application/json')) {
    const json = await res.json()
    showToast(json?.message || '下载失败', 'error')
    throw new Error(json?.message || '下载失败')
  }

  const blob = await res.blob()
  let fileName = `资料_${id}.xmind`
  const disposition = res.headers.get('Content-Disposition')
  if (disposition) {
    const match = disposition.match(/filename\*=UTF-8''([^;]+)/i) || disposition.match(/filename="?([^";]+)"?/i)
    if (match?.[1]) {
      fileName = decodeURIComponent(match[1])
    }
  }

  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = fileName
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}
