import { request } from './http'

export interface HealthResponse {
  message: string
  routes: string
  features: string
}

export interface MockRecord {
  method: string
  path: string
  requestHeaders?: Record<string, string>
  requestBody?: unknown
  requestQuery?: Record<string, string>
  responseStatus?: number
  responseHeaders?: Record<string, string>
  responseBody?: unknown
}

export interface MockListResponse {
  mocks: MockRecord[]
}

export async function getHealth(signal?: AbortSignal) {
  return request<HealthResponse>('/v1/health', { signal })
}

export async function registerMock(payload: MockRecord, signal?: AbortSignal) {
  return request<unknown>('/v1/__mock', {
    method: 'POST',
    body: payload,
    signal,
  })
}

export async function uploadMockConfig(file: File, signal?: AbortSignal) {
  const form = new FormData()
  form.append('file', file)

  return request<unknown>('/v1/__mock/upload', {
    method: 'POST',
    body: form,
    signal,
  })
}

export async function listMocks(signal?: AbortSignal) {
  return request<MockListResponse>('/v1/__mock', { signal })
}

export async function clearMocks(signal?: AbortSignal) {
  return request<unknown>('/v1/__mock/all', {
    method: 'DELETE',
    signal,
  })
}

export async function deleteMockByMethod(method: string, signal?: AbortSignal) {
  return request<unknown>(`/v1/__mock/${encodeURIComponent(method)}`, {
    method: 'DELETE',
    signal,
  })
}
