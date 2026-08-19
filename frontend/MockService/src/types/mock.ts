/** Shared domain types for Mock Service Manager */

export type ResponseType =
  "HTTP" | "SSE" | "WebSocket" | "http" | "sse" | "websocket";

export type HttpMethod = "GET" | "POST" | "PUT" | "DELETE" | "PATCH";

export type MockEndpoint = MockRecord;

export interface SSEEvent {
  event: string;
  data: string;
  delay: number;
}

export interface WebSocketMessage {
  message: string;
  delay: number;
  type: "text" | "binary";
}

export interface HeaderPair {
  key: string;
  value: string;
}

export interface HealthResponse {
  message: string;
  routes: string;
  features: string;
}

export interface MockRecord {
  id: string | number;
  method: string;
  path: string;
  requestHeaders?: HeaderPair[];
  requestBody?: string;
  requestQuery?: HeaderPair[];
  responseStatus?: number;
  responseHeaders?: HeaderPair[];
  responseBody?: string;
  responseType: ResponseType;
  sseEvents?: SSEEvent[];
  websocketMessages?: WebSocketMessage[];
}

export interface MockListResponse {
  mocks: MockRecord[];
}

export type TabKey = "All" | ResponseType;

export interface ApiResponse<T> {
  data: T;
  status: number;
  headers: Record<string, string>;
}

export interface MockApiResponse {
  message: string;
  error: string;
}
