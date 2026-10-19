import { act, renderHook } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { useTheme, type Theme } from './useTheme'

// Makes the system report the given theme.
function mockSystemTheme(theme: Theme) {
  vi.stubGlobal(
    'matchMedia',
    vi.fn((query: string) => ({
      matches: theme === 'dark' && query === '(prefers-color-scheme: dark)',
      media: query,
    })),
  )
}

// Each test starts with an empty storage and no theme applied.
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

  // A stored theme must mean "the user chose this"; otherwise the user would
  // stop following their system.
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

    // act() lets React finish updating before the checks.
    act(() => result.current.toggleTheme())

    expect(result.current.theme).toBe('dark')
    expect(document.documentElement.dataset.theme).toBe('dark')
    expect(localStorage.getItem('theme')).toBe('dark')

    act(() => result.current.toggleTheme())

    expect(result.current.theme).toBe('light')
    expect(document.documentElement.dataset.theme).toBe('light')
    expect(localStorage.getItem('theme')).toBe('light')
  })

  // A browser that blocks localStorage: reads and writes throw. The spies stay
  // active for the whole test, so it cannot inspect localStorage itself.
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
