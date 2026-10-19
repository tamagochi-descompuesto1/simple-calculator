// Package calculator implements the arithmetic operations and their math rules.
// It knows nothing about HTTP.
package calculator

import (
	"errors"
	"math"
)

var (
	// ErrDivisionByZero is returned for a zero divisor.
	ErrDivisionByZero = errors.New("division by zero")
	// ErrNegativeSqrt is returned for the square root of a negative number.
	ErrNegativeSqrt = errors.New("square root of a negative number")
	// ErrOverflow is returned when a result is too large to represent (±Inf).
	ErrOverflow = errors.New("result is too large to represent")
	// ErrUndefined is returned when a result is not a real number (NaN).
	ErrUndefined = errors.New("result is not a real number")
)

// Add returns a + b.
func Add(a, b float64) (float64, error) {
	return checkResult(a + b)
}

// Subtract returns a - b.
func Subtract(a, b float64) (float64, error) {
	return checkResult(a - b)
}

// Multiply returns a * b.
func Multiply(a, b float64) (float64, error) {
	return checkResult(a * b)
}

// Divide returns a / b. It returns [ErrDivisionByZero] if b is 0.
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}

	return checkResult(a / b)
}

// Power returns a raised to the power of b. Zero to a negative power is a
// division by zero, and a negative base with a fractional exponent is undefined.
func Power(a, b float64) (float64, error) {
	if a == 0 && b < 0 {
		return 0, ErrDivisionByZero
	}

	return checkResult(math.Pow(a, b))
}

// Sqrt returns the square root of a. It returns [ErrNegativeSqrt] if a < 0.
func Sqrt(a float64) (float64, error) {
	if a < 0 {
		return 0, ErrNegativeSqrt
	}

	return checkResult(math.Sqrt(a))
}

// Percentage returns b percent of a. It computes a * (b / 100) because
// (a * b) / 100 can overflow when the final result still fits.
func Percentage(a, b float64) (float64, error) {
	return checkResult(a * (b / 100))
}

// checkResult turns NaN and ±Inf into errors. float64 math produces them
// silently, and JSON cannot carry them.
func checkResult(r float64) (float64, error) {
	if math.IsNaN(r) {
		return 0, ErrUndefined
	}
	if math.IsInf(r, 0) {
		return 0, ErrOverflow
	}

	return r, nil
}
