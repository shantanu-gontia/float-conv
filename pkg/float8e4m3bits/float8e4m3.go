package F8E4M3

import (
	"errors"
	"math"
	"math/big"
	"slices"

	floatBit "github.com/shantanu-gontia/float-conv/pkg"
	F32 "github.com/shantanu-gontia/float-conv/pkg/float32bits"
)

// Some constants that will help with bit manipulation we'll need to
// perform
const (
    // Mantissa bits retained in float8e4m3  
    f32Float8E4M3MantissaMask uint32 = 0b0_00000000_11100000000000000000000
    // Mantissa bits not retained in float8e4m3
    f32Float8E4M3HalfSubnormalMask uint32 = 
        0b0_00000000_00011111111111111111111
    // LSB of float8e4m3 and the rest of the extra precision
    f32Float8E4M3SubnormalMask uint32 = 0b0_00000000_00111111111111111111111
    // LSB of float8e4m3
    f32Float8E4M3SubnormalLSB uint32 = 0b0_00000000_00100000000000000000000
    // Most significant bit not retained in float8e4m3
    f32Float8E4M3HalfSubnormalLSB uint32 = 0b0_00000000_00010000000000000000000
)

// Alias type for uint8. This is used to represent the bits that make up a
// Float8E4M3 number. This type also comes with utility methods to support
// Floating point conversions with different Rounding Modes and Out of Bounds
// responses
type Bits uint8

// Alias type for uint8. This is used to represent the FP8E8M0 format which
// is used to specify the scale factor for a OCP MXFP8 floating point format.
// The actual floating point represented by an OCP MXPF8 number is the
// value encoded in the bits scaled by 2^(scale_factor) when scale_factor
// is interpreted as an fp8e8m0 number. For example, to apply no scaling,
// the number passed must be 127, because the actual scale that is
// multiplied is 2^(scale_factor - 127). If the exponent in the resulting
// number is out of the range supported by the format, then the result is
// undefined per the spec. In our case, we clamp to Infinity of the same sign
// as the original number
type ScaleFactor uint8

// Convert the given [Bits] type to the floating point number it represents,
// inside a float32 value. This is effectively, a bit_cast to float8e4m3,
// followed by a upcast to float32. Since Go doesn't natively support
// float8e4m3 values, this method performs some bit-twiddling, 
// to align the bits per the float32 bit representation and then scaling
// the final result with the [ScaleFactor]
func (input Bits) ToFloat32(scaleFactor ScaleFactor) float32 {
    asUint8 := uint8(input)
    // Extract the Sign, Exponent and Mantissa
    signBit := (asUint8 & SignMask) >> 7
    exponentBits := (asUint8 & ExponentMask) >> 3
    mantissaBits := (asUint8 & MantissaMask)

    // in F8E8M0, 255 signals NaN
    if (scaleFactor == 255) {
        if (signBit == 0) {
            return math.Float32frombits(F32.PositiveNaN)
        }
        return math.Float32frombits(F32.NegativeNaN)
    }

    // Special Values like Inf, NaN etc. need to be handled before applying
    // the general algorithm to calculate the number
    if asUint8 == PositiveNaN {
        return math.Float32frombits(F32.PositiveNaN)
    }
    if asUint8 == NegativeNaN {
        return math.Float32frombits(F32.NegativeNaN)
    }
    if asUint8 == PositiveZero {
        return math.Float32frombits(F32.PositiveZero)
    }
    if asUint8 == NegativeZero {
        return math.Float32frombits(F32.NegativeZero)
    }

    // Variables to store the sign, exponent and mantissa bits that will be
    // used to construct the float32 number
    var float32SignBit, float32ExponentBits, float32MantissaBits uint32

    float32SignBit = uint32(signBit) << 31
    float32Exponent := 0

    if exponentBits == 0 {
        // Subnormals in F8E4M3, are normals in F32. So, they need special
        // handling. If the float8e4m3 number is subnornmal, then it is
        // evaluated as (assuming default scale-factor)
        // (-1)^sign * 2^(1-7) * (m0/2 + m1/4 + m2/8)
        // = (-1)^sign * 2^-6 * (m0/2 + m1/4 + m2/8)
        // So, what we want to do is , find the first bit in the mantissa that
        // is 1. This will become the implicit precision bit in the float32
        // value. Let's suppose it is m0. In that case, we have
        // (-1)^sign * 2^-6 * (1/2 + m1/4 + m2/8)
        // = (-1)^sign * 2^-7 * (1 + m1/2 + m2/4)
        // So, if the MSB set bit is m0, then the result exponent = Emin - 1
        // and, we need to shift the mantissa to the right when it's in the
        // float32 container. And, there is an extra mantissa left-shift by 1
        // Suppose now it's m2. In that case, we have
        // (-1)^sign * 2^(-6) * (0/2 + 0/4 + 1/8)
        // = (-1)^sign * 2^(-9) * ([m2=1])
        // So, the result exponent is Emin - 3 and, in the float32, the
        // mantissa is shifted to the left by 3 bits.

        // We find the value we need to subtract from the minimum exponent
        // F32 mantissa starts at the bit index 31 and ends at 23 (inclusive)
        // 31 | 30 29 28 27 26 25 24 23 | 22 ...
        currMantissaBitMask := uint8(0b0_0000_100)
        resultMantissaBits := mantissaBits
        resultExponent := ExponentMin
        extraShift := 0
        for ; currMantissaBitMask != 0; currMantissaBitMask >>= 1 {
            currMantissaBit := currMantissaBitMask & mantissaBits
            resultExponent -= 1
            extraShift++
            if currMantissaBit != 0 {
                // We need to zero out this one bit, since this is what
                // becomes the implicit bit in the float32
                resultMantissaBits = mantissaBits & ^currMantissaBitMask
                break
            }
        }

        // F32 has 23 mantissa bits, and F8E4M3 has 3. Therefore, to align the
        // bits, we need to shift to the left by 20 bits. To account for the
        // right shift (since in float32 the first set bit is implicit),
        // we can actually just shift left by 19 bits instead of 20.
        float32MantissaBits = uint32(resultMantissaBits) << (20 + extraShift)
        float32Exponent = resultExponent
    } else {
        // For the normal case, all we need to do is correct the exponent to
        // use the bias of the float32 format
        float32MantissaBits = uint32(mantissaBits) << 20
        actualExponent := int(exponentBits) - ExponentBias
        float32Exponent = actualExponent
    }

    // Apply the Scale Factor
    float32Exponent = float32Exponent + (int(scaleFactor) - 127)
    if (float32Exponent > F32.ExponentMax) {
        return math.Float32frombits(float32SignBit | F32.PositiveInfinity)
    } else if (float32Exponent < F32.ExponentMin) {
        return math.Float32frombits(float32SignBit | F32.PositiveZero)
    }

    float32ExponentBits = uint32(float32Exponent + F32.ExponentBias) << 23
    return math.Float32frombits(float32SignBit | float32ExponentBits |
        float32MantissaBits)
}

