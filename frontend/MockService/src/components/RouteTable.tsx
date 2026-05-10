import { Button, Popconfirm, Space, Table, Tag } from 'antd'
import { methodOptions, type MockRoute } from './types'

type RouteTableProps = {
  items: MockRoute[]
  onEdit: (record: MockRoute) => void
  onDelete: (key: string) => void
}

export function RouteTable({ items, onEdit, onDelete }: RouteTableProps) {
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
          <Button type="link" onClick={() => onEdit(record)}>
            Edit
          </Button>
          <Popconfirm title="Delete this item?" onConfirm={() => onDelete(record.key)} okText="Yes" cancelText="No">
            <Button type="link" danger>
              Delete
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <Table
      dataSource={items}
      columns={columns}
      rowKey="key"
      pagination={{ pageSize: 6 }}
      locale={{ emptyText: 'No routes found. Add a route to begin.' }}
    />
  )
}
