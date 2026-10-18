import { OPERATIONS, type OperationId } from '../api/operations'

interface OperationPickerProps {
  value: OperationId
  onChange: (operation: OperationId) => void
}

// A keypad of operations. Each key is a radio button, so the keyboard (arrow
// keys) and screen readers work without extra code; the radio itself is hidden
// and its label is styled as a key that looks pressed when selected.
export function OperationPicker({ value, onChange }: OperationPickerProps) {
  return (
    <fieldset>
      <legend className="sr-only">Operation</legend>
      <div className="grid grid-cols-4 gap-2 md:grid-cols-7">
        {OPERATIONS.map((operation) => (
          <label key={operation.id}>
            <input
              type="radio"
              name="operation"
              className="peer sr-only"
              checked={value === operation.id}
              onChange={() => onChange(operation.id)}
            />
            <span className="flex cursor-pointer flex-col items-center rounded-sm border border-line-strong bg-cream-dim px-1 py-2 text-ink shadow-bevel-raised peer-checked:bg-brand-light peer-checked:text-white peer-checked:shadow-bevel-sunken peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-brand">
              {/* The symbol is decorative: the name below is what is announced. */}
              <span aria-hidden="true" className="font-mono text-xl">
                {operation.symbol}
              </span>
              <span className="text-xs">{operation.label}</span>
            </span>
          </label>
        ))}
      </div>
    </fieldset>
  )
}
