package F8E4M3

import (
    "math/big"
)

// Utility function that returns the number rounded to an number that is
// representable in float8e4m3. If y is the input number and x < y < x + 1ULP
// where x is a float8e4m3 number. Then this rounding mode picks up x + 1ULP
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the appropriate scaleFactor and float8e4m3
// bias already applied
// mantissaBits must be passed in their float32 locations
// NOTE: This doesn't handle the underflow and overflow cases
func roundTowardsPositiveInf(signBit, exponentBits, mantissaBits uint32,
    lostPrecision bool) (Bits, big.Accuracy) {
    return roundUp(signBit, exponentBits, mantissaBits, lostPrecision)
}

func roundUp(signBit, exponentBits, mantissaBits uint32,
    lostPrecision bool) (Bits, big.Accuracy) {
    
    mantissaF8E4M3Precision := mantissaBits & f32Float8E4M3MantissaMask
    mantissaExtraPrecision := mantissaBits & f32Float8E4M3HalfSubnormalMask

    float8E4M3Sign := uint32(signBit << 7)
    float8E4M3Exponent := uint32(exponentBits << 3)
    float8E4M3Mantissa := uint32(mantissaF8E4M3Precision >> 20)

    // For this rounding mode, we only need to add 1 to the least-precision
    // mantissa, if the input was positive, to bring it closer to +inf.
    // For negative numbers, this is achieved by simply truncating.

    exponentMantissaComposite := (float8E4M3Exponent | float8E4M3Mantissa)

    // If positive and there is extra precision, then add 1
    if (float8E4M3Sign == 0) && 
        (mantissaExtraPrecision != 0 || lostPrecision) {
        exponentMantissaComposite += 1
    }
    
    // Since we don't handle overlfow, all we need to do now is attach the sign
    resultVal := Bits(float8E4M3Sign | exponentMantissaComposite)

    resultAcc := big.Exact
    // If there was any extra precision left, then we need to update the
    // accuracy
    if mantissaExtraPrecision != 0 || lostPrecision {
        resultAcc = big.Above
    }

    return resultVal, resultAcc
}
