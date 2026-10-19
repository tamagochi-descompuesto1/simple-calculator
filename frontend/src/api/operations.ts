// Every operation the calculator offers. The UI and the API client are built
// from it, so adding one starts here. id is the backend route (POST /<id>),
// unary marks one-operand operations, and fields are the input labels.
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

// The union of all ids, derived so it cannot drift from the list.
export type OperationId = (typeof OPERATIONS)[number]['id']
