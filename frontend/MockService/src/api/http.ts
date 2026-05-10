export class ApiError extends Error {
  status: number
  data: unknown

  constructor(message: string, status: number, data: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.data = data
  }
}

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | 'HEAD' | 'OPTIONS' | 'TRACE' | 'CONNECT'

type RequestOptions = {
  method?: HttpMethod
  body?: unknown
  headers?: HeadersInit
  signal?: AbortSignal
}

const DEFAULT_BASE_URL = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? 'http://192.168.20.21:8080'

function joinUrl(path: string) {
  if (!DEFAULT_BASE_URL) {
    return path
  }
  const base = DEFAULT_BASE_URL.endsWith('/') ? DEFAULT_BASE_URL.slice(0, -1) : DEFAULT_BASE_URL
  const route = path.startsWith('/') ? path : `/${path}`
  return `${base}${route}`
}

function isFormData(value: unknown): value is FormData {
  return typeof FormData !== 'undefined' && value instanceof FormData
}

export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, headers, signal } = options

  const finalHeaders = new Headers(headers)
  const requestInit: RequestInit = {
    method,
    headers: finalHeaders,
    signal,
  }

  if (typeof body !== 'undefined') {
    if (isFormData(body)) {
      requestInit.body = body
    } else {
      if (!finalHeaders.has('Content-Type')) {
        finalHeaders.set('Content-Type', 'application/json')
      }
      requestInit.body = JSON.stringify(body)
    }
  }
  const urlPath = joinUrl(path)
  console.log(`Making request to ${method} ${urlPath}`)
  const response = await fetch(urlPath, requestInit)
  const contentType = response.headers.get('content-type')
  const hasJsonBody = contentType?.includes('application/json')
  const payload = hasJsonBody ? await response.json() : await response.text()

  if (!response.ok) {
    throw new ApiError(`Request failed: ${response.status}`, response.status, payload)
  }

  return payload as T
}
