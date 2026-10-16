package calculator

import (
	"errors"
	"fmt"
	"math"
	"testing"
)

// assertResult fails the test if err does not match wantErr, or if got is not
// within a small relative tolerance of want. Exact float equality is too strict
// for results such as 54.2 + 31.6, which is 85.80000000000001 in float64.
func assertResult(t *testing.T, call string, got float64, err error, want float64, wantErr error) {
	t.Helper()

	if !errors.Is(err, wantErr) {
		t.Fatalf("%s error = %v, want %v", call, err, wantErr)
	}

	tolerance := 1e-9 * math.Max(1, math.Abs(want))
	if math.Abs(got-want) > tolerance {
		t.Errorf("%s = %v, want %v", call, got, want)
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name    string
		a       float64
		b       float64
		want    float64
		wantErr error
	}{
		{name: "positive whole numbers", a: 50, b: 34, want: 85},
		{name: "negative whole numbers", a: -68, b: -32, want: -100},
		{name: "mixed signs", a: 50, b: -80, want: -30},
		{name: "positive decimals", a: 54.2, b: 31.6, want: 85.8},
		{name: "negative decimals", a: -63.3, b: -94.8, want: -158.1},
		{name: "positive overflow", a: 1e308, b: 1e308, wantErr: ErrOverflow},
		{name: "negative overflow", a: -1e308, b: -1e308, wantErr: ErrOverflow},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Add(tc.a, tc.b)
			assertResult(t, fmt.Sprintf("Add(%v, %v)", tc.a, tc.b), got, err, tc.want, tc.wantErr)
		})
	}
}

func TestSubtract(t *testing.T) {
	tests := []struct {
		name    string
		a       float64
		b       float64
		want    float64
		wantErr error
	}{
		{name: "positive whole numbers", a: 50, b: 34, want: 16},
		{name: "negative whole numbers", a: -68, b: -32, want: -36},
		{name: "negative result", a: 34, b: 50, want: -16},
		{name: "positive decimals", a: 54.2, b: 31.6, want: 22.6},
		{name: "negative decimals", a: -63.3, b: -94.8, want: 31.5},
		{name: "positive overflow", a: 1e308, b: -1e308, wantErr: ErrOverflow},
		{name: "negative overflow", a: -1e308, b: 1e308, wantErr: ErrOverflow},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Subtract(tc.a, tc.b)
			assertResult(t, fmt.Sprintf("Subtract(%v, %v)", tc.a, tc.b), got, err, tc.want, tc.wantErr)
		})
	}
}

func TestMultiply(t *testing.T) {
	tests := []struct {
		name    string
		a       float64
		b       float64
		want    float64
		wantErr error
	}{
		{name: "positive whole numbers", a: 5, b: 3, want: 15},
		{name: "negative whole numbers", a: -5, b: -3, want: 15},
		{name: "positive by negative", a: 5, b: -3, want: -15},
		{name: "negative by positive", a: -5, b: 3, want: -15},
		{name: "positive decimals", a: 5.4, b: 3.3, want: 17.82},
		{name: "negative decimals", a: -5.4, b: -3.3, want: 17.82},
		{name: "mixed sign decimals", a: 5.4, b: -3.3, want: -17.82},
		{name: "zero times number", a: 0, b: 6, want: 0},
		{name: "number times zero", a: 5.3, b: 0, want: 0},
		{name: "positive overflow", a: 1e308, b: 10, wantErr: ErrOverflow},
		{name: "negative overflow", a: 1e308, b: -10, wantErr: ErrOverflow},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Multiply(tc.a, tc.b)
			assertResult(t, fmt.Sprintf("Multiply(%v, %v)", tc.a, tc.b), got, err, tc.want, tc.wantErr)
		})
	}
}

