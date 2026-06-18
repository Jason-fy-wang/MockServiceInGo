import { useEffect, useState } from 'react'
import PageHeader from './components/PageHeader'
import StatCard from './components/StatCard'
import SearchInput from './components/SearchInput'
import MockRow from './components/MockRow'
import AddMockModal from './components/AddMockModal'
import type { MockEndpoint, TabKey } from './types/mock'
import { listMocks,registerMock,deleteMockByMethod } from './api'
import ImportModal from './components/ImportModel'

const initialMocks: MockEndpoint[] = [
  // { id: 1, method: 'POST', path: '/v1/sse1', responseStatus: 200, responseType: 'SSE', sseEvents: [], websocketMessages: [] },
  // { id: 2, method: 'GET', path: '/v1/ws1', responseStatus: 200, responseType: 'WebSocket', sseEvents: [], websocketMessages: [] },
  // { id: 3, method: 'POST', path: '/v1/post1', responseStatus: 201, responseType: 'HTTP', sseEvents: [], websocketMessages: [] },
]

const TABS: TabKey[] = ['All', 'HTTP', 'SSE', 'WebSocket']

export default function App() {
  
  const [activeTab, setActiveTab] = useState<TabKey>('All')
  const [search, setSearch] = useState('')
  const [mocks, setMocks] = useState<MockEndpoint[]>(initialMocks)
  const [showModal, setShowModal] = useState(false)
  const [showImport, setShowImport] = useState(false)

  const convertType = (type: string): 'HTTP' | 'SSE' | 'WebSocket' => {
    switch (type.toLowerCase()) {
      case 'http':
        return 'HTTP'
      case 'sse':
        return 'SSE'
      case 'websocket':
        return 'WebSocket'
      default:
        throw new Error(`Unhandled type: ${type}`)
    }
  }

  useEffect(() => {
    let active = true

    const loadInitialData = async () => {
      try {
        const backendItems = await listMocks()
        if (!active) {
          return
        }

        const mappedItems: MockEndpoint[] = backendItems.mocks?.map((item, index) => ({
          id: `${item.method}-${item.path}-${index}`,
          method: item.method,
          path: item.path,
          requestHeaders: item.requestHeaders ?? {},
          requestBody:
            typeof item.requestBody === 'string' ? item.requestBody : JSON.stringify(item.requestBody ?? ''),
          requestQuery: item.requestQuery ?? {},
          responseStatus: item.responseStatus ?? 200,
          responseHeaders: item.responseHeaders ?? {},
          responseBody:
            typeof item.responseBody === 'string'
              ? item.responseBody
              : JSON.stringify(item.responseBody ?? ''),
          responseType: convertType(item.responseType),
          sseEvents: item.sseEvents ?? [],
          websocketMessages: item.websocketMessages ?? [],
        }))

        setMocks(mappedItems)
      } catch (error) {
        if (!active) {
          return
        }
      }
    }

    loadInitialData()
    return () => {
      active = false
    }
  }, [])

  const counts: Record<TabKey, number> = Object.fromEntries(
    TABS.map(t => [t, t === 'All' ? mocks.length : mocks.filter(m => m.responseType === t).length]),
  ) as Record<TabKey, number>

  const filtered = mocks.filter(m => {
    if (activeTab !== 'All' && m.responseType !== activeTab) return false
    if (search && !`${m.method} ${m.path}`.toLowerCase().includes(search.toLowerCase())) return false
    return true
  })

  const handleAdd = (newMock: MockEndpoint) =>{
    registerMock(newMock)
    setMocks(prev => [...prev, { ...newMock, id: Date.now() }])
  }
  const handleDelete = (mock: MockEndpoint) =>{
    deleteMockByMethod(mock.method, mock.path)
    setMocks(prev => prev.filter(m => m.id !== mock.id))
  }
  const handleImport = (data: MockEndpoint[]) =>
    setMocks(data.map((m, i) => ({ ...m, id: Date.now() + i })))
  const handleEdit = (mock: MockEndpoint) => {

  }

    // ── Export JSON (direct download) ──
  const handleExport = () => {
    // Strip internal id field for clean export
    const exportData = mocks.map(({ id, ...rest }) => rest)
    const json = JSON.stringify(exportData, null, 2)
    const blob = new Blob([json], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `mocks_${new Date().toISOString().slice(0, 10)}.json`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }


  return ( 
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-5xl mx-auto">
        <PageHeader onImport={() => setShowImport(true)} onAddMock={() => setShowModal(true)} />

        {/* Tab cards */}
        <div className="grid grid-cols-4 gap-4 mt-6">
          {TABS.map(tab => (
            <StatCard
              key={tab}
              label={tab}
              count={counts[tab]}
              active={activeTab === tab}
              onClick={() => setActiveTab(tab)}
            />
          ))}
        </div>

        {/* Search */}
        <div className="mt-6">
          <SearchInput
            value={search}
            onChange={setSearch}
            placeholder="Search by path or method..."
          />
        </div>

        {/* List */}
        <div className="mt-4 space-y-2">
          {filtered.map(mock => (
            <MockRow key={mock.id} mock={mock} onDelete={handleDelete} onEdit={handleEdit} />
          ))}
          {filtered.length === 0 && (
            <div className="text-center py-12 text-gray-400 text-sm">
              No mock endpoints found
            </div>
          )}
        </div>
      </div>

      <AddMockModal
        open={showModal}
        onClose={() => setShowModal(false)}
        onAdd={handleAdd}
      />

      <ImportModal
        open={showImport}
        onClose={() => setShowImport(false)}
        onImport={handleImport}
      />
    </div>
  )
}
