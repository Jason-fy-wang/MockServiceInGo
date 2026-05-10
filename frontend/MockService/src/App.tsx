import { useEffect, useMemo, useState } from 'react'
import { Form, message } from 'antd'
import 'antd/dist/reset.css'
import './App.css'
import { listMocks, registerMock } from './api'
import { RouteFilters } from './components/RouteFilters'
import { RouteFormModal } from './components/RouteFormModal'
import { RoutePageHeader } from './components/RoutePageHeader'
import { RouteTable } from './components/RouteTable'
import { defaultRoute, type MockRoute } from './components/types'

function formListFromMap(map: Record<string, string> | undefined) {
  return Object.entries(map ?? {}).map(([name, value]) => ({ name, value }))
}

function mapFromFormList(list: { name?: string; value?: string }[] | undefined) {
  return (list ?? []).reduce<Record<string, string>>((acc, entry) => {
    if (entry?.name) {
      acc[entry.name] = entry.value ?? ''
    }
    return acc
  }, {})
}

function App() {
  const [items, setItems] = useState<MockRoute[]>([])
  const [filterMethod, setFilterMethod] = useState<string | undefined>(undefined)
  const [filterPath, setFilterPath] = useState('')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [editingKey, setEditingKey] = useState<string | null>(null)
  const [form] = Form.useForm()
  const [messageApi, contextHolder] = message.useMessage()

  useEffect(() => {
    let active = true

    const loadInitialData = async () => {
      try {
        const backendItems = await listMocks()
        if (!active) {
          return
        }

        const mappedItems: MockRoute[] = backendItems.mocks?.map((item, index) => ({
          key: `${item.method}-${item.path}-${index}`,
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
        }))

        setItems(mappedItems)
      } catch (error) {
        if (!active) {
          return
        }
        // console.error('Failed to load mocks:', error)
        messageApi.error('Failed to load routes from backend')
      }
    }

    loadInitialData()

    return () => {
      active = false
    }
  }, [messageApi])

  const filteredItems = useMemo(
    () =>
      items.filter((item) => {
        const matchesMethod = filterMethod ? item.method === filterMethod : true
        const matchesPath = filterPath ? item.path.includes(filterPath.trim()) : true
        return matchesMethod && matchesPath
      }),
    [items, filterMethod, filterPath]
  )

  const openCreate = () => {
    form.setFieldsValue({
      ...defaultRoute,
      requestHeaders: [],
      requestQuery: [],
      responseHeaders: [],
    })
    setEditingKey(null)
    setIsModalOpen(true)
  }

  const openEdit = (record: MockRoute) => {
    form.setFieldsValue({
      ...record,
      requestHeaders: formListFromMap(record.requestHeaders),
      requestQuery: formListFromMap(record.requestQuery),
      responseHeaders: formListFromMap(record.responseHeaders),
    })
    setEditingKey(record.key)
    setIsModalOpen(true)
  }

  const handleDelete = (key: string) => {
    setItems((current) => current.filter((item) => item.key !== key))
  }

  const handleSubmit = async () => {
    const values = await form.validateFields()

    const nextItem: MockRoute = {
      key: editingKey ?? `${Date.now()}-${Math.random().toString(36).slice(2)}`,
      method: values.method,
      path: values.path,
      requestHeaders: mapFromFormList(values.requestHeaders),
      requestBody: values.requestBody || '',
      requestQuery: mapFromFormList(values.requestQuery),
      responseStatus: values.responseStatus,
      responseHeaders: mapFromFormList(values.responseHeaders),
      responseBody: values.responseBody || '',
    }

    if (!editingKey) {
      try {
        await registerMock({
          method: nextItem.method,
          path: nextItem.path,
          requestHeaders: nextItem.requestHeaders,
          requestBody: nextItem.requestBody,
          requestQuery: nextItem.requestQuery,
          responseStatus: nextItem.responseStatus,
          responseHeaders: nextItem.responseHeaders,
          responseBody: nextItem.responseBody,
        })
      } catch (error) {
        // console.error('Failed to register mock:', error)
        messageApi.error('Failed to create route in backend')
        return
      }
    }

    setItems((current) => {
      if (editingKey) {
        return current.map((item) => (item.key === editingKey ? nextItem : item))
      }
      return [...current, nextItem]
    })
    setIsModalOpen(false)
  }

  return (
    <div className="app-container" style={{ padding: 24, minHeight: '100vh' }}>
      {contextHolder}
      <RoutePageHeader onAddRoute={openCreate} />

      <RouteFilters
        filterMethod={filterMethod}
        filterPath={filterPath}
        onFilterMethodChange={setFilterMethod}
        onFilterPathChange={setFilterPath}
      />

      <RouteTable items={filteredItems} onEdit={openEdit} onDelete={handleDelete} />

      <RouteFormModal
        open={isModalOpen}
        editing={Boolean(editingKey)}
        form={form}
        onSubmit={handleSubmit}
        onCancel={() => setIsModalOpen(false)}
      />
    </div>
  )
}

export default App
