package F32

import (
	"math/big"
)

// Utility function that returns the number rounded to the closest float32
// value. Ties are broken by rounding towards the value closer to +Infinity.
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the float32 bias applied
// mantissaBits must be passed in their float64 locations.
// NOTE: This doesn't handle the underflow and overflow cases.
// The parameter lostPrecision indicates whether the mantissa passed had already
// lost precision during any preprocessing
func roundHalfTowardsPositiveInf(signBit, exponentBits,
	mantissaBits uint64, lostPrecision bool) (Bits, big.Accuracy) {

	mantissaF32Precision := mantissaBits & f64float32MantissaMask
	mantissaExtraPrecision := mantissaBits & f64float32HalfSubnormalMask

	float32Sign := uint32(signBit << 31)
	float32Exponent := uint32(exponentBits << 23)
	float32Mantissa := uint32(mantissaF32Precision >> 29)

	exponentMantissaComposite := float32Exponent | float32Mantissa

	addedOne := false
	// We definitely add 1, if we're greater than the mid-point
	if mantissaExtraPrecision > f64float32HalfSubnormalLSB {
		exponentMantissaComposite += 1
		addedOne = true
	}

	// If extra precision was lost before, then we need to add one if we're
	// halfway through in the adjusted mantissa (because this means we're
	// actually greater than the midpoint)
	if mantissaExtraPrecision == f64float32HalfSubnormalLSB && lostPrecision {
		exponentMantissaComposite += 1
		addedOne = true
	}

	// In the case that we're halfway through,
	// We add 1, only if the sign was positive, otherwise we truncate
	if (mantissaExtraPrecision ==
		f64float32HalfSubnormalLSB) && (float32Sign == 0) && !lostPrecision {
		exponentMantissaComposite += 1
		addedOne = true
	}

	// For all other case we truncate, so now we can construct the result
	// by attaching the sign
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