// Convert the given [Bits] type to a [big.Float] arbitrary precision
// floating-point number with the given scale factor (Note that
// this works by intermediate conversion to float32, so if the number
// is not representable in float32 with the given scale factor, the result
// is same as what would be after conversion to float32
func (input Bits) ToBigFloat(scaleFactor ScaleFactor) big.Float {
    asFloat32 := input.ToFloat32(scaleFactor)
    asBigFloat := *big.NewFloat(float64(asFloat32))
    return asBigFloat
}

// Convert the given [big.Float] arbitrary-precison floating-point number
// to a [Bits] type representing the bits of a float8e4m3 number. The input
// is scaled by the given scaleFactor. If the number cannot be represented
// in the float8e4m3 format exactly, then the rounding mode, overflow mode,
// and underflow mode decide the result. Returns the result [Bits],
// a [big.Accuracy] which encodes whether the result value was the same,
// larger, or smaller than the input, and a [floatBit.Status] which encodes,
// whether the result ft in the [Bits], caused overflow, or underflow.
func FromBigFloat(input big.Float, scaleFactor ScaleFactor,
    rm floatBit.RoundingMode, om floatBit.OverflowMode,
    um floatBit.UnderflowMode) (Bits, big.Accuracy, floatBit.Status) {
    return Bits(0), big.Exact, floatBit.Fits
}

// Convert the given float32 number into a [Bits] type which represents the
// bits of a OCP MXFP8E4M3 number. Signature and usage is identical to
// [FromBigFloat] except the input is a float32
func FromFloat32(input float32, scaleFactor ScaleFactor,
    rm floatBit.RoundingMode, om floatBit.OverflowMode,
    um floatBit.UnderflowMode) (Bits, big.Accuracy, floatBit.Status) {
    return Bits(0), big.Exact, floatBit.Fits
}

