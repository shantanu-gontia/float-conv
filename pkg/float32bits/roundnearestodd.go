package F32

import (
	"math/big"
)

// Utility function that returns the number rounded to the closest float32
// value. Ties are broken by rounding to the odd value (the LSB mantissa
// bit is 1)
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the float32 bias applied
// mantissaBits must be passed in their float64 locations.
// NOTE: This doesn't handle the underflow and overflow cases.
func roundNearestOdd(signBit, exponentBits, mantissaBits uint64) (Bits,
	big.Accuracy) {

	// For rounding to nearest even, we round to the number that is closest and
	// break ties by rounding towards the number that is even (LSB is 0)

	// LSB  |  Extra Precision Bits
	//  m30    m29 m28 m27
	// 1. if m29 m28 m27 ... > 1 0 0 0 ... (more than half) we round up
	// 2. if m29 m28 m27 ... < 1 0 0 0 ... (less than half) we truncate
	// 3. if m29 m28 m27 ... == 1 0 0 0 ... (exactly half), then
	// 	  3.1 m30 == 1, we truncate
	//    3.2 m30 == 0, we round up

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

	mantissaF32LSB := mantissaBits & f64float32SubnormalLSB
	// In the case we're at the mid-point, we only add 1, if the LSB of the
	// float32 retained mantissa is 0
	if (mantissaF32LSB == 0) && (mantissaExtraPrecision ==
		f64float32HalfSubnormalLSB) {
		exponentMantissaComposite += 1
		addedOne = true
	}

	// For all other case we truncate, so now we can construct the result
	// by attaching the sign
	resultVal := Bits(float32Sign | exponentMantissaComposite)
	resultAcc := big.Exact

	// Result is larger if the input was positive and we added 1, or
	// if the input was negative and we truncated.
	if mantissaExtraPrecision != 0 {
		resultAcc = big.Below
		if (float32Sign == 0) == addedOne {
			resultAcc = big.Above
		}
	}

	return resultVal, resultAcc
}
