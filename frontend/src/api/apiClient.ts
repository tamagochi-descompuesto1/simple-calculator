import { type OperationId } from './operations'

// Backend address. Vite embeds VITE_API_URL at build time (see .env.example).
const BASE_URL: string = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

interface SuccessBody {
  result: number
}

/**
 * An error from the calculator API. The message is safe to show to the user.
 * status is the HTTP status code, or undefined if no response was received.
 */
export class ApiError extends Error {
  readonly status?: number

  constructor(message: string, status?: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

function isSuccessBody(data: unknown): data is SuccessBody {
  return (
    typeof data === 'object' &&
    data !== null &&
    'result' in data &&
    typeof data.result === 'number'
  )
}

// The backend answers failures with {"error": "..."}, but some responses (such
// as the router's 404 and 405) are plain text, so the body may not be JSON.
async function readErrorMessage(response: Response): Promise<string> {
  const data: unknown = await response.json().catch(() => null)

  if (
    typeof data === 'object' &&
    data !== null &&
    'error' in data &&
    typeof data.error === 'string'
  ) {
    return data.error
  }

  return `Request failed (${response.status})`
}

/**
 * Runs an operation on the backend and returns its result. Operations with one
 * operand (such as sqrt) take only a. Rejects with an ApiError on any failure.
 */
export async function calculate(
  operation: OperationId,
  a: number,
  b?: number,
): Promise<number> {
  const body = b === undefined ? { a } : { a, b }

  // fetch only rejects when no response arrives at all (no network, server down).
  let response: Response
  try {
    response = await fetch(`${BASE_URL}/${operation}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
  } catch {
    throw new ApiError('Cannot reach the server')
  }

  // A 4xx or 5xx is still a normal response to fetch, so it must be checked here.
  if (!response.ok) {
    throw new ApiError(await readErrorMessage(response), response.status)
  }

  const data: unknown = await response.json().catch(() => null)
  if (!isSuccessBody(data)) {
    throw new ApiError('Unexpected response from the server', response.status)
  }

  return data.result
}
