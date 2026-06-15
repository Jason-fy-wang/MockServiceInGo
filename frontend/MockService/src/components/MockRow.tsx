import { METHOD_STYLE, TYPE_BADGE } from '../constants/mock'
import type { MockEndpoint } from '../types/mock'

interface MockRowProps {
  mock: MockEndpoint
  onEdit?: (mock: MockEndpoint) => void
  onDelete?: (mock: MockEndpoint) => void
  onExpand?: (mock: MockEndpoint) => void
}

/**
 * MockRow — a single mock endpoint row in the list.
 */
export default function MockRow({ mock, onEdit, onDelete, onExpand }: MockRowProps) {
  const methodCls = METHOD_STYLE[mock.method] ?? METHOD_STYLE.GET
  const badgeCls = TYPE_BADGE[mock.type] ?? ''

  return (
    <div className="flex items-center justify-between bg-white border border-gray-200 rounded-xl px-5 py-3.5 hover:shadow-sm transition-shadow group">
      {/* Left: method + path */}
      <div className="flex items-center gap-3">
        <span className={`px-2.5 py-1 rounded-md text-xs font-semibold ${methodCls}`}>
          {mock.method}
        </span>
        <span className="text-sm font-mono text-gray-800">{mock.path}</span>
      </div>

      {/* Right: status + badge + actions */}
      <div className="flex items-center gap-3">
        <span className="px-2.5 py-0.5 rounded-full text-xs font-bold text-green-600 bg-green-50">
          {mock.status}
        </span>
        <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium ${badgeCls}`}>
          <TypeIcon type={mock.type} />
          {mock.type}
        </span>

        {onEdit && (
          <button
            onClick={() => onEdit(mock)}
            className="p-1.5 text-gray-400 hover:text-gray-600 opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
            title="Edit"
          >
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24" className="w-4 h-4">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
            </svg>
          </button>
        )}
        {onDelete && (
          <button
            onClick={() => onDelete(mock)}
            className="p-1.5 text-gray-400 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
            title="Delete"
          >
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24" className="w-4 h-4">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        )}
        {onExpand && (
          <button
            onClick={() => onExpand(mock)}
            className="p-1.5 text-gray-400 hover:text-gray-600 opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer"
            title="Expand"
          >
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24" className="w-4 h-4">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </button>
        )}
      </div>
    </div>
  )
}

function TypeIcon({ type }: { type: MockEndpoint['type'] }) {
  if (type === 'SSE' || type === 'WebSocket') {
    return (
      <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
          d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0" />
      </svg>
    )
  }
  if (type === 'HTTP') {
    return (
      <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
          d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
      </svg>
    )
  }
  return null
}
