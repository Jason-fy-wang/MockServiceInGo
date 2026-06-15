import { useState } from 'react'
import PageHeader from './components/PageHeader'
import StatCard from './components/StatCard'
import SearchInput from './components/SearchInput'
import MockRow from './components/MockRow'
import AddMockModal from './components/AddMockModal'
import type { MockEndpoint, TabKey, NewMockPayload } from './types/mock'

const initialMocks: MockEndpoint[] = [
  { id: 1, method: 'POST', path: '/v1/sse1', status: 200, type: 'SSE' },
  { id: 2, method: 'GET', path: '/v1/ws1', status: 200, type: 'WebSocket' },
  { id: 3, method: 'POST', path: '/v1/post1', status: 201, type: 'HTTP' },
]

const TABS: TabKey[] = ['All', 'HTTP', 'SSE', 'WebSocket']

export default function App() {
  const [activeTab, setActiveTab] = useState<TabKey>('All')
  const [search, setSearch] = useState('')
  const [mocks, setMocks] = useState<MockEndpoint[]>(initialMocks)
  const [showModal, setShowModal] = useState(false)

  const counts: Record<TabKey, number> = Object.fromEntries(
    TABS.map(t => [t, t === 'All' ? mocks.length : mocks.filter(m => m.type === t).length]),
  ) as Record<TabKey, number>

  const filtered = mocks.filter(m => {
    if (activeTab !== 'All' && m.type !== activeTab) return false
    if (search && !`${m.method} ${m.path}`.toLowerCase().includes(search.toLowerCase())) return false
    return true
  })

  const handleAdd = (newMock: NewMockPayload) =>
    setMocks(prev => [...prev, { ...newMock, id: Date.now() }])
  const handleDelete = (mock: MockEndpoint) =>
    setMocks(prev => prev.filter(m => m.id !== mock.id))

  const handleEdit = (mock: MockEndpoint) => {}

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-5xl mx-auto">
        <PageHeader onAddMock={() => setShowModal(true)} />

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
    </div>
  )
}
