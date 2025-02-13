package F16

import (
	"math/big"
)

// Utility function that returns the number rounded to the closest float16
// value. Ties are broken by rounding towards the value closer to +Infinity.
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the float16 bias applied
// mantissaBits must be passed in their float32 locations.
// NOTE: This doesn't handle the underflow and overflow cases.
func roundHalfTowardsPositiveInf(signBit, exponentBits,
	mantissaBits uint32, lostPrecision bool) (Bits, big.Accuracy) {

	mantissaF16Precision := mantissaBits & f32Float16MantissaMask
	mantissaExtraPrecision := mantissaBits & f32Float16HalfSubnormalMask

	float16Sign := uint16(signBit << 15)
	float16Exponent := uint16(exponentBits << 10)
	float16Mantissa := uint16(mantissaF16Precision >> 13)

	exponentMantissaComposite := float16Exponent | float16Mantissa

	addedOne := false
	// We definitely add 1, if we're greater than the mid-point
	if mantissaExtraPrecision > f32Float16HalfSubnormalLSB {
		exponentMantissaComposite += 1
		addedOne = true
	}

	// If extra precision was lost before, then we need to add one if we're
	// halfway through in the adjusted mantissa (because this means we're
	// actually greater than the midpoint)
	if mantissaExtraPrecision == f32Float16HalfSubnormalLSB && lostPrecision {
		exponentMantissaComposite += 1
		addedOne = true
	}

	// In the case that we're halfway through,
	// We add 1, only if the sign was positive, otherwise we truncate
	if (mantissaExtraPrecision ==
		f32Float16HalfSubnormalLSB) && (float16Sign == 0) && !lostPrecision {
		exponentMantissaComposite += 1
		addedOne = true
	}

	// For all other case we truncate, so now we can construct the result
	// by attaching the sign
	resultVal := Bits(float16Sign | exponentMantissaComposite)
	resultAcc := big.Exact

	// Result is larger if the input was positive and we added 1, or
	// if the input was negative and we truncated.
	if mantissaExtraPrecision != 0 || lostPrecision {
		resultAcc = big.Below
		if (float16Sign == 0) == addedOne {
			resultAcc = big.Above
		}
	}

	return resultVal, resultAcc
}
