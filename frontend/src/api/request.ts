import type { Router } from 'vue-router'
import type { ApiResponse } from '@/types/common'
import { showToast } from '@/utils/toast'
import { getErrorMessage } from '@/utils/error'

const BASE_URL = import.meta.env.DEV ? '/api/v1' : 'http://www.fazhiyinqing.cn:3000/api/v1'

// OPT-22: 由 main.ts 在挂载前注入 router 实例，避免循环依赖
let _router: Router | null = null

export function setRouter(router: Router) {
  _router = router
}

// OPT-22: 并发 401 时仅触发一次跳转
let _redirectingToLogin = false

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
      // 未登录 / Token 失效：清掉本地 token，保留 from 路径后跳到登录页
      localStorage.removeItem('access_token')
      if (_router && !_redirectingToLogin) {
        _redirectingToLogin = true
        const from = _router.currentRoute.value.fullPath
        _router
          .push({ name: 'Login', query: from && from !== '/login' ? { from } : undefined })
          .finally(() => {
            // 给本轮并发请求留出充分时间，500ms 后再开放
            setTimeout(() => {
              _redirectingToLogin = false
            }, 500)
          })
      } else if (!_router) {
        // router 尚未注入（理论不应该发生），降级为硬跳转
        window.location.href = '/login'
      }
      // 401 跳转到登录后不再 toast（避免连续弹多条）
      throw new Error(json.message)
    }
    showToast(getErrorMessage(json.code, json.message))
    throw new Error(json.message || getErrorMessage(json.code))
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
