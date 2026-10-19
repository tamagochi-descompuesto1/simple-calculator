import { describe, expect, it } from 'vitest'
import { formatResult, parseNumber } from './numbers'

describe('parseNumber', () => {
  // [typed text, expected number]
  it.each([
    ['5', 5],
    ['-5', -5],
    ['3.14', 3.14],
    ['0', 0],
    ['007', 7],
    ['.5', 0.5],
    // A trailing point is normal while typing.
    ['5.', 5],
    ['  7  ', 7],
  ])('parses %s', (text, expected) => {
    expect(parseNumber(text)).toBe(expected)
  })

  // [description for the test name, text]
  it.each([
    ['an empty string', ''],
    ['only spaces', '  '],
    ['only a sign', '-'],
    ['only a point', '.'],
    ['a sign and a point', '-.'],
    ['letters', 'abc'],
    ['a comma as decimal separator', '1,5'],
    ['two numbers', '5 5'],
    ['a repeated sign', '--5'],
    ['a plus sign', '+5'],
    ['scientific notation', '1e5'],
    ['a hexadecimal number', '0x10'],
    ['two decimal points', '1.2.3'],
    ['Infinity', 'Infinity'],
    ['NaN', 'NaN'],
    // Matches the pattern but overflows to Infinity.
    ['digits beyond the range of a number', '9'.repeat(400)],
  ])('rejects %s', (_description, text) => {
    expect(parseNumber(text)).toBeNull()
  })
})

describe('formatResult', () => {
  // [value, expected text]
  it.each([
    // Float noise is removed.
    [0.1 + 0.2, '0.3'],
    [1.1 * 1.1, '1.21'],
    [1 / 3, '0.333333333333333'],
    // Normal values are unchanged.
    [30, '30'],
    [-5, '-5'],
    [1.3993, '1.3993'],
    // Exact integers up to 15 digits are not cut off.
    [123456789012345, '123456789012345'],
    // JavaScript uses exponent notation for very large and small values.
    [1e21, '1e+21'],
    [1e-7, '1e-7'],
    // Negative zero shows as 0.
    [-0, '0'],
  ])('formats %s as %s', (value, expected) => {
    expect(formatResult(value)).toBe(expected)
  })
})
