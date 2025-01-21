package BF16

import (
	"math/big"
)

// Utility function that returns the number rounded to the closest float32
// value. Ties are broken by rounding towards the value closer to +Infinity.
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the bfloat16 bias applied
// mantissaBits must be passed in their float32 locations.
// NOTE: This doesn't handle the underflow and overflow cases.
func roundHalfTowardsPositiveInf(signBit, exponentBits,
	mantissaBits uint32) (Bits, big.Accuracy) {

	mantissaBF16Precision := mantissaBits & 0x007f_0000
	mantissaExtraPrecision := mantissaBits & 0x0000_ffff

	bfloat16Sign := uint16(signBit << 15)
	bfloat16Exponent := uint16(exponentBits << 7)
	bfloat16Mantissa := uint16(mantissaBF16Precision >> 16)

	exponentMantissaComposite := bfloat16Exponent | bfloat16Mantissa

	addedOne := false
	// We definitely add 1, if we're greater than the mid-point
	if mantissaExtraPrecision > f32BF16HalfSubnormalLSB {
		exponentMantissaComposite += 1
		addedOne = true
	}

	// In the case that we're halfway through,
	// We add 1, only if the sign was positive, otherwise we truncate
	if (mantissaExtraPrecision ==
		f32BF16HalfSubnormalLSB) && (bfloat16Sign == 0) {
		exponentMantissaComposite += 1
		addedOne = true
	}

	// For all other case we truncate, so now we can construct the result
	// by attaching the sign
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
