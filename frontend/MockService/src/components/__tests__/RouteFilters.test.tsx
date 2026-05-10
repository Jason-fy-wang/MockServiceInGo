import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { RouteFilters } from '../RouteFilters'

describe('RouteFilters', () => {
  it('updates path filter through callback', () => {
    const onFilterMethodChange = vi.fn()
    const onFilterPathChange = vi.fn()

    render(
      <RouteFilters
        filterMethod={undefined}
        filterPath=""
        onFilterMethodChange={onFilterMethodChange}
        onFilterPathChange={onFilterPathChange}
      />
    )

    fireEvent.change(screen.getByPlaceholderText('Enter path substring'), {
      target: { value: '/api' },
    })

    expect(onFilterPathChange).toHaveBeenCalledWith('/api')
  })
})
