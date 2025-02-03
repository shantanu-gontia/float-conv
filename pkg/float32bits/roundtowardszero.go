package F32

import (
	"math/big"
)

// Utility function that returns the number truncated to a number that can
// be represented as a float32 number.
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the float32 bias applied
// mantissaBits must be passed in their float64 locations.
// NOTE: This doesn't handle the underflow and overflow cases.
func roundTowardsZero(signBit, exponentBits, mantissaBits uint64,
	lostPrecision bool) (Bits, big.Accuracy) {
	return truncate(signBit, exponentBits, mantissaBits, lostPrecision)
}

// truncation is the same as rounding towards zero
func truncate(signBit, exponentBits, mantissaBits uint64,
	lostPrecision bool) (Bits, big.Accuracy) {
	mantissaF32Precision := mantissaBits & f64float32MantissaMask
	mantissaExtraPrecision := mantissaBits & f64float32HalfSubnormalMask

	float32Sign := uint32(signBit << 31)
	float32Exponent := uint32(exponentBits << 23)
	// we need to move the mantissa bits to the right, so they align with the
	// mantissa bits in the float32 format.
	float32Mantissa := uint32(mantissaF32Precision >> 29)
	resultVal := Bits(float32Sign | float32Exponent | float32Mantissa)
	resultAcc := big.Exact

	// If there was extra precision, then the number did not fit in the
	// float32 format, so we need to report the status appropriately
	if mantissaExtraPrecision != 0 || lostPrecision {
		if signBit == 0 {
			resultAcc = big.Below
		} else {
			resultAcc = big.Above
		}
	}

	return resultVal, resultAcc
}
