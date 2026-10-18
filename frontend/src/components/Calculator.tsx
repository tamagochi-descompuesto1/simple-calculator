import { useState, type SubmitEvent } from 'react'
import { ApiError, calculate } from '../api/apiClient'
import { OPERATIONS, type OperationId } from '../api/operations'
import { formatResult, parseNumber } from '../lib/numbers'
import { NumberField } from './NumberField'
import { OperationPicker } from './OperationPicker'

// What the result panel shows: nothing yet, a result, or an error message.
type Outcome =
  { kind: 'result'; text: string } | { kind: 'error'; message: string } | null

const INVALID_NUMBER_MESSAGE = 'Enter a valid number'
const UNEXPECTED_ERROR_MESSAGE = 'Something went wrong'

// The numbers an operation is run with. b is missing for operations with a
// single operand.
interface Operands {
  a: number
  b?: number
}

// Writes the calculation the way a person would read it, with the numbers
// formatted like the result ("007" is shown as 7).
function formatCalculation(
  operationId: OperationId,
  symbol: string,
  { a, b }: Operands,
  result: number,
): string {
  const formattedA = formatResult(a)
  const formattedResult = formatResult(result)

  if (b === undefined) {
    return `${symbol}${formattedA} = ${formattedResult}`
  }

  const formattedB = formatResult(b)

  // "15% of 200" reads better than "200 % 15".
  if (operationId === 'percentage') {
    return `${formattedB}% of ${formattedA} = ${formattedResult}`
  }

  return `${formattedA} ${symbol} ${formattedB} = ${formattedResult}`
}

export function Calculator() {
  const [operationId, setOperationId] = useState<OperationId>('add')
  // The numbers stay as the text the user typed; they are parsed on every
  // render instead of being stored twice.
  const [a, setA] = useState('')
  const [b, setB] = useState('')
  const [outcome, setOutcome] = useState<Outcome>(null)
  const [isLoading, setIsLoading] = useState(false)

  // operationId always comes from OPERATIONS, so the operation always exists.
  const operation = OPERATIONS.find((op) => op.id === operationId)!

  const parsedA = parseNumber(a)
  const parsedB = operation.unary ? null : parseNumber(b)

  // The operands to send, or null while the inputs are not valid yet.
  let operands: Operands | null = null
  if (parsedA !== null) {
    if (operation.unary) {
      operands = { a: parsedA }
    } else if (parsedB !== null) {
      operands = { a: parsedA, b: parsedB }
    }
  }

  // An empty field is not an error yet; the button is just disabled.
  const errorA =
    a !== '' && parsedA === null ? INVALID_NUMBER_MESSAGE : undefined
  const errorB =
    !operation.unary && b !== '' && parsedB === null
      ? INVALID_NUMBER_MESSAGE
      : undefined

  // Any change makes the shown result stale, so it is cleared.
  function handleOperationChange(id: OperationId) {
    setOperationId(id)
    setOutcome(null)
  }

  function handleAChange(value: string) {
    setA(value)
    setOutcome(null)
  }

  function handleBChange(value: string) {
    setB(value)
    setOutcome(null)
  }

  async function handleSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    if (operands === null || isLoading) {
      return
    }

    setIsLoading(true)
    try {
      const result = await calculate(operationId, operands.a, operands.b)
      setOutcome({
        kind: 'result',
        text: formatCalculation(
          operationId,
          operation.symbol,
          operands,
          result,
        ),
      })
    } catch (error) {
      const message =
        error instanceof ApiError ? error.message : UNEXPECTED_ERROR_MESSAGE
      setOutcome({ kind: 'error', message })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <OperationPicker value={operationId} onChange={handleOperationChange} />

      <NumberField
        label={operation.unary ? 'Number' : 'First number'}
        value={a}
        onChange={handleAChange}
        error={errorA}
      />
      {!operation.unary && (
        <NumberField
          label="Second number"
          value={b}
          onChange={handleBChange}
          error={errorB}
        />
      )}

      <button
        type="submit"
        disabled={operands === null || isLoading}
        className="w-full rounded-sm border border-line-strong bg-linear-to-b from-brand-light to-brand px-4 py-2 font-semibold text-white shadow-bevel-raised active:shadow-bevel-sunken disabled:cursor-not-allowed disabled:opacity-50 disabled:active:shadow-bevel-raised"
      >
        {isLoading ? 'Calculating…' : 'Calculate'}
      </button>

      {outcome?.kind === 'result' && (
        <p
          role="status"
          className="rounded-sm border border-line bg-cream-dim px-3 py-2 text-right font-mono text-2xl break-all text-ink shadow-bevel-sunken"
        >
          {outcome.text}
        </p>
      )}
      {outcome?.kind === 'error' && (
        <p
          role="alert"
          className="rounded-sm border border-danger bg-cream-dim px-3 py-2 font-semibold text-danger shadow-bevel-sunken"
        >
          {outcome.message}
        </p>
      )}
    </form>
  )
}
