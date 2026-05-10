import { Col, Input, Row, Select, Typography } from 'antd'
import { methodOptions } from './types'

type RouteFiltersProps = {
  filterMethod?: string
  filterPath: string
  onFilterMethodChange: (value: string | undefined) => void
  onFilterPathChange: (value: string) => void
}

export function RouteFilters({
  filterMethod,
  filterPath,
  onFilterMethodChange,
  onFilterPathChange,
}: RouteFiltersProps) {
  return (
    <Row gutter={16} style={{ marginBottom: 24 }}>
      <Col xs={24} md={8}>
        <Typography.Text strong>Filter by Method</Typography.Text>
        <Select
          allowClear
          placeholder="Select method"
          value={filterMethod}
          onChange={onFilterMethodChange}
          options={methodOptions.map((method) => ({ value: method, label: method }))}
          style={{ width: '100%', marginTop: 8 }}
        />
      </Col>
      <Col xs={24} md={8}>
        <Typography.Text strong>Filter by Path</Typography.Text>
        <Input
          placeholder="Enter path substring"
          value={filterPath}
          onChange={(event) => onFilterPathChange(event.target.value)}
          style={{ marginTop: 8 }}
        />
      </Col>
    </Row>
  )
}
