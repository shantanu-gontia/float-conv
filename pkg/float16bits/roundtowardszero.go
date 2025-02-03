package F16

import (
	"math/big"
)

// Utility function that returns the number truncated to a number that can
// be represented as a float16 number.
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the float16 bias applied
// mantissaBits must be passed in their float32 locations.
// NOTE: This doesn't handle the underflow and overflow cases.
func roundTowardsZero(signBit, exponentBits, mantissaBits uint32) (Bits,
	big.Accuracy) {
	return truncate(signBit, exponentBits, mantissaBits)
}

// truncation is the same as rounding towards zero
func truncate(signBit, exponentBits, mantissaBits uint32) (Bits, big.Accuracy) {
	mantissaF16Precision := mantissaBits & f32Float16MantissaMask
	mantissaExtraPrecision := mantissaBits & f32Float16HalfSubnormalMask

	float16Sign := uint16(signBit << 15)
	float16Exponent := uint16(exponentBits << 10)
	// we need to move the mantissa bits to the right, so they align with the
	// mantissa bits in the float32 format.
	float16Mantissa := uint16(mantissaF16Precision >> 13)
	resultVal := Bits(float16Sign | float16Exponent | float16Mantissa)
	resultAcc := big.Exact

	// If there was extra precision, then the number did not fit in the
	// float32 format, so we need to report the status appropriately
	if mantissaExtraPrecision != 0 {
		if signBit == 0 {
			resultAcc = big.Below
		} else {
			resultAcc = big.Above
		}
	}

	return resultVal, resultAcc
}
