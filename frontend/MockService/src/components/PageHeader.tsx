interface PageHeaderProps {
  onImport?: () => void
  onExport?: () => void
  onAddMock?: () => void
}

/**
 * PageHeader — top bar with title, subtitle, and action buttons.
 */
export default function PageHeader({ onImport, onExport, onAddMock }: PageHeaderProps) {
  return (
    <div className="flex items-start justify-between mb-2">
      <div>
        <h1 className="text-3xl font-semibold text-gray-900">Mock Service Manager</h1>
        <p className="text-sm text-gray-500 mt-0.5">
          Manage HTTP, SSE, and WebSocket mock endpoints
        </p>
      </div>
      <div className="flex gap-2">
        <HeaderButton onClick={onImport}>
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
          </svg>
          Import JSON
        </HeaderButton>
        <HeaderButton onClick={onExport} variant="default">
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
          </svg>
          Export JSON
        </HeaderButton>
        <HeaderButton onClick={onAddMock} variant="primary">
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          Add Mock
        </HeaderButton>
      </div>
    </div>
  )
}

interface HeaderButtonProps {
  children: React.ReactNode
  onClick?: () => void
  variant?: 'default' | 'primary'
}

function HeaderButton({ children, onClick, variant = 'default' }: HeaderButtonProps) {
  const base = 'flex items-center gap-1.5 px-4 py-2 text-sm font-medium rounded-lg cursor-pointer transition-colors'
  const styles = {
    default: 'text-gray-700 bg-white border border-gray-300 hover:bg-gray-50',
    primary: 'text-white bg-blue-600 hover:bg-blue-700',
  }
  return (
    <button onClick={onClick} className={`${base} ${styles[variant]}`}>
      {children}
    </button>
  )
}
