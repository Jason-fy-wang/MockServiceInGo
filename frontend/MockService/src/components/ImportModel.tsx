import { useState, useRef } from 'react'
import type { MockEndpoint } from '../types/mock'

interface ImportModalProps {
  open: boolean
  onClose: () => void
  onImport: (data: MockEndpoint[]) => void
}

/**
 * ImportModal — file picker modal for importing mock endpoints from JSON.
 */
export default function ImportModal({ open, onClose, onImport }: ImportModalProps) {
  const [fileName, setFileName] = useState('')
  const [error, setError] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  if (!open) return null

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    setFileName(file.name)
    setError('')

    const reader = new FileReader()
    reader.onload = () => {
      try {
        const json = JSON.parse(reader.result as string)
        if (!Array.isArray(json)) throw new Error('JSON root must be an array')
        onImport(json as MockEndpoint[])
        onClose()
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Invalid JSON')
      }
      // Reset so re-selecting same file works again
      if (inputRef.current) inputRef.current.value = ''
    }
    reader.readAsText(file)
  }

  const handleDrop = (e: React.DragEvent<HTMLLabelElement>) => {
    e.preventDefault()
    const file = e.dataTransfer.files?.[0]
    if (!file || !file.name.endsWith('.json')) {
      setError('Please drop a .json file')
      return
    }
    setFileName(file.name)
    setError('')

    const reader = new FileReader()
    reader.onload = () => {
      try {
        const json = JSON.parse(reader.result as string)
        if (!Array.isArray(json)) throw new Error('JSON root must be an array')
        onImport(json as MockEndpoint[])
        onClose()
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Invalid JSON')
      }
    }
    reader.readAsText(file)
  }

  const handleDragOver = (e: React.DragEvent<HTMLLabelElement>) => {
    e.preventDefault()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/20" onClick={onClose} />

      <div className="relative w-full max-w-md bg-white rounded-2xl shadow-xl">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-5 border-b border-gray-100">
          <h2 className="text-lg font-semibold text-gray-900">Import JSON</h2>
          <button onClick={onClose} className="p-1 text-gray-400 hover:text-gray-600 cursor-pointer">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Body */}
        <div className="px-6 py-8 space-y-4">
          {/* Drop zone / click area */}
          <label
            className={`flex flex-col items-center justify-center gap-3 rounded-xl border-2 border-dashed cursor-pointer transition-colors ${
              error ? 'border-red-300 bg-red-50' : 'border-gray-300 hover:border-blue-400 hover:bg-blue-50/50'
            }`}
            style={{ minHeight: '180px', padding: '32px' }}
            onDrop={handleDrop}
            onDragOver={handleDragOver}
          >
            <input
              ref={inputRef}
              type="file"
              accept=".json,application/json"
              onChange={handleFileChange}
              className="hidden"
            />

            {/* Upload icon circle */}
            <div className={`w-14 h-14 rounded-full flex items-center justify-center ${error ? 'bg-red-100' : 'bg-gray-100'}`}>
              <svg className={`w-7 h-7 ${error ? 'text-red-400' : 'text-gray-400'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M12 11V4m0 0l-4 4m4-4l4 4" />
              </svg>
            </div>

            <div className="text-center">
              <p className="text-sm font-medium text-gray-700">Select a JSON file</p>
              <p className="text-xs text-gray-400 mt-0.5">Click to browse</p>
            </div>

            {fileName && !error && (
              <span className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-green-50 text-green-700 text-xs font-medium">
                <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
                {fileName}
              </span>
            )}
          </label>

          {/* Warning text */}
          <p className="text-xs text-gray-400 text-center">
            Warning: importing will replace all existing mock services.
          </p>

          {/* Error message */}
          {error && (
            <p className="text-xs text-red-500 text-center">{error}</p>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-100">
          <button
            onClick={onClose}
            className="px-5 py-2.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 cursor-pointer"
          >
            Cancel
          </button>
          <button
            onClick={() => inputRef.current?.click()}
            disabled={!!fileName && !error}
            className="px-5 py-2.5 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer"
          >
            Import
          </button>
        </div>
      </div>
    </div>
  )
}
