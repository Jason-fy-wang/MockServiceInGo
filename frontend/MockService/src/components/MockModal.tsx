import { useState, useEffect } from 'react'
import { HTTP_METHODS, STATUS_CODE_PRESETS, RESPONSE_TYPES } from '../constants/mock'
import type { HeaderPair, MockEndpoint, SSEEvent, WebSocketMessage } from '../types/mock'

const SELECT_ARROW_BG = `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%236b7280'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'%3E%3C/path%3E%3C/svg%3E")`

const selectCls = `w-full px-3.5 py-2.5 bg-white border border-gray-300 rounded-lg text-sm text-gray-800 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:border-transparent appearance-none cursor-pointer`
const inputCls = `w-full px-3.5 py-2.5 bg-white border border-gray-300 rounded-lg text-sm text-gray-800 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:border-transparent`

const TYPE_TO_FORM: Record<string, string> = {
  HTTP: 'Http', SSE: 'Sse', WebSocket: 'WebSocket',
}

interface MockModalProps {
  open: boolean
  onClose: () => void
  onAdd?: (payload: MockEndpoint) => void
  onEdit?: (old: MockEndpoint, payload: MockEndpoint) => void
  editing?: MockEndpoint | null
}

export default function MockModal({ open, onClose, onAdd, onEdit, editing }: MockModalProps) {
  const isEdit = !!editing

  // ── Common fields ──
  const [method, setMethod] = useState('GET')
  const [path, setPath] = useState('/v1/')
  const [statusCodePreset, setStatusCodePreset] = useState('200')
  const [statusCodeInput, setStatusCodeInput] = useState('200')
  const [responseType, setResponseType] = useState('Http')
  const [body, setBody] = useState('{ "key": "value" }')
  const [headers, setHeaders] = useState<HeaderPair[]>([
    { key: 'content-type', value: 'application/json' },
  ])

  // ── SSE events array ──
  const [sseEvents, setSseEvents] = useState<SSEEvent[]>([
    { event: 'message', data: '{"foo":"bar"}', delay: 1000 },
  ])

  // ── WebSocket messages array ──
  const [wsMessages, setWsMessages] = useState<WebSocketMessage[]>([
    { message: '{"type":"welcome","data":"hello"}', delay: 1000, type: 'text' },
  ])

  // Sync state when opening modal
  useEffect(() => {
    if (!open) return
    if (isEdit && editing) {
      // Edit mode → pre-fill
      setMethod(editing.method)
      setPath(editing.path)
      const s = String(editing.responseStatus ?? 200)
      setStatusCodePreset(s)
      setStatusCodeInput(s)
      setResponseType(TYPE_TO_FORM[editing.responseType] ?? 'Http')
      setBody('{ "key": "value" }')
      setHeaders([{ key: 'content-type', value: 'application/json' }])
      setSseEvents(editing.sseEvents?.length ? [...editing.sseEvents] : [{ event: 'message', data: '{"foo":"bar"}', delay: 1000 }])
      setWsMessages(editing.websocketMessages?.length ? [...editing.websocketMessages] : [{ message: '{"type":"welcome"}', delay: 1000, type: 'text' }])
    } else {
      // Add mode → reset
      setMethod('GET')
      setPath('/v1/')
      setStatusCodePreset('200')
      setStatusCodeInput('200')
      setResponseType('Http')
      setBody('{ "key": "value" }')
      setHeaders([{ key: 'content-type', value: 'application/json' }])
      setSseEvents([{ event: 'message', data: '{"foo":"bar"}', delay: 1000 }])
      setWsMessages([{ message: '{"type":"welcome"}', delay: 1000, type: 'text' }])
    }
  }, [open, isEdit, editing])

  if (!open) return null

  // ── Header helpers ──
  const addHeader = () => setHeaders([...headers, { key: '', value: '' }])
  const removeHeader = (i: number) => setHeaders(headers.filter((_, idx) => idx !== i))
  const updateHeaderKey = (i: number, v: string) =>
    setHeaders(headers.map((h, idx) => idx === i ? { ...h, key: v } : h))
  const updateHeaderValue = (i: number, v: string) =>
    setHeaders(headers.map((h, idx) => idx === i ? { ...h, value: v } : h))

  // ── SSE event helpers ──
  const addSseEvent = () =>
    setSseEvents([...sseEvents, { event: 'message', data: '', delay: 1000 }])
  const removeSseEvent = (i: number) =>
    setSseEvents(sseEvents.filter((_, idx) => idx !== i))
  const updateSseEvent = (i: number, field: keyof SSEEvent, v: string) =>
    setSseEvents(sseEvents.map((e, idx) => idx === i ? { ...e, [field]: field === 'delay' ? parseInt(v) || 0 : v } : e))

  // ── WS message helpers ──
  const addWsMessage = () =>
    setWsMessages([...wsMessages, { message: '', delay: 1000, type: 'text' }])
  const removeWsMessage = (i: number) =>
    setWsMessages(wsMessages.filter((_, idx) => idx !== i))
  const updateWsMessage = (i: number, field: keyof WebSocketMessage, v: string) =>
    setWsMessages(wsMessages.map((m, idx) => idx === i ? { ...m, [field]: field === 'delay' ? parseInt(v) || 0 : v } : m))

  const handleSubmit = () => {
    const typeMap: Record<string, MockEndpoint['responseType']> = {
      Http: 'HTTP', Sse: 'SSE', WebSocket: 'WebSocket',
    }
    const payload: MockEndpoint = {
      id: Date.now(),
      method: method as MockEndpoint['method'],
      path,
      responseStatus: parseInt(statusCodeInput),
      responseType: typeMap[responseType] ?? 'HTTP',
    }

    if (responseType === 'Sse') {
      payload.sseEvents = sseEvents
    }
    if (responseType === 'WebSocket') {
      payload.websocketMessages = wsMessages
    }

    if (isEdit && onEdit && editing) {
      onEdit(editing, payload)
    } else if (onAdd) {
      onAdd(payload)
    }
    onClose()
  }

  const isSSE = responseType === 'Sse'
  const isWS = responseType === 'WebSocket'

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-16">
      <div className="absolute inset-0 bg-black/20" onClick={onClose} />
      <div className="relative w-full max-w-2xl mx-4 bg-white rounded-2xl shadow-xl">
        <ModalHeader title={isEdit ? 'Edit Mock Service' : 'Add Mock Service'} onClose={onClose} />

        <div className="px-6 py-5 space-y-5 max-h-[70vh] overflow-y-auto">
          {/* Row 1: Method + Path */}
          <div className="grid grid-cols-2 gap-4">
            <LabeledField label="Method">
              <select value={method} onChange={e => setMethod(e.target.value)} className={selectCls}
                style={{ backgroundImage: SELECT_ARROW_BG, backgroundRepeat: 'no-repeat', backgroundPosition: 'right 10px center', backgroundSize: '16px' }}>
                {HTTP_METHODS.map(m => <option key={m} value={m}>{m}</option>)}
              </select>
            </LabeledField>
            <LabeledField label="Path">
              <input type="text" value={path} onChange={e => setPath(e.target.value)} className={`${inputCls} font-mono`} />
            </LabeledField>
          </div>

          {/* Row 2: Status Code + Response Type */}
          <div className="grid grid-cols-[120px_140px_1fr] gap-4 items-start">
            <LabeledField label="Status Code">
              <select value={statusCodePreset} onChange={e => { setStatusCodePreset(e.target.value); setStatusCodeInput(e.target.value) }}
                className={selectCls}
                style={{ backgroundImage: SELECT_ARROW_BG, backgroundRepeat: 'no-repeat', backgroundPosition: 'right 10px center', backgroundSize: '16px' }}>
                {STATUS_CODE_PRESETS.map(s => <option key={s} value={s}>{s}</option>)}
              </select>
            </LabeledField>
            <div>
              <label className="block text-sm font-medium text-transparent mb-1.5">&nbsp;</label>
              <input type="text" value={statusCodeInput} onChange={e => setStatusCodeInput(e.target.value)} className={inputCls} />
            </div>
            <LabeledField label="Response Type">
              <SegmentedControl options={[...RESPONSE_TYPES]} value={responseType} onChange={setResponseType} />
            </LabeledField>
          </div>

          {/* ─── SSE Events (array) ─── */}
          {isSSE && (
            <div className="space-y-3 p-4 bg-purple-50/50 border border-purple-100 rounded-xl">
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-purple-600 uppercase tracking-wider">SSE Events</span>
                <button onClick={addSseEvent} className="inline-flex items-center gap-1 text-xs font-medium text-purple-600 hover:text-purple-700 cursor-pointer">
                  <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                  </svg>
                  Add Event
                </button>
              </div>
              {sseEvents.map((evt, i) => (
                <SseEventCard
                  key={i}
                  evt={evt}
                  onUpdate={(field, v) => updateSseEvent(i, field, v)}
                  onRemove={() => sseEvents.length > 1 ? removeSseEvent(i) : undefined}
                  canRemove={sseEvents.length > 1}
                />
              ))}
            </div>
          )}

          {/* ─── WebSocket Messages (array) ─── */}
          {isWS && (
            <div className="space-y-3 p-4 bg-green-50/50 border border-green-100 rounded-xl">
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-green-600 uppercase tracking-wider">WebSocket Messages</span>
                <button onClick={addWsMessage} className="inline-flex items-center gap-1 text-xs font-medium text-green-600 hover:text-green-700 cursor-pointer">
                  <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                  </svg>
                  Add Message
                </button>
              </div>
              {wsMessages.map((msg, i) => (
                <WsMessageCard
                  key={i}
                  msg={msg}
                  onUpdate={(field, v) => updateWsMessage(i, field, v)}
                  onRemove={() => wsMessages.length > 1 ? removeWsMessage(i) : undefined}
                  canRemove={wsMessages.length > 1}
                />
              ))}
            </div>
          )}

          {/* Row 3: Response Body (HTTP only) */}
          {!isSSE && !isWS && (
            <LabeledField label="Response Body">
              <textarea value={body} onChange={e => setBody(e.target.value)} rows={5}
                className={`w-full px-3.5 py-3 ${inputCls} font-mono resize-y`} />
            </LabeledField>
          )}

          {/* Row 4: Response Headers */}
          <HeaderEditor headers={headers} onAdd={addHeader} onRemove={removeHeader}
            onKeyChange={updateHeaderKey} onValueChange={updateHeaderValue} />
        </div>

        <ModalFooter onCancel={onClose} onSubmit={handleSubmit} submitLabel={isEdit ? 'Save Changes' : 'Add Mock'} />
      </div>
    </div>
  )
}

