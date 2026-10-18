import type { Theme } from '../hooks/useTheme'

interface ThemeToggleProps {
  theme: Theme
  onToggle: () => void
}

export function ThemeToggle({ theme, onToggle }: ThemeToggleProps) {
  const next: Theme = theme === 'dark' ? 'light' : 'dark'
  return (
    <button
      type="button"
      aria-label={`Switch to ${next} theme`}
      onClick={onToggle}
      className="flex h-6 min-w-7 items-center justify-center rounded-sm border border-line-strong bg-cream-dim text-xs text-ink shadow-bevel-raised hover:brightness-105 active:shadow-bevel-sunken"
    >
      <span aria-hidden="true">{theme === 'dark' ? '☀️' : '🌙'}</span>
    </button>
  )
}