// Utility function to check if the number with the given exponent and mantissa
// bits would overflow when trying to represent it in a float8e4m3 value
// with the given scale factor.
// mantissaBits should occupy the bits as they would in a float32 number
func checkOverflow(actualExponent int, mantissaBits uint32,
    scaleFactor ScaleFactor) bool {

    // Remove the scaling
    scaledExponent := actualExponent - (int(scaleFactor) - 127)

    if scaledExponent > ExponentMax {
        return true
    }

    // If the exponent is equal to the maximum exponent, and all the float32
    // mantissa bits are set, but there is additional precision in the number
    // than can be represented in float32, then it exceeds the maximum normal
    // and so, overflows.
    if (scaledExponent == ExponentMax) &&
        (mantissaBits & 0b0_00000000_11000000000000000000000 ==
        0b0_00000000_11000000000000000000000) && 
        (mantissaBits & f32Float8E4M3SubnormalMask > 0){
        return true
    }
    return false
}

// Utility function to check if the number with the given exponent and mantissa
// bits would underflow when trying to represent it in a float8e4m3 value
// with the given scaleFactor. Subnormals require shifting the mantissa to
// align the exponents. This might cause loss of precision that cannot be
// detected by the mantissaBits alone as they are already shifted. The
// lostPrecision parameter helps us with that. If it's set to true then there
// was precision lost when the mantissa was being aligned.
func checkUnderflow(mantissaBits uint32, lostPrecision bool) bool {
    // This assumes that the exponent after scaling  is 0, so any extra 
    // precision in the mantissa means underflow
    f8e4m3PrecisionMantissa := mantissaBits & f32Float8E4M3MantissaMask
    f8e4m3ExtraPrecisionMantissa := mantissaBits & 
        f32Float8E4M3HalfSubnormalMask
    if f8e4m3PrecisionMantissa == 0 && (f8e4m3ExtraPrecisionMantissa != 0 ||
        lostPrecision) {
        return true
    }
    return false
}

// ToFloatFormat converts the given [Bits] type representing the bits that
// make up a OCP MXFP8E4M3 umber into [floatBit.FloatBitFormat]
// Implements the FloatBitFormatter interface
func (b* Bits) ToFloatFormat() floatBit.FloatBitFormat {
    // Iterate over the bits and construct the return values

    asUint := uint8(*b)
    signBits := (asUint & SignMask) >> 7
    exponentBits := (asUint & ExponentMask) >> 3
    mantissaBits := (asUint & MantissaMask)

    // 1 Sign Bit
    signRetVal := make([]byte, 0, 1)
    if signBits == 0 {
        signRetVal = append(signRetVal, byte('0'))
    } else {
        signRetVal = append(signRetVal, byte('1'))
    }

    // 4 Exponent Bits
    exponentRetVal := make([]byte, 0, 4)
    for range 5 {
        currentExponentBit := exponentBits & 0x1
        var valueToAppend byte
        if currentExponentBit == 0 {
            valueToAppend = '0'
        } else {
            valueToAppend = '1'
        }
        exponentRetVal = append(exponentRetVal, valueToAppend)
        exponentBits >>= 1
    }

    // 3 Mantissa Bits
    mantissaRetVal := make([]byte, 0, 3)
    for range 3 {
        currentMantissaBit := mantissaBits & 0x1
        var valueToAppend byte
        if currentMantissaBit == 0 {
            valueToAppend = '0'
        } else {
            valueToAppend = '1'
        }
        mantissaRetVal = append(mantissaRetVal, valueToAppend)
        mantissaBits >>= 1
    }
    slices.Reverse(mantissaRetVal)

    return floatBit.FloatBitFormat{Sign: signRetVal, Exponent: exponentRetVal,
    Mantissa: mantissaRetVal}
}

// Conversion error returns the difference between the input [big.Float]
// number and the float8e4m3 number represented by the bits in the [Bits]
// receiver when it's scaled with the given scale factor
func (b* Bits) ConversionError(input* big.Float, scaleFactor ScaleFactor) (
    big.Float, error) {
    // If the receiver is a NaN then we return an error
    asFloat32 := b.ToFloat32(scaleFactor)
    if math.IsNaN(float64(asFloat32)) {
        return *big.NewFloat(0.0), errors.New("NaN encountered")
    }

    // Positive Infinity == Positiive Infinity
    if math.IsInf(float64(asFloat32), 1) && 
        (input.IsInf() && input.Sign() > 0) {
        return *big.NewFloat(0), nil
    }

    // Negative Infinity == Negative Infinity
    if math.IsInf(float64(asFloat32), -1) &&
        (input.IsInf() && input.Sign() < 0) {
        return *big.NewFloat(0), nil
    }

    asBigFloat := b.ToBigFloat(scaleFactor)
    convDiff := asBigFloat.Sub(&asBigFloat, input)
    return *convDiff, nil
}

