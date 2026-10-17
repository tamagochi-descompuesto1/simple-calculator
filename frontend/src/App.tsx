// Temporary theme preview, replaced by the real calculator layout later.
function App() {
  return (
    <main className="flex min-h-screen items-center justify-center p-4 font-head">
      <div className="w-full max-w-md rounded-lg border-2 border-line-strong bg-paper shadow-window">
        <div className="flex items-center justify-between border-b-2 border-line-strong bg-linear-to-b from-brand-light to-brand px-3 py-1.5 font-mono text-sm text-white">
          <span>calculator.exe</span>
          <span aria-hidden="true">_ □ ×</span>
        </div>

        <div className="space-y-4 p-6">
          <h1 className="text-3xl font-bold text-ink">Hello World!</h1>

          <div className="space-y-1">
            <p className="text-ink">Ink (main text)</p>
            <p className="text-ink-soft">Ink soft (secondary text)</p>
            <p className="text-ink-faint">Ink faint (hints)</p>
            <p className="text-danger">Danger (errors)</p>
          </div>

          <div className="rounded-sm border border-line bg-cream-dim px-3 py-2 text-right font-mono text-2xl text-ink">
            6 + 3 = 9
          </div>

          <div className="flex gap-2">
            <button className="rounded-sm border border-line-strong bg-cream-dim px-4 py-2 font-semibold text-ink shadow-bevel-raised active:shadow-bevel-sunken">
              Raised button
            </button>
            <button className="rounded-sm border border-line-strong bg-linear-to-b from-brand-light to-brand px-4 py-2 font-semibold text-white shadow-bevel-raised active:shadow-bevel-sunken">
              Primary
            </button>
          </div>

          <div className="flex gap-2" aria-hidden="true">
            <span className="h-6 w-6 rounded-sm bg-brand-light" />
            <span className="h-6 w-6 rounded-sm bg-brand" />
            <span className="h-6 w-6 rounded-sm bg-brand-dark" />
            <span className="h-6 w-6 rounded-sm bg-danger" />
            <span className="h-6 w-6 rounded-sm bg-success" />
          </div>
        </div>
      </div>
    </main>
  )
}

export default App
