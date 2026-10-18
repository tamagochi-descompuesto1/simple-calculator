import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { NumberField } from './NumberField'

describe('NumberField', () => {
  it('shows its label and value', () => {
    render(<NumberField label="First number" value="12" onChange={vi.fn()} />)

    expect(screen.getByLabelText('First number')).toHaveValue('12')
  })

  it('reports what the user types', async () => {
    const user = userEvent.setup()
    const onChange = vi.fn()
    render(<NumberField label="First number" value="" onChange={onChange} />)

    await user.type(screen.getByLabelText('First number'), '7')

    expect(onChange).toHaveBeenCalledWith('7')
  })

  it('shows an error linked to the field', () => {
    render(
      <NumberField
        label="First number"
        value="abc"
        onChange={vi.fn()}
        error="Enter a valid number"
      />,
    )

    const input = screen.getByLabelText('First number')

    expect(input).toBeInvalid()
    expect(input).toHaveAccessibleDescription('Enter a valid number')
  })

  it('has no error state without an error', () => {
    render(<NumberField label="First number" value="1" onChange={vi.fn()} />)

    expect(screen.getByLabelText('First number')).toBeValid()
    expect(screen.queryByText('Enter a valid number')).not.toBeInTheDocument()
  })
})
