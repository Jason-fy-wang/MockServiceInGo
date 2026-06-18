import { useState } from 'react'
import { HTTP_METHODS, STATUS_CODE_PRESETS, RESPONSE_TYPES } from '../constants/mock'
import type { HeaderPair, MockEndpoint } from '../types/mock'

const SELECT_ARROW_BG = `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%236b7280'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'%3E%3C/path%3E%3C/svg%3E")`

const selectCls = `w-full px-3.5 py-2.5 bg-white border border-gray-300 rounded-lg text-sm text-gray-800 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:border-transparent appearance-none cursor-pointer`
const inputCls = `w-full px-3.5 py-2.5 bg-white border border-gray-300 rounded-lg text-sm text-gray-800 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:border-transparent`

interface AddMockModalProps {
  open: boolean
  onClose: () => void
  onAdd: (payload: MockEndpoint) => void
}

/**
 * AddMockModal — modal dialog for creating a new mock endpoint.
 */
export default function AddMockModal({ open, onClose, onAdd }: AddMockModalProps) {
  const [method, setMethod] = useState<string>('GET')
  const [path, setPath] = useState<string>('/v1/')
  const [statusCodePreset, setStatusCodePreset] = useState<string>('200')
  const [statusCodeInput, setStatusCodeInput] = useState<string>('200')
  const [responseType, setResponseType] = useState<string>('Http')
  const [body, setBody] = useState<string>('{ "key": "value" }')
  const [headers, setHeaders] = useState<HeaderPair[]>([
    { key: 'content-type', value: 'application/json' },
  ])

  if (!open) return null

  const handleAddHeader = () => setHeaders([...headers, { key: '', value: '' }])
  const removeHeader = (i: number) => setHeaders(headers.filter((_, idx) => idx !== i))
  const updateHeaderKey = (i: number, v: string) =>
    setHeaders(headers.map((h, idx) => idx === i ? { ...h, key: v } : h))
  const updateHeaderValue = (i: number, v: string) =>
    setHeaders(headers.map((h, idx) => idx === i ? { ...h, value: v } : h))

  const handleSubmit = () => {
    const typeMap: Record<string, MockEndpoint['responseType']> = {
      Http: 'HTTP', Sse: 'SSE', WebSocket: 'WebSocket',
    }
    onAdd({
      id: Date.now(),
      method: method as MockEndpoint['method'],
      path,
      responseStatus: parseInt(statusCodeInput),
      responseType: typeMap[responseType] ?? 'HTTP',
      requestHeaders: headers.reduce((acc, { key, value }) => ({ ...acc, [key]: value }), {}),
      responseBody: body,
    })
    onClose()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-16">
      <div className="absolute inset-0 bg-black/20" onClick={onClose} />
      <div className="relative w-full max-w-2xl mx-4 bg-white rounded-2xl shadow-xl">
        <ModalHeader title="Add Mock Service" onClose={onClose} />

        <div className="px-6 py-5 space-y-5 max-h-[65vh] overflow-y-auto">
          {/* Row 1: Method + Path */}
          <div className="grid grid-cols-2 gap-4">
            <LabeledField label="Method">
              <select
                value={method}
                onChange={e => setMethod(e.target.value)}
                className={selectCls}
                style={{ backgroundImage: SELECT_ARROW_BG, backgroundRepeat: 'no-repeat', backgroundPosition: 'right 10px center', backgroundSize: '16px' }}
              >
                {HTTP_METHODS.map(m => <option key={m} value={m}>{m}</option>)}
              </select>
            </LabeledField>
            <LabeledField label="Path">
              <input
                type="text"
                value={path}
                onChange={e => setPath(e.target.value)}
                className={`${inputCls} font-mono`}
              />
            </LabeledField>
          </div>

          {/* Row 2: Status Code (preset + custom) + Response Type */}
          <div className="grid grid-cols-[120px_140px_1fr] gap-4 items-start">
            <LabeledField label="Status Code">
              <select
                value={statusCodePreset}
                onChange={e => { setStatusCodePreset(e.target.value); setStatusCodeInput(e.target.value) }}
                className={selectCls}
                style={{ backgroundImage: SELECT_ARROW_BG, backgroundRepeat: 'no-repeat', backgroundPosition: 'right 10px center', backgroundSize: '16px' }}
              >
                {STATUS_CODE_PRESETS.map(s => <option key={s} value={s}>{s}</option>)}
              </select>
            </LabeledField>
            <div>
              <label className="block text-sm font-medium text-transparent mb-1.5">&nbsp;</label>
              <input
                type="text"
                value={statusCodeInput}
                onChange={e => setStatusCodeInput(e.target.value)}
                className={inputCls}
              />
            </div>
            <LabeledField label="Response Type">
              <SegmentedControl
                options={[...RESPONSE_TYPES]}
                value={responseType}
                onChange={setResponseType}
              />
            </LabeledField>
          </div>

          {/* Row 3: Response Body */}
          <LabeledField label="Response Body">
            <textarea
              value={body}
              onChange={e => setBody(e.target.value)}
              rows={5}
              className={`w-full px-3.5 py-3 ${inputCls} font-mono resize-y`}
            />
          </LabeledField>

          {/* Row 4: Response Headers */}
          <HeaderEditor
            headers={headers}
            onAdd={handleAddHeader}
            onRemove={removeHeader}
            onKeyChange={updateHeaderKey}
            onValueChange={updateHeaderValue}
          />
        </div>

        <ModalFooter onCancel={onClose} onSubmit={handleSubmit} submitLabel="Add Mock" />
      </div>
    </div>
  )
}

