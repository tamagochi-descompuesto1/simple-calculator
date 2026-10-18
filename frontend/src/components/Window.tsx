import type { ReactNode } from 'react'

interface WindowProps {
  title: string
  actions?: ReactNode
  children: ReactNode
}

export function Window({ title, actions, children }: WindowProps) {
  return (
    <section
      aria-label={title}
      className="w-full max-w-md rounded-lg border-2 border-line-strong bg-paper shadow-window"
    >
      <div className="flex items-center justify-between gap-3 border-b-2 border-line-strong bg-linear-to-b from-brand-light to-brand px-3 py-1.5 font-mono text-sm text-white">
        <h1>{title}</h1>
        <div className="flex items-center gap-2">
          {actions}
          <span aria-hidden="true">_ □ ×</span>
        </div>
      </div>
      <div className="space-y-4 p-6">{children}</div>
    </section>
  )
}
