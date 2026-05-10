import { Button, Col, Row, Typography } from 'antd'

type RoutePageHeaderProps = {
  onAddRoute: () => void
}

export function RoutePageHeader({ onAddRoute }: RoutePageHeaderProps) {
  return (
    <Row justify="space-between" align="middle" style={{ marginBottom: 20 }}>
      <Col>
        <Typography.Title level={3}>Mock Route Manager</Typography.Title>
        <Typography.Paragraph type="secondary">
          Add, query, update, and delete route definitions for mock services.
        </Typography.Paragraph>
      </Col>
      <Col>
        <Button type="primary" onClick={onAddRoute}>
          Add Route
        </Button>
      </Col>
    </Row>
  )
}
