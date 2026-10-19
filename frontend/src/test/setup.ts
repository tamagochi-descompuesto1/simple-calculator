import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'

// Testing Library only cleans up after each test by itself when test globals
// are enabled, which this project does not use.
afterEach(() => {
  cleanup()
})

// jsdom has no matchMedia. Default to a light system theme; tests that need
// another replace it with vi.stubGlobal.
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query: string) => ({ matches: false, media: query }),
})
