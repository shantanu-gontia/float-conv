package floatBit

import "math/big"

// Package roundingModes provides constants (enum) for different rounding modes.
// Rounding modes determine how to round a number when it cannot be represented exactly.
// See https://en.wikipedia.org/wiki/Rounding for information on rounding.

type RoundingMode byte

const (
	// Round towards zero is equivalent to truncation
	RTZ              RoundingMode = 0
	RoundTowardsZero RoundingMode = 0

	// Rounding down means choosing the number closer to negative infinity
	RoundDown               RoundingMode = 1
	RoundTowardsNegativeInf RoundingMode = 1

	// Rounding up means choosing the number closer to positive infinity
	RoundUp                 RoundingMode = 2
	RoundTowardsPositiveInf RoundingMode = 2

	// Round half towards zero means rounding to the nearest representable, and
	// breaking ties, by rounding towards zero
	RoundHalfTowardsZero RoundingMode = 3

	// Rounding half down means rounding to the nearest, and breaking ties
	// by rounding towards negative infinity
	RoundHalfDown               RoundingMode = 4
	RoundHalfTowardsNegativeInf RoundingMode = 4

	// Rounding half up means rounding to the nearest, and breaking ties
	// by rounding towards positive infinity
	RoundHalfUp                 RoundingMode = 5
	RoundHalfTowardsPositiveInf RoundingMode = 5

	// Round Nearest Even breaks ties by choosing the number which is even.
	// For rounding to non-integers, this means the least significant
	// digit (in whichever radix) after rounding must be even
	// This is the "round half to even" in https://en.wikipedia.org/wiki/Rounding
	RNE              RoundingMode = 6
	RoundNearestEven RoundingMode = 6

	// Round Nearest Even breaks ties by choosing the number which is even.
	// For rounding to non-integers, this means the least significant
	// digit (in whichever radix) after rounding must be odd
	// This is the "round half to odd" in https://en.wikipedia.org/wiki/Rounding
	RNO             RoundingMode = 7
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
