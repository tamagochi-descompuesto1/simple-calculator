import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Window } from './Window'

describe('Window', () => {
  it('shows title as header', () => {
    render(<Window title="calculator.exe">x</Window>)

    const heading = screen.getByRole('heading', { name: 'calculator.exe' })

    expect(heading).toBeInTheDocument()
  })

  it('shows content', () => {
    render(
      <Window title="calculator.exe">
        <p>Body text</p>
      </Window>,
    )

    const component = screen.getByText('Body text')

    expect(component).toBeInTheDocument()
  })

  it('shows actions', () => {
    render(
      <Window title="calculator.exe" actions={<button>Act</button>}>
        x
      </Window>,
    )

    const action = screen.getByRole('button', { name: 'Act' })

    expect(action).toBeInTheDocument()
  })

  it('renders without actions', () => {
    render(
      <Window title="calculator.exe">
        <p>Body text</p>
      </Window>,
    )

    const heading = screen.getByRole('heading', { name: 'calculator.exe' })
    const component = screen.getByText('Body text')

    expect(heading).toBeInTheDocument()
    expect(component).toBeInTheDocument()
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
  })
})
