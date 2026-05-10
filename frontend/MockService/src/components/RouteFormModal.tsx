import { Button, Col, Form, Input, InputNumber, Modal, Row, Select, Space, Typography } from 'antd'
import type { FormInstance } from 'antd'
import { methodOptions } from './types'

type RouteFormModalProps = {
  open: boolean
  editing: boolean
  form: FormInstance
  onSubmit: () => void
  onCancel: () => void
}

export function RouteFormModal({ open, editing, form, onSubmit, onCancel }: RouteFormModalProps) {
  return (
    <Modal
      title={editing ? 'Edit Route' : 'Add Route'}
      open={open}
      onOk={onSubmit}
      onCancel={onCancel}
      okText={editing ? 'Save' : 'Create'}
      width={800}
    >
      <Form form={form} layout="vertical" initialValues={{ responseStatus: 200 }}>
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item name="method" label="HTTP Method" rules={[{ required: true, message: 'Method is required' }]}>
              <Select options={methodOptions.map((method) => ({ value: method, label: method }))} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item name="path" label="Path" rules={[{ required: true, message: 'Path is required' }]}>
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
                    key={`${field.key}-name`}
                    name={[field.name, 'name']}
                    rules={[{ required: true, message: 'Header key required' }]}
                  >
                    <Input placeholder="Header Name" />
                  </Form.Item>
                  <Form.Item key={`${field.key}-value`} name={[field.name, 'value']}>
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
                    key={`${field.key}-name`}
                    name={[field.name, 'name']}
                    rules={[{ required: true, message: 'Query key required' }]}
                  >
                    <Input placeholder="Query Name" />
                  </Form.Item>
                  <Form.Item key={`${field.key}-value`} name={[field.name, 'value']}>
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
                    key={`${field.key}-name`}
                    name={[field.name, 'name']}
                    rules={[{ required: true, message: 'Header key required' }]}
                  >
                    <Input placeholder="Header Name" />
                  </Form.Item>
                  <Form.Item key={`${field.key}-value`} name={[field.name, 'value']}>
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
  )
}
