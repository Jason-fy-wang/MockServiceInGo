/** Shared domain types for Mock Service Manager */

export type ResponseType = 'HTTP' | 'SSE' | 'WebSocket'

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'

export interface MockEndpoint {
  id: number
  method: HttpMethod
  path: string
  status: number
  type: ResponseType
}

export interface HeaderPair {
  key: string
  value: string
}

export type TabKey = 'All' | ResponseType

export interface NewMockPayload {
  method: HttpMethod
  path: string
  status: number
  type: ResponseType
}
