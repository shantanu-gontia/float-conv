package BF16

import (
	"math/big"
)

// Utility function that returns the number rounded to the closest bfloat16
// value. Ties are broken by rounding to the odd value (the LSB mantissa
// bit is 1)
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the bfloat16 bias applied
// mantissaBits must be passed in their float32 locations.
// NOTE: This doesn't handle the underflow and overflow cases.
func roundNearestOdd(signBit, exponentBits, mantissaBits uint32) (Bits,
	big.Accuracy) {

	// For rounding to nearest even, we round to the number that is closest and
	// break ties by rounding towards the number that is even (LSB is 0)

	// LSB  |  Extra Precision Bits
	//  m16    m15 m14 m13
	// 1. if m15 m14 m13 ... > 1 0 0 0 ... (more than half) we round up
	// 2. if m15 m14 m13 ... < 1 0 0 0 ... (less than half) we truncate
	// 3. if m15 m14 m13 ... == 1 0 0 0 ... (exactly half), then
	// 	  3.1 m16 == 1, we truncate
	//    3.2 m16 == 0, we round up

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

	mantissaF32LSB := mantissaBits & 0x0001_0000
	// In the case we're at the mid-point, we only add 1, if the LSB of the
	// float32 retained mantissa is 0
	if (mantissaF32LSB == 0) && (mantissaExtraPrecision ==
		f32BF16HalfSubnormalLSB) {
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
