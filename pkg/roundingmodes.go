package floatBit

import "math/big"

// Package roundingModes provides constants (enum) for different rounding modes.
// Rounding modes determine how to round a number when it cannot be represented exactly.
// See https://en.wikipedia.org/wiki/Rounding for information on rounding.

type RoundingMode byte

const (
	RoundTowardsZero RoundingMode = 0

	RoundTowardsNegativeInf RoundingMode = 1

	RoundTowardsPositiveInf RoundingMode = 2

	RoundHalfTowardsZero RoundingMode = 3

	RoundHalfTowardsNegativeInf RoundingMode = 4

	RoundHalfTowardsPositiveInf RoundingMode = 5

	RoundNearestEven RoundingMode = 6

	RoundNearestOdd RoundingMode = 7
)

// Returns the equivalent rounding mode constant defined in the big stdlib
// package
func (r RoundingMode) ToBigRoundingMode() big.RoundingMode {
	switch r {
	case 0:
		return big.ToZero
	default:
		return big.ToNearestEven
	}
}

// Stringer interface for RoundingMode
func (r RoundingMode) String() string {
	switch r {
	case RoundTowardsZero:
		return "RoundTowardsZero"
	case RoundTowardsPositiveInf:
		return "RoundTowardsPositiveInf"
	case RoundTowardsNegativeInf:
		return "RoundTowardsNegativeInf"
	case RoundHalfTowardsZero:
		return "RoundHalfTowardsZero"
	case RoundHalfTowardsPositiveInf:
		return "RoundHalfTowardsPositiveInf"
	case RoundHalfTowardsNegativeInf:
		return "RoundHalfTowardsNegativeInf"
	case RoundNearestEven:
		return "RoundNearestEven"
	case RoundNearestOdd:
		return "RoundNearestOdd"
	default:
		return ""
	}
}
