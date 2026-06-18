/** Shared domain types for Mock Service Manager */

export type ResponseType = 'HTTP' | 'SSE' | 'WebSocket' |'http' | 'sse' | 'websocket'

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'

export type MockEndpoint = MockRecord

export interface SSEEvent {
  event: string
  data: string
  delay: number
}

export interface webSocketMessage {
  message: string
  delay: number
  type: 'text' | 'binary'
}

export interface HeaderPair {
  key: string
  value: string
}


export interface HealthResponse {
  message: string
  routes: string
  features: string
}

export interface MockRecord {
  id: string | number
  method: string
  path: string
  requestHeaders?: Record<string, string>
  requestBody?: string
  requestQuery?: Record<string, string>
  responseStatus?: number
  responseHeaders?: Record<string, string>
  responseBody?: string
  responseType: ResponseType
  sseEvents?: SSEEvent[]
  websocketMessages?: webSocketMessage[]
}


export interface MockListResponse {
  mocks: MockRecord[]
}

export type TabKey = 'All' | ResponseType

export interface NewMockPayload {
  method: HttpMethod
  path: string
  status: number
  responseType: ResponseType
}
