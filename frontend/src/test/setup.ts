import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'

// Testing Library only removes rendered components after each test by itself
// when the test globals are enabled, which this project does not use.
afterEach(() => {
  cleanup()
})

// jsdom does not implement matchMedia. Default to a light system theme; tests
// that need another one replace it with vi.stubGlobal.
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query: string) => ({ matches: false, media: query }),
})
