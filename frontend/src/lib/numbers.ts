// Optional minus, then digits with an optional decimal part ("5." is allowed
// while typing) or just a decimal part (".5"). No exponents, hex, "+" or words.
const NUMBER_PATTERN = /^-?(\d+\.?\d*|\.\d+)$/

// A double holds about 15 significant digits reliably; float noise such as
// 0.1 + 0.2 appears after that.
const SIGNIFICANT_DIGITS = 15

/** Parses typed text into a number, or null if it is not a valid finite number. */
export function parseNumber(text: string): number | null {
  const trimmed = text.trim()

  if (!NUMBER_PATTERN.test(trimmed)) {
    return null
  }

  // Very long digit strings match the pattern but overflow to Infinity.
  const number = Number(trimmed)
  return Number.isFinite(number) ? number : null
}

/** Formats a result for display, hiding float noise (0.30000000000000004 → 0.3). */
export function formatResult(value: number): string {
  // parseFloat drops the trailing zeros and turns -0 into 0.
  return parseFloat(value.toPrecision(SIGNIFICANT_DIGITS)).toString()
}
