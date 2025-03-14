package F8E4M3

import "math/big"

// Utility function that returns the number truncated to a number that can
// be represented as a float8e4m3 number.
// signBit, and exponentBits must be passed with their values shifted all the
// way to the right.
// exponentBits must be passed with the float8e4m3 bias and appropriate
// scale-factor applied
// mantissaBits must be passed in their float32 locations.
// NOTE: This doesn't handle the underflow and overflow cases.

func roundTowardsZero(signBit, exponentBits, mantissaBits uint32,
	lostPrecision bool) (Bits, big.Accuracy) {

    mantissaF8E4M3Precision := mantissaBits & f32Float8E4M3MantissaMask
    mantissaExtraPrecision := mantissaBits & f32Float8E4M3HalfSubnormalMask

    float8E4M3Sign := uint8(signBit << 7)
    float8E4M3Exponent := uint8(exponentBits << 3)
    float8E4M3Mantissa := uint8(mantissaF8E4M3Precision >> 20)

    resultVal := Bits(float8E4M3Sign | float8E4M3Exponent | float8E4M3Mantissa)
    resultAcc := big.Exact

    // If there was extra precision, then the number did not fit in the
    // float8e4m3 format, so we nede to report the status appropriately
    if mantissaExtraPrecision != 0 || lostPrecision {
        if signBit == 0 {
            resultAcc = big.Below
        } else {
            resultAcc = big.Above
        }
    }

    return resultVal, resultAcc

}