/* ───────── Sub-components ───────── */

interface ModalHeaderProps {
  title: string
  onClose: () => void
}

function ModalHeader({ title, onClose }: ModalHeaderProps) {
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

interface ModalFooterProps {
  onCancel: () => void
  onSubmit: () => void
  submitLabel?: string
}

function ModalFooter({ onCancel, onSubmit, submitLabel = 'Submit' }: ModalFooterProps) {
  return (
    <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-100">
      <button
        onClick={onCancel}
        className="px-5 py-2.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 cursor-pointer"
      >
        Cancel
      </button>
      <button
        onClick={onSubmit}
        className="px-5 py-2.5 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 cursor-pointer"
      >
        {submitLabel}
      </button>
    </div>
  )
}

interface LabeledFieldProps {
  label: string
  children: React.ReactNode
}

function LabeledField({ label, children }: LabeledFieldProps) {
  return (
    <div>
      <label className="block text-sm font-medium text-gray-700 mb-1.5">{label}</label>
      {children}
    </div>
  )
}

interface SegmentedControlProps {
  options: readonly string[]
  value: string
  onChange: (value: string) => void
}

function SegmentedControl({ options, value, onChange }: SegmentedControlProps) {
  return (
    <div className="flex gap-2">
      {options.map(opt => (
        <button
          key={opt}
          onClick={() => onChange(opt)}
          className={`flex-1 py-2.5 rounded-lg text-sm font-medium transition-colors cursor-pointer ${
            value === opt
              ? 'bg-blue-600 text-white shadow'
              : 'bg-white border border-gray-200 text-gray-600 hover:bg-gray-50'
          }`}
        >
          {opt}
        </button>
      ))}
    </div>
  )
}

interface HeaderEditorProps {
  headers: HeaderPair[]
  onAdd: () => void
  onRemove: (i: number) => void
  onKeyChange: (i: number, v: string) => void
  onValueChange: (i: number, v: string) => void
}

function HeaderEditor({ headers, onAdd, onRemove, onKeyChange, onValueChange }: HeaderEditorProps) {
  return (
    <div>
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm font-medium text-gray-700">Response Headers</span>
        <button
          onClick={onAdd}
          className="inline-flex items-center gap-1 text-sm font-medium text-blue-600 hover:text-blue-700 cursor-pointer"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Add Header
        </button>
      </div>
      <div className="space-y-2">
        {headers.map((h, i) => (
          <div key={i} className="flex gap-2">
            <input
              type="text"
              value={h.key}
              onChange={e => onKeyChange(i, e.target.value)}
              placeholder="header name"
              className="flex-1 px-3.5 py-2.5 bg-white border border-gray-300 rounded-lg text-sm text-gray-800 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:border-transparent"
            />
            <input
              type="text"
              value={h.value}
              onChange={e => onValueChange(i, e.target.value)}
              placeholder="value"
              className="flex-1 px-3.5 py-2.5 bg-white border border-gray-300 rounded-lg text-sm text-gray-800 placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:border-transparent"
            />
            <button
              onClick={() => onRemove(i)}
              className="p-2.5 text-gray-400 hover:text-red-500 cursor-pointer self-center"
              title="Remove header"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                  d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        ))}
      </div>
    </div>
  )
}
