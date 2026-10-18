import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { Theme } from '../hooks/useTheme'
import { ThemeToggle } from './ThemeToggle'

describe('ThemeToggle', () => {
  // Each row is [current theme, accessible name of the button]. The button
  // always names the theme it switches to, not the current one.
  it.each<[Theme, string]>([
    ['light', 'Switch to dark theme'],
    ['dark', 'Switch to light theme'],
  ])('with the %s theme, its label is "%s"', (theme, label) => {
    render(<ThemeToggle theme={theme} onToggle={vi.fn()} />)

    expect(screen.getByRole('button', { name: label })).toBeInTheDocument()
  })

  it('calls onToggle once each time it is clicked', async () => {
    const user = userEvent.setup()
    const onToggle = vi.fn()
    render(<ThemeToggle theme="light" onToggle={onToggle} />)
    const button = screen.getByRole('button')

    await user.click(button)

    expect(onToggle).toHaveBeenCalledTimes(1)

    await user.click(button)

    expect(onToggle).toHaveBeenCalledTimes(2)
  })

  // The component only reports the click. Changing the theme is up to its parent.
  it('does not call onToggle until it is clicked', () => {
    const onToggle = vi.fn()

    render(<ThemeToggle theme="light" onToggle={onToggle} />)

    expect(onToggle).not.toHaveBeenCalled()
  })
})
