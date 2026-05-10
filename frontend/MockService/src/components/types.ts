export interface MockRoute {
  key: string
  method: string
  path: string
  requestHeaders: Record<string, string>
  requestBody?: string
  requestQuery: Record<string, string>
  responseStatus?: number
  responseHeaders: Record<string, string>
  responseBody?: string
}

export const methodOptions = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE','HEAD', 'OPTIONS', 'TRACE', 'CONNECT']

export const defaultRoute: Omit<MockRoute, 'key'> = {
  method: 'GET',
  path: '/example',
  requestHeaders: {},
  requestBody: '',
  requestQuery: {},
  responseStatus: 200,
  responseHeaders: {},
  responseBody: '{"message":"ok"}',
}
