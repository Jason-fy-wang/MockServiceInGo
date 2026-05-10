import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { RouteTable } from '../RouteTable'
import type { MockRoute } from '../types'

describe('RouteTable', () => {
  const item: MockRoute = {
    key: '1',
    method: 'GET',
    path: '/health',
    requestHeaders: {},
    requestQuery: {},
    responseStatus: 200,
    responseHeaders: {},
    responseBody: '{"message":"ok"}',
  }

  it('renders rows and calls edit handler', () => {
    const onEdit = vi.fn()
    const onDelete = vi.fn()

    render(<RouteTable items={[item]} onEdit={onEdit} onDelete={onDelete} />)

    expect(screen.getByText('/health')).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Edit' }))
    expect(onEdit).toHaveBeenCalledWith(item)
  })
})
