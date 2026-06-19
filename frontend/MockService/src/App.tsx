import { useEffect, useState } from 'react'
import PageHeader from './components/PageHeader'
import StatCard from './components/StatCard'
import SearchInput from './components/SearchInput'
import MockRow from './components/MockRow'
import type { MockEndpoint, TabKey } from './types/mock'
import { listMocks,registerMock,deleteMockByMethod,uploadMockConfig } from './api'
import ImportModal from './components/ImportModel'
import ToastStack, {useToast} from './components/Toast'
import MockModal from './components/MockModal'

const initialMocks: MockEndpoint[] = [
]

const TABS: TabKey[] = ['All', 'HTTP', 'SSE', 'WebSocket']

export default function App() {
  
  const [activeTab, setActiveTab] = useState<TabKey>('All')
  const [search, setSearch] = useState('')
  const [mocks, setMocks] = useState<MockEndpoint[]>(initialMocks)
  const [showModal, setShowModal] = useState(false)
  const [editing, setEditing] = useState<MockEndpoint | null>(null)
  const [showImport, setShowImport] = useState(false)
  const {toasts, toast} = useToast()

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

  const convertMockEndpoints = (data: MockEndpoint[]): MockEndpoint[] => {
    return data?.map((item, index) => ({
          id: `${item.method}-${item.path}-${index}`,
          method: item.method,
          path: item.path,
          requestHeaders: item.requestHeaders ?? [],
          requestBody:
            typeof item.requestBody === 'string' ? item.requestBody : JSON.stringify(item.requestBody ?? ''),
          requestQuery: item.requestQuery ?? [],
          responseStatus: item.responseStatus ?? 200,
          responseHeaders: item.responseHeaders ?? [],
          responseBody:
            typeof item.responseBody === 'string'
              ? item.responseBody
              : JSON.stringify(item.responseBody ?? ''),
          responseType: convertType(item.responseType),
          sseEvents: item.sseEvents ?? [],
          websocketMessages: item.websocketMessages ?? [],
    }))
  }

  useEffect(() => {
    let active = true
    const loadInitialData = async () => {
      try {
        const backendItems = await listMocks()
        if (!active) {
          return
        }
        const mappedItems: MockEndpoint[] = convertMockEndpoints(backendItems.data.mocks)
        console.log('Loaded mocks from backend:', mappedItems)
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

  const IsResponse2xx = (status: number) => {
    return status >= 200 && status < 300
  }

  const IsResponse3xx = (status: number) => {
    return status >= 300 && status < 400
  }

  const IsResponse4xx = (status: number) => {
    return status >= 400 && status < 500
  }

  const IsResponse5xx = (status: number) => {
    return status >= 500 && status < 600
  }

  const counts: Record<TabKey, number> = Object.fromEntries(
    TABS.map(t => [t, t === 'All' ? mocks.length : mocks.filter(m => m.responseType === t).length]),
  ) as Record<TabKey, number>

  const filtered = mocks.filter(m => {
    if (activeTab !== 'All' && m.responseType !== activeTab) return false
    if (search && !`${m.method} ${m.path}`.toLowerCase().includes(search.toLowerCase())) return false
    return true
  })

  const handleAdd = (newMock: MockEndpoint) =>{
    registerMock(newMock).then((res) => {
      if(IsResponse2xx(res.status)){
        newMock.id = `${newMock.method}-${newMock.path}-${Date.now()}`
        setMocks(prev => [...prev, newMock])
        toast({ variant: 'success', text: `Mock endpoint ${newMock.method} ${newMock.path} added` })
      }else{
        toast({ variant: 'error', text: `Failed to add mock endpoint: ${res.data.message}` })
      }
      
    }).catch((error) => {
      toast({ variant: 'error', text: `Failed to add mock endpoint: ${error.message}` })
    })
  }
  const handleDelete = (mock: MockEndpoint) =>{
    deleteMockByMethod(mock.method, mock.path).then((response) => {
      if (IsResponse2xx(response.status)) {
          setMocks(prev => prev.filter(m => m.id !== mock.id))
        toast({ variant: 'success', text: `Mock endpoint ${mock.method} ${mock.path} deleted` })
      }else{
        toast({ variant: 'error', text: `Failed to delete mock endpoint: ${response.data.message}` })
      }
      
    }).catch((error) => {
      toast({ variant: 'error', text: `Failed to delete mock endpoint: ${error.message}` })
    })
    
  }
  const handleImport = (data: MockEndpoint[]) =>{
    const file = new File (
      [JSON.stringify(data)],
      'config.json',
      { type: 'application/json' }
    )
    uploadMockConfig(file).then((response) => {
      if(IsResponse2xx(response.status)){
        data = convertMockEndpoints(data)
        setMocks(prev => [...prev, ...data])
        toast({ variant: 'success', text: 'Mock endpoints imported successfully' })
      } else{
        toast({ variant: 'error', text: `Failed to import mock endpoints: ${response.data.message}` })
      }
    }).catch((error) => {
      toast({ variant: 'error', text: `Failed to import mock endpoints: ${error.message}` })
    })
  }

  const handleEdit = (old: MockEndpoint, mock: MockEndpoint) => {
    handleDelete(old)
    handleAdd(mock)
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
  
  const openEdit = (mock: MockEndpoint) => {
    setEditing(mock)
    setShowModal(true)
  }
  // ── Modal close helper ──
  const closeModal = () => {
    setShowModal(false)
    setEditing(null)
  }

  return ( 
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-5xl mx-auto">
        <PageHeader onImport={() => setShowImport(true)} onAddMock={() => setShowModal(true)} onExport={handleExport} />

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
            <MockRow key={mock.id} mock={mock} onDelete={handleDelete} onEdit={openEdit} />
          ))}
          {filtered.length === 0 && (
            <div className="text-center py-12 text-gray-400 text-sm">
              No mock endpoints found
            </div>
          )}
        </div>
      </div>

      <MockModal
        open={showModal}
        onClose={closeModal}
        onAdd={handleAdd}
        onEdit={handleEdit}
        editing={editing}
      />

      <ImportModal
        open={showImport}
        onClose={() => setShowImport(false)}
        onImport={handleImport}
      />
      <ToastStack toasts={toasts} onClose={()=>{}}/>
    </div>
  )
}
