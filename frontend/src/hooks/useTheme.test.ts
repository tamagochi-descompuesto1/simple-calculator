import { act, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useTheme, type Theme } from './useTheme'

// Makes the system (the prefers-color-scheme media query) report the given theme.
function mockSystemTheme(theme: Theme) {
  vi.stubGlobal(
    'matchMedia',
    vi.fn((query: string) => ({
      matches: theme === 'dark' && query === '(prefers-color-scheme: dark)',
      media: query,
    })),
  )
}

// Every test starts with an empty storage and no theme applied to the page, so
// the tests do not depend on each other.
beforeEach(() => {
  localStorage.clear()
})

afterEach(() => {
  vi.unstubAllGlobals()
  vi.restoreAllMocks()
  delete document.documentElement.dataset.theme
})

describe('useTheme', () => {
  it.each<Theme>(['light', 'dark'])(
    'starts with the %s system theme',
    (system) => {
      mockSystemTheme(system)

      const { result } = renderHook(() => useTheme())

      expect(result.current.theme).toBe(system)
      expect(document.documentElement.dataset.theme).toBe(system)
    },
  )

  it('prefers the stored theme over the system one', () => {
    mockSystemTheme('light')
    localStorage.setItem('theme', 'dark')

    const { result } = renderHook(() => useTheme())

    expect(result.current.theme).toBe('dark')
    expect(document.documentElement.dataset.theme).toBe('dark')
  })

  it('ignores an invalid stored value', () => {
    mockSystemTheme('dark')
    localStorage.setItem('theme', 'blue')

    const { result } = renderHook(() => useTheme())

    expect(result.current.theme).toBe('dark')
    expect(document.documentElement.dataset.theme).toBe('dark')
  })

  // A theme in storage must mean "the user chose this". If loading the page
  // saved the system theme, the user would stop following their system.
  it('does not store anything until the user chooses', () => {
    mockSystemTheme('dark')

    const { result } = renderHook(() => useTheme())

    expect(result.current.theme).toBe('dark')
    expect(document.documentElement.dataset.theme).toBe('dark')
    expect(localStorage.getItem('theme')).toBeNull()
  })

  it('toggles the theme, applies it and stores the choice', () => {
    mockSystemTheme('light')

    const { result } = renderHook(() => useTheme())

    // act() makes React finish updating before the checks below run.
    act(() => result.current.toggleTheme())

    expect(result.current.theme).toBe('dark')
    expect(document.documentElement.dataset.theme).toBe('dark')
    expect(localStorage.getItem('theme')).toBe('dark')

    act(() => result.current.toggleTheme())

    expect(result.current.theme).toBe('light')
    expect(document.documentElement.dataset.theme).toBe('light')
    expect(localStorage.getItem('theme')).toBe('light')
  })

  // Simulates a browser that blocks localStorage (for example in private mode):
  // reading and writing both throw. The spies stay active for the whole test, so
  // it cannot read localStorage itself to check what was stored. The hook must
  // fall back to the system theme and still be able to toggle.
  it('works when the storage is unavailable', () => {
    mockSystemTheme('light')
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('storage blocked')
    })
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('storage blocked')
    })

    const { result } = renderHook(() => useTheme())

    expect(result.current.theme).toBe('light')
    expect(document.documentElement.dataset.theme).toBe('light')

    act(() => result.current.toggleTheme())

    expect(result.current.theme).toBe('dark')
    expect(document.documentElement.dataset.theme).toBe('dark')
  })
})
