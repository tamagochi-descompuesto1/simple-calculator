import { useEffect, useState } from 'react'

export type Theme = 'light' | 'dark'

// localStorage key where the user's explicit choice is saved.
const STORAGE_KEY = 'theme'

// Anything can be stored under the key, so a stored value is only trusted
// after checking that it is a valid theme.
function isTheme(value: unknown): value is Theme {
  return value === 'light' || value === 'dark'
}

// The theme the user chose in a previous visit, or null if there is none or
// the storage cannot be read.
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
    // The storage may be blocked (for example in private mode). The choice then
    // lasts only for this visit, which is safe to accept.
  }
}

// The theme the operating system or browser is set to.
function getSystemTheme(): Theme {
  const isDarkMode = window.matchMedia('(prefers-color-scheme: dark)').matches
  return isDarkMode ? 'dark' : 'light'
}

/**
 * Manages the color theme. It starts from the theme the user saved or, if there
 * is none, the system's, and applies it as the data-theme attribute on <html>,
 * which the CSS tokens in index.css react to.
 *
 * Only toggling saves a choice, so users who never choose keep following their
 * system.
 */
export function useTheme() {
  // The function form runs only on the first render, so the storage is not read
  // again on every render.
  const [theme, setTheme] = useState<Theme>(
    () => readStoredTheme() ?? getSystemTheme(),
  )

  useEffect(() => {
    document.documentElement.dataset.theme = theme
  }, [theme])

  // The choice is saved here and not in the effect above: the effect also runs
  // on load, and that would store the system theme as if the user had chosen it.
  function toggleTheme() {
    const next: Theme = theme === 'dark' ? 'light' : 'dark'
    setTheme(next)
    storeTheme(next)
  }

  return { theme, toggleTheme }
}
