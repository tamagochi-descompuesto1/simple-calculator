// Package calculator implements basic arithmetic operations and enforces
// their math-domain rules. It has no knowledge of HTTP.
package calculator

import (
	"errors"
	"math"
)

var (
	// ErrDivisionByZero is returned when the denominator of a division is 0.
	ErrDivisionByZero = errors.New("division by zero")
	// ErrNegativeSqrt is returned when the operand of a square root is negative.
	ErrNegativeSqrt = errors.New("square root of a negative number")
	// ErrOverflow is returned when the result's magnitude is too large to represent.
	ErrOverflow = errors.New("result is too large to represent")
	// ErrUndefined is returned when the result of an operation is not a real number.
	ErrUndefined = errors.New("result is not a real number")
)

// Add returns the sum of a and b.
// It returns [ErrOverflow] if the result's magnitude is too large to represent.
func Add(a, b float64) (float64, error) {
	return checkResult(a + b)
}

// Subtract returns the difference of a and b.
// It returns [ErrOverflow] if the result's magnitude is too large to represent.
func Subtract(a, b float64) (float64, error) {
	return checkResult(a - b)
}

// Multiply returns the product of a and b.
// It returns [ErrOverflow] if the result's magnitude is too large to represent.
func Multiply(a, b float64) (float64, error) {
	return checkResult(a * b)
}

// Divide returns the quotient of a divided by b.
// It returns [ErrDivisionByZero] if b == 0.
// It returns [ErrOverflow] if the result's magnitude is too large to represent.
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}

	return checkResult(a / b)
}

// Power returns a raised to the power of b.
// It returns [ErrDivisionByZero] if a == 0 and b < 0.
// It returns [ErrOverflow] if the result's magnitude is too large to represent.
// It returns [ErrUndefined] if the result is not a real number.
func Power(a, b float64) (float64, error) {
	if a == 0 && b < 0 {
		// 0^(-n) == 1/0^n, i.e. a division by zero.
		return 0, ErrDivisionByZero
	}

	return checkResult(math.Pow(a, b))
}

// Sqrt returns the square root of a.
// It returns [ErrNegativeSqrt] if a < 0.
func Sqrt(a float64) (float64, error) {
	if a < 0 {
		return 0, ErrNegativeSqrt
	}

	return checkResult(math.Sqrt(a))
}

// Percentage returns b percent of a, i.e. a * b / 100.
// It returns [ErrOverflow] if the result's magnitude is too large to represent.
func Percentage(a, b float64) (float64, error) {
	return checkResult(a * (b / 100))
}

// checkResult guards every operation's output, because float64 arithmetic
// never fails: it silently produces NaN or ±Inf, which cannot be encoded as JSON.
// It returns [ErrUndefined] if r is NaN and [ErrOverflow] if r is ±Inf;
// otherwise it returns r unchanged.
func checkResult(r float64) (float64, error) {
	if math.IsNaN(r) {
		return 0, ErrUndefined
	}
	if math.IsInf(r, 0) {
		return 0, ErrOverflow
	}

	return r, nil
}
