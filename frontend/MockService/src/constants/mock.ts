/** Mock endpoint type color config for tab cards & badges */
export const TYPE_CONFIG: Record<string, { bg: string; text: string; textInactive: string }> = {
  All: { bg: 'bg-gray-900', text: 'text-white', textInactive: 'text-black' },
  HTTP: { bg: 'bg-blue-600', text: 'text-white', textInactive: 'text-black' },
  SSE: { bg: 'bg-purple-600', text: 'text-white', textInactive: 'text-black' },
  WebSocket: { bg: 'bg-green-500', text: 'text-white', textInactive: 'text-black' },
}

/** HTTP method badge colors */
export const METHOD_STYLE: Record<string, string> = {
  POST: 'bg-green-100 text-green-700',
  GET: 'bg-blue-100 text-blue-700',
  PUT: 'bg-yellow-100 text-yellow-700',
  DELETE: 'bg-red-100 text-red-700',
  PATCH: 'bg-purple-100 text-purple-700',
}

/** Type badge (pill) styles in list rows */
export const TYPE_BADGE: Record<string, string> = {
  SSE: 'bg-purple-50 text-purple-600 border border-purple-200',
  WebSocket: 'bg-green-50 text-green-600 border border-green-200',
  HTTP: 'bg-blue-50 text-blue-600 border border-blue-200',
}

/** Available response types in the form (form internal labels) */
export const RESPONSE_TYPES = ['Http', 'Sse', 'WebSocket'] as const

/** HTTP method options */
export const HTTP_METHODS = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'] as const

/** Common status code presets */
export const STATUS_CODE_PRESETS = ['200', '201', '204', '400', '404', '500'] as const
