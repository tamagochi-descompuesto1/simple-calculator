import { useEffect, useState } from 'react'

export type Theme = 'light' | 'dark'

// localStorage key for the user's explicit choice.
const STORAGE_KEY = 'theme'

// A stored value is only trusted after checking it is a valid theme.
function isTheme(value: unknown): value is Theme {
  return value === 'light' || value === 'dark'
}

// The theme saved in a previous visit, or null if there is none or the storage
// is unreadable.
function readStoredTheme(): Theme | null {
  try {
    const theme: string | null = localStorage.getItem(STORAGE_KEY)

    return isTheme(theme) ? theme : null
  } catch {
    return null
  }
}

function storeTheme(theme: Theme) {
  try {
    localStorage.setItem(STORAGE_KEY, theme)
  } catch {
    // Storage may be blocked (private mode); the choice then lasts for this
    // visit only.
  }
}

function getSystemTheme(): Theme {
  const isDarkMode = window.matchMedia('(prefers-color-scheme: dark)').matches
  return isDarkMode ? 'dark' : 'light'
}

/**
 * Manages the color theme: the saved choice, or else the system's, applied as
 * data-theme on <html> (the CSS tokens react to it). Only toggling saves a
 * choice, so users who never choose keep following their system.
 */
export function useTheme() {
  // The function form reads the storage only on the first render.
  const [theme, setTheme] = useState<Theme>(
    () => readStoredTheme() ?? getSystemTheme(),
  )

  useEffect(() => {
    document.documentElement.dataset.theme = theme
  }, [theme])

  // Saved here, not in the effect: the effect also runs on load and would store
  // the system theme as if the user had chosen it.
  function toggleTheme() {
    const next: Theme = theme === 'dark' ? 'light' : 'dark'
    setTheme(next)
    storeTheme(next)
  }

  return { theme, toggleTheme }
}
