import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { OPERATIONS } from '../api/operations'
import { OperationPicker } from './OperationPicker'

describe('OperationPicker', () => {
  it('offers every operation', () => {
    render(<OperationPicker value="add" onChange={vi.fn()} />)

    expect(screen.getAllByRole('radio')).toHaveLength(OPERATIONS.length)
    expect(
      screen.getByRole('radio', { name: 'Square Root' }),
    ).toBeInTheDocument()
  })

  it('marks only the selected operation', () => {
    render(<OperationPicker value="divide" onChange={vi.fn()} />)

    expect(screen.getByRole('radio', { name: 'Divide' })).toBeChecked()
    expect(screen.getByRole('radio', { name: 'Add' })).not.toBeChecked()
  })

  it('reports the operation that is clicked', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<OperationPicker value="add" onChange={onChange} />)

    await user.click(screen.getByRole('radio', { name: 'Multiply' }))

    expect(onChange).toHaveBeenCalledWith('multiply')
  })

  it('is a group named Operation', () => {
    render(<OperationPicker value="add" onChange={vi.fn()} />)

    expect(screen.getByRole('group', { name: 'Operation' })).toBeInTheDocument()
  })
})
