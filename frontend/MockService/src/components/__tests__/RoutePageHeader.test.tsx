import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { RoutePageHeader } from '../RoutePageHeader'

describe('RoutePageHeader', () => {
  it('renders title and triggers add callback', async () => {
    const user = userEvent.setup()
    const onAddRoute = vi.fn()

    render(<RoutePageHeader onAddRoute={onAddRoute} />)

    expect(screen.getByText('Mock Route Manager')).toBeInTheDocument()
    await user.click(screen.getByRole('button', { name: 'Add Route' }))
    expect(onAddRoute).toHaveBeenCalledTimes(1)
  })
})
