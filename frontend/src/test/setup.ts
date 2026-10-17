import '@testing-library/jest-dom/vitest'

// jsdom does not implement matchMedia. Default to a light system theme; tests
// that need another one replace it with vi.stubGlobal.
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query: string) => ({ matches: false, media: query }),
})
