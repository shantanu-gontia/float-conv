package F8E4M3

import (
	"errors"
	"math"
	"math/big"
	"slices"

	floatBit "github.com/shantanu-gontia/float-conv/pkg"
	F32 "github.com/shantanu-gontia/float-conv/pkg/float32bits"
	F8E8M0 "github.com/shantanu-gontia/float-conv/pkg/float8e8m0"
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

// Convert the given [Bits] type to the floating point number it represents,
// inside a float32 value. This is effectively, a bit_cast to float8e4m3,
// followed by a upcast to float32. Since Go doesn't natively support
// float8e4m3 values, this method performs some bit-twiddling, 
// to align the bits per the float32 bit representation and then scaling
// the final result with the [F8E8M0.ScaleFactor]
func (input Bits) ToFloat32(scaleFactor F8E8M0.ScaleFactor) float32 {
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
func (input Bits) ToBigFloat(scaleFactor F8E8M0.ScaleFactor) big.Float {
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
func FromBigFloat(input big.Float, scaleFactor F8E8M0.ScaleFactor,
    rm floatBit.RoundingMode, om floatBit.OverflowMode,
    um floatBit.UnderflowMode) (Bits, big.Accuracy, floatBit.Status) {

    // Since the [big] package's methods do not support rounding modes for
    // direct conversion to float8e4m3. We convert to an intermediate float32
    // number and use our custom conversion function [FromFloat32] to convert
    // to [Bits]
    input.SetMode(big.ToZero)
    closestFloat32, fromBigFloatAcc := input.Float32()

    var asFloat32 float32
    // [big.Float.Float32] returns the float32 closest to the input.
    // This might cause it round UP for some cases.
    // But, we need to get the value with extra precision truncated.
    // To get the truncated result, we need to subtract 1 ULP of
    // precision if the number is positive and the float32 is larger,
    // or if the number is negative and the float32 is smaller, or
    // alternatively if [big.Float.Float32] returns [big.Above] as the
    // accuracy, because for truncation this should always be [big.Below]

    // Note that however, we need to exempt the case where the results
    // become infinity because that counts not as rounding but overflow.
    if math.IsInf(float64(closestFloat32), 1) && fromBigFloatAcc == big.Above {
        // if the input was greater then the float32 maximum normal, then
        // closestFloat32 would be +inf, and the accuracy returned would be
        // [big.Above]. To pass through the overflow handling to
        // [FromFloat32], we thus cap the infinities to the maximum normal
        // numbers in float32.
        // F32.PositiveMaxNormal will trigger overflow response
        // in F8E4M3
        asFloat32 = math.Float32frombits(F32.PositiveMaxNormal)
    } else if math.IsInf(float64(closestFloat32), -1) && fromBigFloatAcc == big.Below {
        asFloat32 = math.Float32frombits(F32.NegativeMaxNormal)
    } else if closestFloat32 == 0.0 && fromBigFloatAcc == big.Below {
        // Similar to the infinity case, we also need to make sure that the
        // underflow response handling is passed through to [FromFloat32].
        // So, if the closest float32 number results in 0 and the accuracy is
        // big.Below or +ve numbers, then we just pass in F32.PositiveMinSubnormal
        // which will trigger underflow in [FromFloat32]
        asFloat32 = math.Float32frombits(F32.PositiveMinSubnormal)
    } else if closestFloat32 == math.Float32frombits(F32.NegativeZero) &&
        fromBigFloatAcc == big.Above {
        asFloat32 = math.Float32frombits(F32.NegativeMinSubnormal)
    } else if (input.Sign() > 0 && fromBigFloatAcc == big.Above) ||
        (input.Sign() < 0 && fromBigFloatAcc == big.Below) {
        // For positive numbers if the accuracy was [big.Above], then
        // [big.Float.Float32] caused rounding away from zero. This is 
        // undesirable. To make it truncation we need to subtract 1 ULP
        // from the number
        closestFloat32Bits := math.Float32bits(closestFloat32)
        asFloat32 = math.Float32frombits(closestFloat32Bits - 1)
    } else {
        asFloat32 = closestFloat32
    }

    resultBits, resultAcc, resultStatus := FromFloat32(asFloat32, scaleFactor,
        rm, om, um)
    return resultBits, resultAcc, resultStatus
}


// Convert the given float32 number into a [Bits] type which represents the
// bits of a OCP MXFP8E4M3 number. Signature and usage is identical to
// [FromBigFloat] except the input is a float32
func FromFloat32(input float32, scaleFactor F8E8M0.ScaleFactor,
    rm floatBit.RoundingMode, om floatBit.OverflowMode,
    um floatBit.UnderflowMode) (Bits, big.Accuracy, floatBit.Status) {

    // we need to access the underlying bits of the float32 number
    asUint32 := math.Float32bits(input)

    // With the number interpreted as uint32, we can now extract the underlying
    // sign, exponent, and mantissa bits
    signBit := (asUint32 & F32.SignMask) >> 31
    exponentBits := (asUint32 & F32.ExponentMask) >> 23
    mantissaBits := asUint32 & F32.MantissaMask

    // Special Case #1:
    // scaleFactor is NaN. Then the result is also NaN
    if scaleFactor == F8E8M0.NaN {
        if signBit == 0 {
            return Bits(PositiveNaN), big.Exact, floatBit.Overflow
        }
        return Bits(NegativeNaN), big.Exact, floatBit.Overflow
    }

    // Special Case #2: Infinities -> NaN
    // F8E8M0 does not have infinites, so we return a NaN
    // and overflow
    if math.IsInf(float64(input), 1) {
        return Bits(PositiveNaN), big.Below, floatBit.Overflow
    }
    if math.IsInf(float64(input), -1) {
        return Bits(NegativeNaN), big.Above, floatBit.Overflow
    }

    // Special Case #3: NaN
    // NaNs always convert to NaNs. For our case,we consider the conversion
    // to be exact
    if math.IsNaN(float64(input)) {
        return Bits(NaN), big.Exact, floatBit.Fits
    }

    // Special Case #4: Zeros
    if asUint32 == F32.PositiveZero {
        return Bits(PositiveZero), big.Exact, floatBit.Fits
    }
    if asUint32 == F32.NegativeZero {
        return Bits(NegativeZero), big.Exact, floatBit.Fits
    }

    // Special Case #4: scaledExponent is smaller than the minimum float8e4m3
    // representable exponent. (Underflow)
    actualScaleFactor := int(scaleFactor) - int(F8E8M0.ExponentBias)
    actualExponent := int(exponentBits) - F32.ExponentBias
    scaledExponent := actualExponent - actualScaleFactor
    // These exponents correspond to the subnormals in float32.
    // They will underflow in float8e4m3
    if scaledExponent < F32.ExponentMin {
        return handleUnderflow(signBit, um)
    }

    // Special Case #5: scaledExponent is larger than the maximum float8e4m3
    // representable exponent. (Overflow)
    if scaledExponent > F32.ExponentMax {
        return handleOverflow(signBit, om)
    }

    // Special Case #6: Input exceeds the maximum normal value (in magnitude)
    // that can be represented in the float8e4m3 format. In this case, the
    // input om [floatBit.OverflowMode] determines the response
    if checkOverflow(scaledExponent, mantissaBits) {
        return handleOverflow(signBit, om)
    }

    // Variables to store the return values in
    var resultVal Bits
    var resultAcc big.Accuracy

    // To simplify the calculation, we need to calculate two quantities
    // 1. aligned mantissa - For normal numbers this is the same as the
    // original, but for subnormals we need to adjust it, because subnormals
    // don't have an implicit 1.0 addition like normals do.
    // 2. the adjusted exponent - For normal numbers, all we need to do is
    // subtract the float32 bias and apply the float8e4m3 bias. But for
    // subnormal numbers this should be exactly 0.
    alignedMantissa := mantissaBits
    adjustedExponent := uint32(scaledExponent + ExponentBias)

    // Value that indicates whether any precision was lost when preprocessing
    // the mantissa before passing it down to the rounding routines
    lostPrecision := false

    // Before performing any rounding, we need to make sure this exponent
    // can actually be represented in the float32 format. If the exponent
    // is smaller than the float8e4m3 format. If the exponent is smaller than
    // the minimum exponent allowed in float8e4m3 (-6), this either results
    // in underflow, or it rounds up or truncates to some subnormal number
    // in the float8e4m3 format. We will need to take special care for the
    // cases we truncate, because, we might underflow and we need to report
    // that.
    if actualExponent < ExponentMin {
        // We start with the assumption that the number can be represented
        // by a subnormal number in float8e4m3. Since subnormal numbers do not
        // have an implicit 1.0 addition), we add an implicit 1.0, to the
        // exponent LSB.
        const (
            float32ExponentLSB uint32 = 0b0_00000001_00000000000000000000000
        )

        // For subnormals, we need to appropriately calculate the aligned
        // mantissa to accout for the deficit of the implicit 1.0 addition
        alignedMantissa = mantissaBits | float32ExponentLSB

        // And also, for the subnormal case, we need to set the adjusted
        // exponent bits to 0
        adjustedExponent = 0

        // Now, to aligned the bits of the original format, with the mantissa
        // of the destination format as a subnormal, we have to shift right
        // until this implicit 1.0 addition falls to the mantissa bit
        // corresponding to the appropriate power of 2 (this is the largest
        // power of 2 in the number)
        //
        // Consider the following example. We have,
        // f32_n = 2^(-8) * (1 + m_0/2 + m_1/4 + m_2/8 + ...)
        //       = 2^(-6) * (1/4 + m_0/8 + m_1/16 + m_2/32 + ...)
        // 
        // Clearly, this caused the mantissa bit corresponding to 1/2 to
        // turn to 0 and m_0 which originally would have been for 1/2, now
        // corresponds to the 1/8 power. This is equivalent to a right-shift
        // by 2 = (-6 - (-8))
        //
        // In general, by following the pattern, this shift amount is equal
        // to the difference between the minimum representable exponent in
        // float8e4m3 and the actual value of the exponent in float32 after
        // scaling
        shiftAmount := uint32(ExponentMin - scaledExponent)
        for ; shiftAmount > 0; shiftAmount-- {
            lastDigit := alignedMantissa & 0x1
            if lastDigit == 1 {
                // There was precision lost due to shifting which wouldn't
                // be retained in the aligned mantissa. We need to track this
                // to record accuracy
                lostPrecision = true
            }
            alignedMantissa >>= 1
        }

        // Now that we have the value for the mantissa, we can determine
        // the underflow case. There is underflow in the case when the part of
        // the mantissa that has precision that can be represented in
        // float8e4m3 is 0, but the rest of the mantissa has atleast 1 bit set
        // i.e. all of the precision in the number is higher than that could
        // be represented in float8e4m3. In this case, the response is
        // handled by the input um [floatBit.UnderflowMode]
        if checkUnderflow(alignedMantissa, lostPrecision) {
            return handleUnderflow(signBit, um)
        }

    }

    // Now that the mantissa bits are properly placed, and exponents are
    // aligned and the overflow, underflow case is handled. We can handle the
    // normal -> normal, subnormal case by performing the correct rounding

    switch rm {
        case floatBit.RoundTowardsZero:
            resultVal, resultAcc = roundTowardsZero(signBit,
                adjustedExponent, alignedMantissa, lostPrecision)
        case floatBit.RoundTowardsNegativeInf:
            resultVal, resultAcc = roundTowardsNegativeInf(signBit,
                adjustedExponent, alignedMantissa, lostPrecision)
        case floatBit.RoundTowardsPositiveInf:
            resultVal, resultAcc = roundTowardsPositiveInf(signBit,
                adjustedExponent, alignedMantissa, lostPrecision)
        case floatBit.RoundHalfTowardsZero:
            resultVal, resultAcc = roundHalfTowardsZero(signBit,
                adjustedExponent, alignedMantissa, lostPrecision)
        case floatBit.RoundHalfTowardsNegativeInf:
            resultVal, resultAcc = roundHalfTowardsNegativeInf(signBit,
                adjustedExponent, alignedMantissa, lostPrecision)
        case floatBit.RoundHalfTowardsPositiveInf:
            resultVal, resultAcc = roundHalfTowardsPositiveInf(signBit,
                adjustedExponent, alignedMantissa, lostPrecision)
        case floatBit.RoundNearestEven:
            resultVal, resultAcc = roundNearestEven(signBit,
                adjustedExponent, alignedMantissa, lostPrecision)
        case floatBit.RoundNearestOdd:
            resultVal, resultAcc = roundNearestOdd(signBit,
                adjustedExponent, alignedMantissa, lostPrecision)
    }

    return resultVal, resultAcc, floatBit.Fits
}

// Utility function to check if the number with the given exponent and mantissa
// bits would overflow when trying to represent it in a float8e4m3 value
// with the given scale factor.
// mantissaBits should occupy the bits as they would in a float32 number
func checkOverflow(scaledExponent int, mantissaBits uint32) bool {
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

// Utility function to check if the number with the given mantissa
// bits would underflow when trying to represent it in a float8e4m3 value
// assuming that exponent is min. Subnormals require shifting the mantissa to
// align the exponents. This might cause loss of precision that cannot be
// detected by the mantissaBits alone as they are already shifted. The
// lostPrecision parameter helps us with that. If it's set to true then there
// was precision lost when the mantissa was being aligned.
func checkUnderflow(mantissaBits uint32, lostPrecision bool) bool {
    // This assumes that the exponent bits after scaling is 0, so any extra 
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


// Utility function that resturns the result for the case when
// the conversion results in overflow. Since float8e4m3 does not
// support infinites, the [floatBit.SaturateInf] returns NaN instead
func handleOverflow(signBit uint32, om floatBit.OverflowMode) (Bits,
    big.Accuracy, floatBit.Status) {
    switch om {
    case floatBit.SaturateInf:
        fallthrough
    case floatBit.MakeNaN:
        if signBit == 0 {
            return Bits(PositiveNaN), big.Above, floatBit.Overflow
        }
        return Bits(NegativeNaN), big.Below, floatBit.Overflow
    case floatBit.SaturateMax:
        if signBit == 0 {
            // The maximum normal in float8e4m3 is smaller than any
            // number this function will be invoked for
            return Bits(PositiveMaxNormal), big.Below, floatBit.Overflow
        }
        return Bits(NegativeMaxNormal), big.Above, floatBit.Overflow
    default:
        panic("Unsupported OverflowMode encountered")
    }
}

// Utility function that resturns the result for the case when
// the conversion results in underflow
func handleUnderflow(signBit uint32, um floatBit.UnderflowMode) (Bits,
big.Accuracy, floatBit.Status) {
    switch um {
    case floatBit.FlushToZero:
        if signBit == 0 {
            // Zero is less than any positive subnormal
            return Bits(PositiveZero), big.Below, floatBit.Underflow
        }
        return Bits(NegativeZero), big.Above, floatBit.Underflow
    case floatBit.SaturateMin:
        if signBit == 0 {
            // Min subnormal of float32 is larger than any float64 subnormal
            return Bits(PositiveMinSubnormal), big.Above, floatBit.Underflow
        }
        return Bits(NegativeMinSubnormal), big.Below, floatBit.Underflow
    default:
        panic("Unsupported UnderflowMode encountered")
    }
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
func (b* Bits) ConversionError(input* big.Float, scaleFactor F8E8M0.ScaleFactor) (
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

