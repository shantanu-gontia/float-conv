package BF16

import "math/big"

// Utility function that returns the number truncated to a number that can
// be represented as a bfloat16 number.
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the bfloat16 bias applied
// mantissaBits must be passed in their float32 locations.
// NOTE: This doesn't handle the underflow and overflow cases.
func roundTowardsZero(signBit, exponentBits, mantissaBits uint32) (Bits,
	big.Accuracy) {
	return truncate(signBit, exponentBits, mantissaBits)
}

// truncation is the same as rounding towards zero
func truncate(signBit, exponentBits, mantissaBits uint32) (Bits, big.Accuracy) {
	mantissaBF16Precision := mantissaBits & 0x007f_0000
	mantissaExtraPrecision := mantissaBits & 0x0000_ffff

	bfloat16Sign := uint16(signBit << 15)
	bfloat16Exponent := uint16(exponentBits << 7)
	// we need to move mantissa bits to the right, so they align with the
	// mantissa bits in the bfloat16 format
	bfloat16Mantissa := uint16(mantissaBF16Precision >> 16)

	resultVal := Bits(bfloat16Sign | bfloat16Exponent | bfloat16Mantissa)
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