func TestDivide(t *testing.T) {
	tests := []struct {
		name    string
		a       float64
		b       float64
		want    float64
		wantErr error
	}{
		{name: "positive whole numbers", a: 15, b: 3, want: 5},
		{name: "negative whole numbers", a: -15, b: -3, want: 5},
		{name: "positive by negative", a: 15, b: -3, want: -5},
		{name: "negative by positive", a: -15, b: 3, want: -5},
		{name: "positive decimals", a: 5.4, b: 5, want: 1.08},
		{name: "mixed sign decimals", a: -5.4, b: 5, want: -1.08},
		{name: "non-terminating result", a: 1, b: 3, want: 0.3333333333333333},
		{name: "zero numerator", a: 0, b: 5, want: 0},
		{name: "division by zero", a: 5, b: 0, wantErr: ErrDivisionByZero},
		{name: "division by negative zero", a: 5, b: math.Copysign(0, -1), wantErr: ErrDivisionByZero},
		{name: "zero by zero", a: 0, b: 0, wantErr: ErrDivisionByZero},
		{name: "positive overflow", a: 1e308, b: 1e-10, wantErr: ErrOverflow},
		{name: "negative overflow", a: -1e308, b: 1e-10, wantErr: ErrOverflow},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Divide(tc.a, tc.b)
			assertResult(t, fmt.Sprintf("Divide(%v, %v)", tc.a, tc.b), got, err, tc.want, tc.wantErr)
		})
	}
}

func TestPower(t *testing.T) {
	tests := []struct {
		name    string
		a       float64
		b       float64
		want    float64
		wantErr error
	}{
		{name: "positive whole numbers", a: 2, b: 10, want: 1024},
		{name: "negative exponent", a: 2, b: -1, want: 0.5},
		{name: "negative base and exponent", a: -2, b: -1, want: -0.5},
		{name: "decimal base", a: 2.5, b: 3, want: 15.625},
		{name: "negative decimal base and exponent", a: -2.5, b: -3, want: -0.064},
		{name: "fractional exponent", a: 9, b: 0.5, want: 3},
		{name: "zero exponent", a: 10, b: 0, want: 1},
		{name: "zero to the zero", a: 0, b: 0, want: 1},
		{name: "zero base positive exponent", a: 0, b: 3, want: 0},
		{name: "zero base negative exponent", a: 0, b: -1, wantErr: ErrDivisionByZero},
		{name: "negative base fractional exponent", a: -8, b: 0.5, wantErr: ErrUndefined},
		{name: "positive overflow", a: 10, b: 400, wantErr: ErrOverflow},
		{name: "negative overflow", a: -10, b: 401, wantErr: ErrOverflow},
		// Results too small to represent round to 0 (underflow); that is not an error.
		{name: "underflow rounds to zero", a: 10, b: -400, want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Power(tc.a, tc.b)
			assertResult(t, fmt.Sprintf("Power(%v, %v)", tc.a, tc.b), got, err, tc.want, tc.wantErr)
		})
	}
}

func TestSqrt(t *testing.T) {
	tests := []struct {
		name    string
		a       float64
		want    float64
		wantErr error
	}{
		{name: "perfect square", a: 9, want: 3},
		{name: "decimal perfect square", a: 1.44, want: 1.2},
		{name: "irrational result", a: 2, want: math.Sqrt2},
		{name: "zero", a: 0, want: 0},
		{name: "largest magnitude", a: 1e308, want: 1e154},
		{name: "negative whole number", a: -1, wantErr: ErrNegativeSqrt},
		{name: "negative decimal", a: -9.5, wantErr: ErrNegativeSqrt},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Sqrt(tc.a)
			assertResult(t, fmt.Sprintf("Sqrt(%v)", tc.a), got, err, tc.want, tc.wantErr)
		})
	}
}

func TestPercentage(t *testing.T) {
	tests := []struct {
		name    string
		a       float64
		b       float64
		want    float64
		wantErr error
	}{
		{name: "whole numbers", a: 200, b: 15, want: 30},
		{name: "decimal amount", a: 19.99, b: 7, want: 1.3993},
		{name: "decimal percent", a: 80, b: 12.5, want: 10},
		{name: "hundred percent", a: 80, b: 100, want: 80},
		{name: "over hundred percent", a: 80, b: 250, want: 200},
		{name: "zero percent", a: 80, b: 0, want: 0},
		{name: "percent of zero", a: 0, b: 50, want: 0},
		{name: "negative percent", a: 50, b: -10, want: -5},
		{name: "negative amount", a: -200, b: 15, want: -30},
		// Regression: computing (a*b)/100 overflowed here even though the result fits.
		{name: "large amount without intermediate overflow", a: 1e308, b: 50, want: 5e307},
		{name: "overflow", a: 1e308, b: 1000, wantErr: ErrOverflow},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Percentage(tc.a, tc.b)
			assertResult(t, fmt.Sprintf("Percentage(%v, %v)", tc.a, tc.b), got, err, tc.want, tc.wantErr)
		})
	}
}
