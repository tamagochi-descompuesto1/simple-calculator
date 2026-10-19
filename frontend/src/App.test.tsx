import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import App from './App'

// The theme lives in localStorage and on <html>, so each test starts clean.
// The setup's default matchMedia reports a light system theme.
beforeEach(() => {
  localStorage.clear()
})

afterEach(() => {
  delete document.documentElement.dataset.theme
})

describe('App', () => {
  it('shows the window with its title', () => {
    render(<App />)

    expect(
      screen.getByRole('heading', { name: 'calculator.exe' }),
    ).toBeInTheDocument()
  })

  it('shows the calculator inside the window', () => {
    render(<App />)

    expect(screen.getByRole('group', { name: 'Operation' })).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: 'Calculate' }),
    ).toBeInTheDocument()
  })

  it('starts with the system theme', () => {
    render(<App />)

    expect(document.documentElement.dataset.theme).toBe('light')
    expect(
      screen.getByRole('button', { name: 'Switch to dark theme' }),
    ).toBeInTheDocument()
  })

  it('switches the theme when the toggle is clicked', async () => {
    const user = userEvent.setup()
    render(<App />)

    await user.click(
      screen.getByRole('button', { name: 'Switch to dark theme' }),
    )

    expect(document.documentElement.dataset.theme).toBe('dark')
    expect(
      screen.getByRole('button', { name: 'Switch to light theme' }),
    ).toBeInTheDocument()
  })

  // Rendering again stands in for a reload.
  it('remembers the chosen theme after a reload', async () => {
    const user = userEvent.setup()
    const { unmount } = render(<App />)
    await user.click(
      screen.getByRole('button', { name: 'Switch to dark theme' }),
    )
    unmount()

    render(<App />)

    expect(document.documentElement.dataset.theme).toBe('dark')
    expect(
      screen.getByRole('button', { name: 'Switch to light theme' }),
    ).toBeInTheDocument()
  })
})