/* ═──────── Sub-components ═──────── */

function ModalHeader({ title, onClose }: { title: string; onClose: () => void }) {
  return (
    <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
      <h2 className="text-lg font-semibold text-gray-900">{title}</h2>
      <button onClick={onClose} className="p-1 text-gray-400 hover:text-gray-600 cursor-pointer">
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
  )
}

function ModalFooter({ onCancel, onSubmit, submitLabel = 'Submit' }: { onCancel: () => void; onSubmit: () => void; submitLabel?: string }) {
  return (
    <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-100">
      <button onClick={onCancel} className="px-5 py-2.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 cursor-pointer">Cancel</button>
      <button onClick={onSubmit} className="px-5 py-2.5 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 cursor-pointer">{submitLabel}</button>
    </div>
  )
}

function LabeledField({ label, children }: { label: string; children: React.ReactNode }) {
  return <div><label className="block text-sm font-medium text-gray-700 mb-1.5">{label}</label>{children}</div>
}

function SegmentedControl({ options, value, onChange }: { options: readonly string[]; value: string; onChange: (v: string) => void }) {
  return (
    <div className="flex gap-2">
      {options.map(opt => (
        <button key={opt} onClick={() => onChange(opt)}
          className={`flex-1 py-2.5 rounded-lg text-sm font-medium transition-colors cursor-pointer ${value === opt ? 'bg-blue-600 text-white shadow' : 'bg-white border border-gray-200 text-gray-600 hover:bg-gray-50'}`}>
          {opt}
        </button>
      ))}
    </div>
  )
}

