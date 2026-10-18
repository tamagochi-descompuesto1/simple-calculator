import { Calculator } from './components/Calculator'
import { ThemeToggle } from './components/ThemeToggle'
import { Window } from './components/Window'
import { useTheme } from './hooks/useTheme'

function App() {
  const { theme, toggleTheme } = useTheme()

  return (
    <main className="flex min-h-screen items-center justify-center p-4">
      <Window
        title="calculator.exe"
        actions={<ThemeToggle theme={theme} onToggle={toggleTheme} />}
      >
        <Calculator />
      </Window>
    </main>
  )
}

export default App
