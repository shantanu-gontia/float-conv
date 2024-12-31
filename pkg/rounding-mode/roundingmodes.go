package roundingModes

// Package roundingModes provides constants (enum) for different rounding modes.
// Rounding modes determine how to round a number when it cannot be represented exactly.

const (
	// Round towards zero is equivalent to truncation
	// Ties are broken by choosing the number closer to zero
	RoundTowardsZero = 0
	// Round Nearest Even breaks ties by choosing the number which is even.
	// For rounding to non-integers, this means the least significant
	// digit (in whichever radix) after rounding must be even
	RoundNearestEven = 1
)
