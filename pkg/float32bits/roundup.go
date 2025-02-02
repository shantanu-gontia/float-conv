package F32

import (
	"math/big"
)

// Utility function that returns the number rounded to a number that is
// representable in float32. If y is the input number and x < y < x + 1ULP
// where x is a float32 number. Then this rounding mode picks up x + 1ULP
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the float32 bias applied
// mantissaBits must be passed in their float64 locations.
// NOTE: This doesn't handle the underflow and overflow cases.
func roundTowardsPositiveInf(signBit, exponentBits,
	mantissaBits uint64) (Bits, big.Accuracy) {
	return roundUp(signBit, exponentBits, mantissaBits)
}

func roundUp(signBit, exponentBits, mantissaBits uint64) (Bits, big.Accuracy) {

	mantissaF32Precision := mantissaBits & f64float32MantissaMask
	mantissaExtraPrecision := mantissaBits & f64float32HalfSubnormalMask

	float32Sign := uint32(signBit << 31)
	float32Exponent := uint32(exponentBits << 23)
	float32Mantissa := uint32(mantissaF32Precision >> 29)

	// For this rounding mode, we only need to add 1 to the Least-precision
	// mantissa, if the input was positive, to bring it closer to +inf.
	// For negative numbers, this is achieved by simply truncating.

	exponentMantissaComposite := (float32Exponent | float32Mantissa)

	// If positive and there is extra precision, then add 1
	if (float32Sign == 0) && (mantissaExtraPrecision != 0) {
		exponentMantissaComposite += 1
	}
	// Since, we don't handle overflow, all we need to do now is attach the sign
	resultVal := Bits(float32Sign | exponentMantissaComposite)

	resultAcc := big.Exact
	// If there was extra precision bits set, then we need to
	if mantissaExtraPrecision != 0 {
		// We always round to a larger value
		resultAcc = big.Above
	}

	return resultVal, resultAcc
}
