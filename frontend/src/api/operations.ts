// Every operation the calculator offers. The UI builds its selector from this
// list, and the API client uses it for the type of valid operations, so adding an
// operation starts here. Each id must match the backend route (POST /<id>), and
// unary marks operations that take a single operand. fields holds the labels of
// the inputs, named after what each number means when the order is not obvious
// (the base and the exponent, the amount and the percent).
export const OPERATIONS = [
  {
    id: 'add',
    label: 'Add',
    symbol: '+',
    unary: false,
    fields: { a: 'First number', b: 'Second number' },
  },
  {
    id: 'subtract',
    label: 'Subtract',
    symbol: '-',
    unary: false,
    fields: { a: 'First number', b: 'Second number' },
  },
  {
    id: 'multiply',
    label: 'Multiply',
    symbol: '×',
    unary: false,
    fields: { a: 'First number', b: 'Second number' },
  },
  {
    id: 'divide',
    label: 'Divide',
    symbol: '÷',
    unary: false,
    fields: { a: 'First number', b: 'Second number' },
  },
  {
    id: 'power',
    label: 'Power',
    symbol: '^',
    unary: false,
    fields: { a: 'Base', b: 'Exponent' },
  },
  {
    id: 'sqrt',
    label: 'Square Root',
    symbol: '√',
    unary: true,
    fields: { a: 'Number' },
  },
  {
    id: 'percentage',
    label: 'Percentage',
    symbol: '%',
    unary: false,
    fields: { a: 'Amount', b: 'Percent' },
  },
] as const

// The union of all ids ('add' | 'subtract' | ...), derived from the list above so
// the two can never get out of sync.
export type OperationId = (typeof OPERATIONS)[number]['id']
