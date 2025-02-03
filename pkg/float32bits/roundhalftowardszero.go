package F32

import "math/big"

// Utility function that returns the number rounded to the closest float32
// value. Ties are broken by rounding to the value closest to zero (truncation)
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the float32 bias applied
// mantissaBits must be passed in their float64 locations.
// NOTE: This doesn't handle the underflow and overflow cases.
// The parameter lostPrecision indicates whether the mantissa passed had already
// lost precision during any preprocessing
func roundHalfTowardsZero(signBit, exponentBits,
	mantissaBits uint64, lostPrecision bool) (Bits, big.Accuracy) {

	mantissaF32Precision := mantissaBits & f64float32MantissaMask
	mantissaExtraPrecision := mantissaBits & f64float32HalfSubnormalMask

	float32Sign := uint32(signBit << 31)
	float32Exponent := uint32(exponentBits << 23)
	float32Mantissa := uint32(mantissaF32Precision >> 29)

	exponentMantissaComposite := float32Exponent | float32Mantissa

	// If the extra precision bits exceed 1 0 0 0 0....
	// we need to add 1 to LSB of F32 mantissa, otherwise truncate
	// For all other cases we truncate
	addedOne := false
	if mantissaExtraPrecision > f64float32HalfSubnormalLSB {
		exponentMantissaComposite += 1
		addedOne = true
	}

	if mantissaExtraPrecision == f64float32HalfSubnormalLSB && lostPrecision {
		exponentMantissaComposite += 1
		addedOne = true
	}

	// All we need to do now is attach the sign
	resultVal := Bits(float32Sign | exponentMantissaComposite)
	resultAcc := big.Exact

	// Result is larger if the input was positive and we added 1, or
	// if the input was negative and we truncated.
	if mantissaExtraPrecision != 0 || lostPrecision {
		resultAcc = big.Below
		if (float32Sign == 0) == addedOne {
			resultAcc = big.Above
		}
	}

	return resultVal, resultAcc
}
