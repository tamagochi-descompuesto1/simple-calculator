// An optional minus sign followed by digits with an optional decimal part
// ("5", "-5", "3.14", "5."), or just a decimal part (".5"). "5." is accepted
// because it is a normal state while the user is still typing. Exponents,
// hex, "+", thousands separators and words like "Infinity" are rejected.
const NUMBER_PATTERN = /^-?(\d+\.?\d*|\.\d+)$/

// A double holds about 15 significant digits reliably; the noise in results
// such as 0.1 + 0.2 appears after that. Rounding to 15 hides the noise without
// cutting off exact results.
const SIGNIFICANT_DIGITS = 15

/**
 * Converts what the user typed into a number, ignoring surrounding spaces.
 * Returns null if the text is not a valid finite number.
 */
export function parseNumber(text: string): number | null {
  const trimmed = text.trim()

  if (!NUMBER_PATTERN.test(trimmed)) {
    return null
  }

  // A very long string of digits matches the pattern but overflows to
  // Infinity, which the backend would reject.
  const number = Number(trimmed)
  return Number.isFinite(number) ? number : null
}

/**
 * Formats a result for display, removing floating-point noise
 * (0.30000000000000004 is shown as 0.3).
 */
export function formatResult(value: number): string {
  // parseFloat drops the trailing zeros that toPrecision adds, and turns -0 into 0.
  return parseFloat(value.toPrecision(SIGNIFICANT_DIGITS)).toString()
}
