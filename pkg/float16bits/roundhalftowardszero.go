package F16

import "math/big"

// Utility function that returns the number rounded to the closest float16
// value. Ties are broken by rounding to the value closest to zero (truncation)
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the float16 bias applied
// mantissaBits must be passed in their float32 locations.
// NOTE: This doesn't handle the underflow and overflow cases.
// The parameter lostPrecision indicates whether the mantissa passed had already
// lost precision during any preprocessing
func roundHalfTowardsZero(signBit, exponentBits,
	mantissaBits uint32, lostPrecision bool) (Bits, big.Accuracy) {

	mantissaF16Precision := mantissaBits & f32Float16MantissaMask
	mantissaExtraPrecision := mantissaBits & f32Float16HalfSubnormalMask

	float16Sign := uint16(signBit << 15)
	float16Exponent := uint16(exponentBits << 10)
	float16Mantissa := uint16(mantissaF16Precision >> 13)

	exponentMantissaComposite := float16Exponent | float16Mantissa

	// If the extra precision bits exceed 1 0 0 0 0....
	// we need to add 1 to LSB of F32 mantissa, otherwise truncate
	// For all other cases we truncate
	addedOne := false
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

	// All we need to do now is attach the sign
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
