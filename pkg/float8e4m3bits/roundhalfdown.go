package F8E4M3

import "math/big"


// Utility function that returns the number rounded to the closest float8e4m3
// value. Ties are broken by rounding to the value closest to zero (truncation)
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the float8e4m3 bias and the appropriate
// scale-factor applied
// mantissaBits must be passed in their float32 locations.
// NOTE: This doesn't handle the underflow and overflow cases.
func roundHalfTowardsNegativeInf(signBit, exponentBits,
    mantissaBits uint32, lostPrecision bool) (Bits, big.Accuracy) {
    
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
    // actually greater than the midpoint
    if mantissaExtraPrecision == f32Float8E4M3HalfSubnormalLSB && 
        lostPrecision {
        exponentMantissaComposite += 1
        addedOne = true
    }

    // In the case that we're halfway through
    // we add 1, only if the sign was negative, otherwise we truncate
    if (mantissaExtraPrecision == f32Float8E4M3HalfSubnormalLSB) && 
        (float8E4M3Sign != 0) && !lostPrecision {
        exponentMantissaComposite += 1
        addedOne = true
    }

    // For all other cases, we truncate, so now we can construct the result
    // by attaching the sign
    resultVal := Bits(float8E4M3Sign | exponentMantissaComposite)
    resultAcc := big.Exact

    // Result is larger if the input was positive and we added 1, or
    // if the input was negative and we truncated.
    if mantissaExtraPrecision != 0 || lostPrecision {
        resultAcc = big.Below
        if (float8E4M3Sign == 0) == addedOne {
            resultAcc = big.Above
        }
    }

    return resultVal, resultAcc
}