/** One SSE event card inside the array editor */
function SseEventCard({ evt, onUpdate, onRemove, canRemove }: {
  evt: SSEEvent
  onUpdate: (field: keyof SSEEvent, v: string) => void
  onRemove: (() => void) | undefined
  canRemove: boolean
}) {
  return (
    <div className="p-3 bg-white border border-purple-100 rounded-lg space-y-2 relative">
      {canRemove && (
        <button onClick={onRemove} className="absolute top-2 right-2 p-1 text-purple-300 hover:text-red-500 cursor-pointer">
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      )}
      <div className="grid grid-cols-2 gap-3">
        <LabeledField label="Event">
          <input type="text" value={evt.event} onChange={e => onUpdate('event', e.target.value)}
            placeholder="e.g. message" className={inputCls} />
        </LabeledField>
        <LabeledField label="Delay (ms)">
          <input type="number" value={evt.delay} onChange={e => onUpdate('delay', e.target.value)} min={0} className={inputCls} />
        </LabeledField>
      </div>
      <LabeledField label="Data">
        <textarea value={evt.data} onChange={e => onUpdate('data', e.target.value)} rows={2}
          placeholder='data: {"foo":"bar"}\n\n' className={`w-full px-3 py-2 ${inputCls} font-mono resize-y text-xs`} />
      </LabeledField>
    </div>
  )
}

