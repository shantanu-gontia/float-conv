package floatBit

import "math/big"

// Package roundingModes provides constants (enum) for different rounding modes.
// Rounding modes determine how to round a number when it cannot be represented exactly.

type RoundingMode byte

const (
	// Round towards zero is equivalent to truncation
	// Ties are broken by choosing the number closer to zero
	RTZ              RoundingMode = 0
	RoundTowardsZero RoundingMode = 0

	// Round Nearest Even breaks ties by choosing the number which is even.
	// For rounding to non-integers, this means the least significant
	// digit (in whichever radix) after rounding must be even
	RNE              RoundingMode = 1
	RoundNearestEven RoundingMode = 1
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
