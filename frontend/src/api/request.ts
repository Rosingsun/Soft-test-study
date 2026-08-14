import type { ApiResponse } from '@/types/common'
import { showToast } from '@/utils/toast'

const BASE_URL = 'http://1.12.248.91:9000/api/v1'

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('access_token')
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options?.headers,
  }

  const res = await fetch(`${BASE_URL}${url}`, { ...options, headers })

  let json: ApiResponse<T>
  try {
    json = await res.json()
  } catch {
    const msg = `服务器返回了非 JSON 响应 (${res.status} ${res.statusText})`
    showToast(msg)
    throw new Error(msg)
  }

  if (json.code !== 0) {
    if (json.code === 10003) {
      localStorage.removeItem('access_token')
      window.location.href = '/login'
    }
    showToast(json.message || '请求失败')
    throw new Error(json.message)
  }

  return json.data as T
}

export function get<T>(url: string, params?: Record<string, string | number>) {
  const query = params
    ? '?' + new URLSearchParams(
        Object.entries(params).map(([k, v]) => [k, String(v)])
      ).toString()
    : ''
  return request<T>(url + query)
}

export function post<T>(url: string, body?: unknown) {
  return request<T>(url, {
    method: 'POST',
    body: body ? JSON.stringify(body) : undefined,
  })
}

export function put<T>(url: string, body?: unknown) {
  return request<T>(url, {
    method: 'PUT',
    body: body ? JSON.stringify(body) : undefined,
  })
}

export function patch<T>(url: string, body?: unknown) {
  return request<T>(url, {
    method: 'PATCH',
    body: body ? JSON.stringify(body) : undefined,
  })
}

export function del<T>(url: string) {
  return request<T>(url, { method: 'DELETE' })
}