/** One WebSocket message card inside the array editor */
function WsMessageCard({ msg, onUpdate, onRemove, canRemove }: {
  msg: WebSocketMessage
  onUpdate: (field: keyof WebSocketMessage, v: string) => void
  onRemove: (() => void) | undefined
  canRemove: boolean
}) {
  return (
    <div className="p-3 bg-white border border-green-100 rounded-lg space-y-2 relative">
      {canRemove && (
        <button onClick={onRemove} className="absolute top-2 right-2 p-1 text-green-300 hover:text-red-500 cursor-pointer">
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      )}
      <div className="grid grid-cols-2 gap-3">
        <LabeledField label="Delay (ms)">
          <input type="number" value={msg.delay} onChange={e => onUpdate('delay', e.target.value)} min={0} className={inputCls} />
        </LabeledField>
        <LabeledField label="Type">
          <select value={msg.type} onChange={e => onUpdate('type', e.target.value)}
            className={selectCls}
            style={{ backgroundImage: SELECT_ARROW_BG, backgroundRepeat: 'no-repeat', backgroundPosition: 'right 10px center', backgroundSize: '16px' }}>
            <option value="text">text</option>
            <option value="binary">binary</option>
          </select>
        </LabeledField>
      </div>
      <LabeledField label="Message">
        <textarea value={msg.message} onChange={e => onUpdate('message', e.target.value)} rows={2}
          placeholder='{"type":"welcome","data":"hello"}' className={`w-full px-3 py-2 ${inputCls} font-mono resize-y text-xs`} />
      </LabeledField>
    </div>
  )
}

function HeaderEditor({ headers, onAdd, onRemove, onKeyChange, onValueChange }: {
  headers: HeaderPair[]
  onAdd: () => void
  onRemove: (i: number) => void
  onKeyChange: (i: number, v: string) => void
  onValueChange: (i: number, v: string) => void
}) {
  return (
    <div>
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm font-medium text-gray-700">Response Headers</span>
        <button onClick={onAdd} className="inline-flex items-center gap-1 text-sm font-medium text-blue-600 hover:text-blue-700 cursor-pointer">
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Add Header
        </button>
      </div>
      <div className="space-y-2">
        {headers.map((h, i) => (
          <div key={i} className="flex gap-2">
            <input type="text" value={h.key} onChange={e => onKeyChange(i, e.target.value)} placeholder="header name"
              className="flex-1 px-3.5 py-2.5 bg-white border border-gray-300 rounded-lg text-sm text-gray-800 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:border-transparent" />
            <input type="text" value={h.value} onChange={e => onValueChange(i, e.target.value)} placeholder="value"
              className="flex-1 px-3.5 py-2.5 bg-white border border-gray-300 rounded-lg text-sm text-gray-800 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:border-transparent" />
            <button onClick={() => onRemove(i)} className="p-2.5 text-gray-400 hover:text-red-500 cursor-pointer self-center" title="Remove header">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}
