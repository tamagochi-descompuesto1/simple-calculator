import { useId } from 'react'

interface NumberFieldProps {
  label: string
  // The text as typed: "1." or "-" are normal while typing but not numbers.
  value: string
  onChange: (value: string) => void
  error?: string
}

export function NumberField({
  label,
  value,
  onChange,
  error,
}: NumberFieldProps) {
  // Ties the label and the error to the input.
  const id = useId()
  const errorId = `${id}-error`

  return (
    <div className="space-y-1">
      <label htmlFor={id} className="block text-sm font-semibold text-ink">
        {label}
      </label>
      <input
        id={id}
        type="text"
        inputMode="decimal"
        autoComplete="off"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? errorId : undefined}
        className="w-full rounded-sm border border-line bg-cream-dim px-3 py-2 text-right font-mono text-xl text-ink shadow-bevel-sunken"
      />
      {error && (
        <p id={errorId} className="text-sm text-danger">
          {error}
        </p>
      )}
    </div>
  )
}
