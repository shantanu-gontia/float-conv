package BF16

import "math/big"

// Utility function that returns the number rounded to the closest float32
// value. Ties are broken by rounding to the value closest to zero (truncation)
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the float32 bias applied
// mantissaBits must be passed in their float64 locations.
// NOTE: This doesn't handle the underflow and overflow cases.
func roundHalfTowardsZero(signBit, exponentBits,
	mantissaBits uint32) (Bits, big.Accuracy) {

	mantissaBF16Precision := mantissaBits & 0x007f_0000
	mantissaExtraPrecision := mantissaBits & 0x0000_ffff

	bfloat16Sign := uint16(signBit << 15)
	bfloat16Exponent := uint16(exponentBits << 7)
	bfloat16Mantissa := uint16(mantissaBF16Precision >> 16)

	exponentMantissaComposite := bfloat16Exponent | bfloat16Mantissa

	// If the extra precision bits exceed 1 0 0 0 0....
	// we need to add 1 to LSB of F32 mantissa, otherwise truncate
	// For all other cases we truncate
	addedOne := false
	if mantissaExtraPrecision > f32BF16HalfSubnormalLSB {
		exponentMantissaComposite += 1
		addedOne = true
	}

	// All we need to do now is attach the sign
	resultVal := Bits(bfloat16Sign | exponentMantissaComposite)
	resultAcc := big.Exact

	// Result is larger if the input was positive and we added 1, or
	// if the input was negative and we truncated.
	if mantissaExtraPrecision != 0 {
		resultAcc = big.Below
		if (bfloat16Sign == 0) == addedOne {
			resultAcc = big.Above
		}
	}

	return resultVal, resultAcc
}
