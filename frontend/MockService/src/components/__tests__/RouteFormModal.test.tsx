import { Form } from 'antd'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { RouteFormModal } from '../RouteFormModal'

function TestHost({ onSubmit, onCancel }: { onSubmit: () => void; onCancel: () => void }) {
  const [form] = Form.useForm()
  return <RouteFormModal open editing={false} form={form} onSubmit={onSubmit} onCancel={onCancel} />
}

describe('RouteFormModal', () => {
  it('renders form fields and delegates submit/cancel actions', async () => {
    const user = userEvent.setup()
    const onSubmit = vi.fn()
    const onCancel = vi.fn()

    render(<TestHost onSubmit={onSubmit} onCancel={onCancel} />)

    expect(screen.getByText('Add Route')).toBeInTheDocument()
    expect(screen.getByLabelText('Path')).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Create' }))
    expect(onSubmit).toHaveBeenCalledTimes(1)

    await user.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(onCancel).toHaveBeenCalledTimes(1)
  })
})
