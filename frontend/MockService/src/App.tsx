import { useMemo, useState } from 'react'
import {
  Button,
  Col,
  Form,
  Input,
  InputNumber,
  Modal,
  Popconfirm,
  Row,
  Select,
  Space,
  Table,
  Tag,
  Typography,
} from 'antd'
import 'antd/dist/reset.css'
import './App.css'

interface MockRoute {
  key: string
  method: string
  path: string
  requestHeaders: Record<string, string>
  requestBody?: string
  requestQuery: Record<string, string>
  responseStatus?: number
  responseHeaders: Record<string, string>
  responseBody?: string
}

const methodOptions = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']

const defaultRoute: Omit<MockRoute, 'key'> = {
  method: 'GET',
  path: '/example',
  requestHeaders: {},
  requestBody: '',
  requestQuery: {},
  responseStatus: 200,
  responseHeaders: {},
  responseBody: '{"message":"ok"}',
}

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

    setItems((current) => {
      if (editingKey) {
        return current.map((item) => (item.key === editingKey ? nextItem : item))
      }
      return [...current, nextItem]
    })
    setIsModalOpen(false)
  }

  const columns = [
    {
      title: 'Method',
      dataIndex: 'method',
      key: 'method',
      render: (method: string) => <Tag color="blue">{method}</Tag>,
      filters: methodOptions.map((method) => ({ text: method, value: method })),
      onFilter: (value: any, record: MockRoute) => record.method === String(value),
    },
    {
      title: 'Path',
      dataIndex: 'path',
      key: 'path',
    },
    {
      title: 'Status',
      dataIndex: 'responseStatus',
      key: 'responseStatus',
      render: (status: number) => <Tag color={status >= 200 && status < 300 ? 'green' : 'volcano'}>{status}</Tag>,
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: unknown, record: MockRoute) => (
        <Space>
          <Button type="link" onClick={() => openEdit(record)}>
            Edit
          </Button>
          <Popconfirm
            title="Delete this item?"
            onConfirm={() => handleDelete(record.key)}
            okText="Yes"
            cancelText="No"
          >
            <Button type="link" danger>
              Delete
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div className="app-container" style={{ padding: 24, minHeight: '100vh' }}>
      <Row justify="space-between" align="middle" style={{ marginBottom: 20 }}>
        <Col>
          <Typography.Title level={3}>Mock Route Manager</Typography.Title>
          <Typography.Paragraph type="secondary">
            Add, query, update, and delete route definitions for mock services.
          </Typography.Paragraph>
        </Col>
        <Col>
          <Button type="primary" onClick={openCreate}>
            Add Route
          </Button>
        </Col>
      </Row>

      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col xs={24} md={8}>
          <Typography.Text strong>Filter by Method</Typography.Text>
          <Select
            allowClear
            placeholder="Select method"
            value={filterMethod}
            onChange={setFilterMethod}
            options={methodOptions.map((method) => ({ value: method, label: method }))}
            style={{ width: '100%', marginTop: 8 }}
          />
        </Col>
        <Col xs={24} md={8}>
          <Typography.Text strong>Filter by Path</Typography.Text>
          <Input
            placeholder="Enter path substring"
            value={filterPath}
            onChange={(event) => setFilterPath(event.target.value)}
            style={{ marginTop: 8 }}
          />
        </Col>
      </Row>

      <Table
        dataSource={filteredItems}
        columns={columns}
        rowKey="key"
        pagination={{ pageSize: 6 }}
        locale={{ emptyText: 'No routes found. Add a route to begin.' }}
      />

      <Modal
        title={editingKey ? 'Edit Route' : 'Add Route'}
        open={isModalOpen}
        onOk={handleSubmit}
        onCancel={() => setIsModalOpen(false)}
        okText={editingKey ? 'Save' : 'Create'}
        width={800}
      >
        <Form form={form} layout="vertical" initialValues={{ responseStatus: 200 }}>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="method"
                label="HTTP Method"
                rules={[{ required: true, message: 'Method is required' }]}
              >
                <Select options={methodOptions.map((method) => ({ value: method, label: method }))} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="path"
                label="Path"
                rules={[{ required: true, message: 'Path is required' }]}
              >
                <Input placeholder="/api/items" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="responseStatus" label="Response Status">
                <InputNumber min={100} max={599} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="requestBody" label="Request Body">
                <Input.TextArea rows={3} placeholder='{"key":"value"}' />
              </Form.Item>
            </Col>
          </Row>

          <Form.List name="requestHeaders">
            {(fields, { add, remove }) => (
              <div style={{ marginBottom: 16 }}>
                <Typography.Text strong>Request Headers</Typography.Text>
                {fields.map((field) => (
                  <Space key={field.key} align="baseline" style={{ display: 'flex', marginBottom: 8 }}>
                    <Form.Item
                      {...field}
                      name={[field.name, 'name']}
                      rules={[{ required: true, message: 'Header key required' }]}
                    >
                      <Input placeholder="Header Name" />
                    </Form.Item>
                    <Form.Item
                      {...field}
                      name={[field.name, 'value']}
                    >
                      <Input placeholder="Header Value" />
                    </Form.Item>
                    <Button type="link" danger onClick={() => remove(field.name)}>
                      Remove
                    </Button>
                  </Space>
                ))}
                <Form.Item>
                  <Button type="dashed" block onClick={() => add()}>
                    Add Request Header
                  </Button>
                </Form.Item>
              </div>
            )}
          </Form.List>

          <Form.List name="requestQuery">
            {(fields, { add, remove }) => (
              <div style={{ marginBottom: 16 }}>
                <Typography.Text strong>Request Query</Typography.Text>
                {fields.map((field) => (
                  <Space key={field.key} align="baseline" style={{ display: 'flex', marginBottom: 8 }}>
                    <Form.Item
                      {...field}
                      name={[field.name, 'name']}
                      rules={[{ required: true, message: 'Query key required' }]}
                    >
                      <Input placeholder="Query Name" />
                    </Form.Item>
                    <Form.Item
                      {...field}
                      name={[field.name, 'value']}
                    >
                      <Input placeholder="Query Value" />
                    </Form.Item>
                    <Button type="link" danger onClick={() => remove(field.name)}>
                      Remove
                    </Button>
                  </Space>
                ))}
                <Form.Item>
                  <Button type="dashed" block onClick={() => add()}>
                    Add Query Parameter
                  </Button>
                </Form.Item>
              </div>
            )}
          </Form.List>

          <Form.List name="responseHeaders">
            {(fields, { add, remove }) => (
              <div style={{ marginBottom: 16 }}>
                <Typography.Text strong>Response Headers</Typography.Text>
                {fields.map((field) => (
                  <Space key={field.key} align="baseline" style={{ display: 'flex', marginBottom: 8 }}>
                    <Form.Item
                      {...field}
                      name={[field.name, 'name']}
                      rules={[{ required: true, message: 'Header key required' }]}
                    >
                      <Input placeholder="Header Name" />
                    </Form.Item>
                    <Form.Item
                      {...field}
                      name={[field.name, 'value']}
                    >
                      <Input placeholder="Header Value" />
                    </Form.Item>
                    <Button type="link" danger onClick={() => remove(field.name)}>
                      Remove
                    </Button>
                  </Space>
                ))}
                <Form.Item>
                  <Button type="dashed" block onClick={() => add()}>
                    Add Response Header
                  </Button>
                </Form.Item>
              </div>
            )}
          </Form.List>

          <Form.Item name="responseBody" label="Response Body">
            <Input.TextArea rows={4} placeholder='{"message":"ok"}' />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default App
