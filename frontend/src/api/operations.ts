export const OPERATIONS = [
    { id: 'add', label: 'Add', symbol: '+', unary: false },
    { id: 'subtract', label: 'Subtract', symbol: '-', unary: false },
    { id: 'multiply', label: 'Multiply', symbol: '×', unary: false },
    { id: 'divide', label: 'Divide', symbol: '÷', unary: false },
    { id: 'power', label: 'Power', symbol: '^', unary: false },
    { id: 'sqrt', label: 'Square Root', symbol: '√', unary: true },
    { id: 'percentage', label: 'Percentage', symbol: '%', unary: false }
] as const

export type OperationId = (typeof OPERATIONS)[number]['id']