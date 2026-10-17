// Every operation the calculator offers. The UI builds its selector from this
// list, and the API client uses it for the type of valid operations, so adding an
// operation starts here. Each id must match the backend route (POST /<id>), and
// unary marks operations that take a single operand.
export const OPERATIONS = [
    { id: 'add', label: 'Add', symbol: '+', unary: false },
    { id: 'subtract', label: 'Subtract', symbol: '-', unary: false },
    { id: 'multiply', label: 'Multiply', symbol: '×', unary: false },
    { id: 'divide', label: 'Divide', symbol: '÷', unary: false },
    { id: 'power', label: 'Power', symbol: '^', unary: false },
    { id: 'sqrt', label: 'Square Root', symbol: '√', unary: true },
    { id: 'percentage', label: 'Percentage', symbol: '%', unary: false }
] as const

// The union of all ids ('add' | 'subtract' | ...), derived from the list above so
// the two can never get out of sync.
export type OperationId =(typeof OPERATIONS)[number]['id']