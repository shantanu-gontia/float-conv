package F8E4M3

import "math/big"

// Utility function that returns the number rounded to the closest float8e4m3
// value. Ties are broken by rounding to the odd value (the LSB mantissa
// bit is 1)
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the float8e4m3 bias and appropriate
// scale-factor applied
// mantissaBits must be passed in their float32 locations.
// NOTE: This doesn't handle the underflow and overflow cases.
// The parameter lostPrecision indicates whether the mantissa passed had already
// lost precision during any preprocessing
func roundNearestOdd(signBit, exponentBits, mantissaBits uint32,
	lostPrecision bool) (Bits,
	big.Accuracy) {

	// For rounding to nearest even, we round to the number that is closest and
	// break ties by rounding towards the number that is even (LSB is 0)

	// LSB  |  Extra Precision Bits
	// m20    m19 m18 m17 ... m0
	// 1. if m19 m18 m17 ... > 1 0 0 0 ... (more than half) we round up
	// 2. if m19 m18 m17 ... < 1 0 0 0 ... (less than half) we truncate
	// 3. if m19 m18 m17 ... == 1 0 0 0 ... (exactly half), then
	// 	  3.1 m20 == 1, we truncate
	//    3.2 m20 == 0, we round up

    mantissaF8E4M3Precision := mantissaBits & f32Float8E4M3MantissaMask
    mantissaExtraPrecision := mantissaBits & f32Float8E4M3HalfSubnormalMask

    float8E4M3Sign := uint8(signBit << 7)
    float8E4M3Exponent := uint8(exponentBits << 3)
    float8E4M3Mantissa := uint8(mantissaF8E4M3Precision >> 20)

    exponentMantissaComposite := float8E4M3Exponent | float8E4M3Mantissa

    addedOne := false
    // We definitely add 1, if we're greater than the mid-point
    if mantissaExtraPrecision > f32Float8E4M3HalfSubnormalLSB {
        exponentMantissaComposite += 1
        addedOne = true
    }

    // If extra precision was lost before, then we need to add one if we're
    // halfway through in the adjusted mantissa (because this means we're
    // actually greater than the midpoint)
    if mantissaExtraPrecision == f32Float8E4M3HalfSubnormalLSB &&
        lostPrecision {
        exponentMantissaComposite += 1
        addedOne = true
    }

    mantissaF32LSB := mantissaBits & f32Float8E4M3SubnormalLSB
    // In the case we're at the mid-point, we only add 1, if the LSB of the
    // float8e4m3 retained mantissa is 0
    if (mantissaF32LSB == 0) && (mantissaExtraPrecision ==
        f32Float8E4M3HalfSubnormalLSB) && !lostPrecision {
        exponentMantissaComposite += 1
        addedOne = true
    }

    // For all other cases we truncate, so, now we can construct the result
    // by attaching the sign
    resultVal := Bits(float8E4M3Sign | exponentMantissaComposite)
    resultAcc := big.Exact

    // Result is larger if the input was positive and we added 1, or,
    // if the input was negative and we truncated
    if mantissaExtraPrecision != 0 || lostPrecision {
        resultAcc = big.Below
        if (float8E4M3Sign == 0) == addedOne {
            resultAcc = big.Above
        }
    }
    
    return resultVal, resultAcc
}

