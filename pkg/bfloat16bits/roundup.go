package BF16

import (
	"math/big"
)

// Utility function that returns the number rounded to a number that is
// representable in bfloat16. If y is the input number and x < y < x + 1ULP
// where x is a bfloat16 number. Then this rounding mode picks up x + 1ULP
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the bfloat16 bias applied
// mantissaBits must be passed in their float32 locations.
// NOTE: This doesn't handle the underflow and overflow cases.
func roundTowardsPositiveInf(signBit, exponentBits,
	mantissaBits uint32) (Bits, big.Accuracy) {
	return roundUp(signBit, exponentBits, mantissaBits)
}

func roundUp(signBit, exponentBits, mantissaBits uint32) (Bits, big.Accuracy) {

	mantissaBF16Precision := mantissaBits & 0x007f_0000
	mantissaExtraPrecision := mantissaBits & 0x0000_ffff

	bfloat16Sign := uint16(signBit << 15)
	bfloat16Exponent := uint16(exponentBits << 7)
	bfloat16Mantissa := uint16(mantissaBF16Precision >> 16)

	// For this rounding mode, we only need to add 1 to the Least-precision
	// mantissa, if the input was positive, to bring it closer to +inf.
	// For negative numbers, this is achieved by simply truncating.

	exponentMantissaComposite := (bfloat16Exponent | bfloat16Mantissa)

	// If positive and there is extra precision, then add 1
	if (bfloat16Sign == 0) && (mantissaExtraPrecision != 0) {
		exponentMantissaComposite += 1
	}
	// Since, we don't handle overflow, all we need to do now is attach the sign
	resultVal := Bits(bfloat16Sign | exponentMantissaComposite)

	resultAcc := big.Exact
	// If there was extra precision bits set, then we need to
	if mantissaExtraPrecision != 0 {
		// We always round to a larger value
		resultAcc = big.Above
	}

	return resultVal, resultAcc
}
