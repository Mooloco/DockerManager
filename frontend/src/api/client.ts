const BASE = '/api/v1'

/** 业务错误(带错误码) */
export class APIError extends Error {
  code: string
  constructor(code: string, message: string) {
    super(message)
    this.code = code
  }
}

interface ApiBody<T> {
  success: boolean
  data?: T
  error?: { code: string; message: string }
}

/**
 * 统一请求封装:
 * - 自动携带 session cookie
 * - 解析 {success, data, error} 信封
 * - 401 时跳转登录页
 */
export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const resp = await fetch(BASE + path, {
    headers: { 'Content-Type': 'application/json' },
    credentials: 'same-origin',
    ...options,
  })

  let body: ApiBody<T>
  try {
    body = await resp.json()
  } catch {
    throw new APIError('BAD_RESPONSE', `服务器响应异常 (HTTP ${resp.status})`)
  }

  if (resp.status === 401) {
    // 登录已过期
    if (!window.location.pathname.startsWith('/login')) {
      window.location.href = '/login?expired=1'
    }
    throw new APIError('UNAUTHORIZED', body.error?.message || '未登录或登录已过期')
  }
  if (resp.status === 403) {
    throw new APIError('FORBIDDEN', body.error?.message || '没有权限')
  }
  if (!body.success) {
    throw new APIError(body.error?.code || 'UNKNOWN', body.error?.message || '请求失败')
  }
  return body.data as T
}

export const get = <T>(path: string) => request<T>(path)
export const post = <T>(path: string, data?: unknown) =>
  request<T>(path, { method: 'POST', body: data ? JSON.stringify(data) : undefined })
export const del = <T>(path: string) => request<T>(path, { method: 'DELETE' })
