package F16

import "math/big"

// Utility function that returns the number rounded to a number that is
// representable in float16. If y is the input number and x < y < x + 1ULP
// where x is a float16 number. Then this rounding mode picks up x
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the float16 bias applied
// mantissaBits must be passed in their float32 locations.
// NOTE: This doesn't handle the underflow and overflow cases.
// The parameter lostPrecision indicates whether the mantissa passed had already
// lost precision during any preprocessing
func roundTowardsNegativeInf(signBit, exponentBits,
	mantissaBits uint32, lostPrecision bool) (Bits, big.Accuracy) {
	return roundDown(signBit, exponentBits, mantissaBits, lostPrecision)
}

func roundDown(signBit, exponentBits,
	mantissaBits uint32, lostPrecision bool) (Bits, big.Accuracy) {

	mantissaF16Precision := mantissaBits & f32Float16MantissaMask
	mantissaExtraPrecision := mantissaBits & f32Float16HalfSubnormalMask

	float16Sign := uint16(signBit << 15)
	float16Exponent := uint16(exponentBits << 10)
	float16Mantissa := uint16(mantissaF16Precision >> 13)

	// For this rounding mode, we only need to add 1 to the Least-precision
	// mantissa, if the input was positive, to bring it closer to +inf.
	// For negative numbers, this is achieved by simply truncating.

	exponentMantissaComposite := (float16Exponent | float16Mantissa)

	// If positive and there is extra precision, then add 1
	if (float16Sign != 0) && (mantissaExtraPrecision != 0 || lostPrecision) {
		exponentMantissaComposite += 1
	}
	// Since, we don't handle overflow, all we need to do now is attach the sign
	resultVal := Bits(float16Sign | exponentMantissaComposite)

	resultAcc := big.Exact
	// If there was extra precision bits set, then we need to
	if mantissaExtraPrecision != 0 || lostPrecision {
		// We always round to a smaller value
		resultAcc = big.Below
	}

	return resultVal, resultAcc
}
