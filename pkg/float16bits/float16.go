package F16

import (
	"errors"
	"math"
	"math/big"
	"slices"

	floatBit "github.com/shantanu-gontia/float-conv/pkg"
	F32 "github.com/shantanu-gontia/float-conv/pkg/float32bits"
)

// Some constants that will help with the bit manipulation we'll need to
// perform
const (
	// Mantissa bits retained in float16
	f32Float16MantissaMask uint32 = 0b0_00000000_11111111110000000000000
	// Mantissa bits not retained in float16
	f32Float16HalfSubnormalMask uint32 = 0b0_00000000_00000000001111111111111
	// LSB of float16 and rest of the extra precision
	f32Float16SubnormalMask uint32 = 0b0_00000000_00000000011111111111111
	// LSB of float16
	f32Float16SubnormalLSB uint32 = 0b0_00000000_00000000010000000000000
	// Most significant bit not retained in float16
	f32Float16HalfSubnormalLSB uint32 = 0b0_00000000_00000000001000000000000
)

// Alias type for uint16. This is used to represent the bits that make up a
// Float16 number. This type also comes with utility methods to support
// Floating point conversions with different Rounding Modes and Out of Bounds
// responses
type Bits uint16

// Convert the given [Bits] type to the floating point number it represents,
// inside a float32 value. This is effectively, a bit_cast to float16,
// followed by a upcast to float32. Since Go doesn't natively support
// float16 values, this method performs some bit-twiddling, to align the bits
// per the float32 bit representation
func (input Bits) ToFloat32() float32 {
	asUint16 := uint16(input)
	// Extract the Sign, Exponent and Mantissa
	signBit := (asUint16 & SignMask) >> 15
	exponentBits := (asUint16 & ExponentMask) >> 10
	mantissaBits := asUint16 & MantissaMask

	// Special values like Inf, NaN, need to be handled before applying the
	// general algorithm to calculate the number
	if exponentBits == 0b00000000000_11111 {
		if mantissaBits == 0 { // Inf
			if signBit == 0 {
				return float32(math.Inf(1))
			} else {
				return float32(math.Inf(-1))
			}
		} else { // NaN
			if signBit == 0 {
				return math.Float32frombits(F32.PositiveNaN)
			} else {
				return math.Float32frombits(F32.NegativeNaN)
			}
		}
	}
	if asUint16 == PositiveZero {
		return math.Float32frombits(F32.PositiveZero)
	}
	if asUint16 == NegativeZero {
		return math.Float32frombits(F32.NegativeZero)
	}

	// Variables to store the sign, exponent and mantissa bits that will
	// be used to construct the float32 number
	var float32SignBit, float32ExponentBits, float32MantissaBits uint32

	float32SignBit = uint32(signBit) << 31

	if exponentBits == 0 {
		// Subnormals in F16, are normals in F32. So, they need special handling
		// If the float16 number is subnormal, then it is evaluated as
		// (-1)^sign * 2^(1-15) * (m0/2 + m1/4 + ...)
		// (-1)^sign * 2^(-14) * (m0/2 + m1/4 + ...)
		// So, what we want to do is, find the first bit in the mantissa,
		// that is 1. This will become the implicit precision bit in the float32
		// value. Let's say it's m0. In that case, we have
		// (-1)^sign * 2^(-14) * (1/2 + m1/4 + ...)
		// = (-1)^sign * 2^(-15) * (1 + m1/2 + m2/4 + m3/8 + ...)
		// So, if the MSB set bit is m0, then the result exponent = Emin - 1
		// and, we need to shift the mantissa to the right when it's in the
		// float32 container. And, there as an extra mantissa left-shift by
		// 1
		// Let's now say it's m2. In that case, we have
		// (-1)^sign * 2^(-14) * (0/2 + 0/4 + 1/8 + m3/16 + ...)
		// = (-1)^sign * 2^(-17) * ([m2=1] + m3/2 + m4/4 + m5/8 + m6/16 + ...)
		// So, the result exponent = Emin - 3
		// and, in the float32 the mantissa is again shifted to the right by 1
		// bit. And, there is a mantissa left shift by 3

		// Now, we find the value we need to subtract from the minimum exponent
		// F32 mantissa starts at bit index 31 _ 30 29 28 27 26 25 24 23 _ 22
		currMantissaBitMask := uint16(0b0_00000_1000000000)
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
		// F32 has 23 mantissa bits, and F16 has 10. Therefore, to align the
		// bits, we need to shift to the left by 13 bits. To account for the
		// right shift mentioned above then, we can actually just shift left
		// by 12 bits instead of 13.
		float32MantissaBits = uint32(resultMantissaBits) << (13 + extraShift)
		float32ExponentBits = uint32(resultExponent+F32.ExponentBias) << 23
	} else {
		// For the normal case, all we need to do is correct the exponent to
		// use the bias of the float32 format
		float32MantissaBits = uint32(mantissaBits) << 13
		actualExponent := int(exponentBits) - ExponentBias
		float32ExponentBits = uint32(actualExponent+F32.ExponentBias) << 23
	}
	return math.Float32frombits(float32SignBit | float32ExponentBits |
		float32MantissaBits)
}

// Convert the given [Bits] type to a [big.Float] arbitrary precision
// floating-point number
func (input Bits) ToBigFloat() big.Float {
	asFloat32 := input.ToFloat32()
	asBigFloat := *big.NewFloat(float64(asFloat32))
	return asBigFloat
}

// Convert the given [big.Float] arbitrary precision floating-point number
// to a [Bits] type representing the bits of a float16 number. If the number
// cannot be represented in float16 format exactly, then the rounding mode,
// overflow mode and underflow mode decide the result. Returns the result
// [Bits], a [big.Accuracy] which encodes whether the result value was the same,
// larger or smaller than the input, and a [floatBit.Status] which encodes,
// whether the result fit in the [Bits], caused overflow or underflow.
func FromBigFloat(input big.Float, rm floatBit.RoundingMode,
	om floatBit.OverflowMode, um floatBit.UnderflowMode) (Bits,
	big.Accuracy, floatBit.Status) {

	// Since the [big] package's methods do not support rounding modes for
	// direct conversion to bfloat16. We convert to an intermediate [float32]
	// number and use our custom conversion functions [FromFloat32] to convert
	// to [Bits]
	input.SetMode(big.ToZero)
	closestFloat32, fromBigFloatAcc := input.Float32()

	var asFloat32 float32
	// big.Float.Float32() returns the float32 closest to the input.
	// This might cause it to round up for some cases.
	// But, we need to get the value with extra precision truncated
	// Therefore, to get the truncated result, we need to subtract 1 ULP of
	// precision if the number is positive and the float32 is larger, or
	// if the number is negative and the float32 is smaller, or alternatively
	// if the .Float32() returns big.Above as the accuracy, because for
	// truncation this should always be big.Below
	// Note that however, we need to exempt, the case where the results
	// becomes infinity.
	if math.IsInf(float64(closestFloat32), 1) && fromBigFloatAcc == big.Above {
		// If input was greater than F32 Maximum Normal, then closestFloat32
		// would be +inf, and the accuracy returned would be big.Above
		// F32.PositiveMaxNormal will trigger overflow repsonse in BF16
		asFloat32 = math.Float32frombits(F32.PositiveMaxNormal)
	} else if math.IsInf(float64(closestFloat32), -1) && fromBigFloatAcc == big.Below {
		// Similarly,
		// for -inf case, it will be big.Below.
		// F32.NegativeMaxNormal will trigger overflow response in BF16
		asFloat32 = math.Float32frombits(F32.NegativeMaxNormal)
	} else if closestFloat32 == 0.0 && fromBigFloatAcc == big.Below {
		// We also need to do this for the cases, where closestFloat32 is smaller
		// than the minimum float32 subnormal, again, because we want to handle
		// the underflow response in the BF16 methods.
		// F32.PositiveMinSubnormal will trigger underflow response in BF16
		asFloat32 = math.Float32frombits(F32.PositiveMinSubnormal)
	} else if closestFloat32 == -0.0 && fromBigFloatAcc == big.Above {
		// And for the negative case
		// F32.NegativeMinSubnormal will trigger underflow response in BF16
		asFloat32 = math.Float32frombits(F32.NegativeMinSubnormal)
	} else if (input.Sign() > 0 && fromBigFloatAcc == big.Above) ||
		(input.Sign() < 0 && fromBigFloatAcc == big.Below) {
		// For positive numbers if the accuracy was big.Above, then Float32()
		// caused rounding away from zero. This is undesirable. To make it
		// truncation we need to subtract 1 ULP from the number
		closestFloat32Bits := math.Float32bits(closestFloat32)
		asFloat32 = math.Float32frombits(closestFloat32Bits - 1)
	} else {
		asFloat32 = closestFloat32
	}

	resultBits, resultAcc, resultStatus := FromFloat32(asFloat32, rm, om, um)
	return resultBits, resultAcc, resultStatus
}

// Convert the given float32 number into a [Bits] type which represents the bits
// of a half-preicision floating point number. Signature and usage is identical
// to [FromBigFloat] except the parameter input is float32
func FromFloat32(input float32, rm floatBit.RoundingMode,
	om floatBit.OverflowMode, um floatBit.UnderflowMode) (Bits, big.Accuracy,
	floatBit.Status) {
	// There are some special cases we need to handle, particularly those for
	// values like Inf, NaN which have special encodings in the two formats

	// Special Case #1: Infinities
	if math.IsInf(float64(input), 1) {
		return Bits(PositiveInfinity), big.Exact, floatBit.Fits
	}
	if math.IsInf(float64(input), -1) {
		return Bits(NegativeInfinity), big.Exact, floatBit.Fits
	}

	// Special Case #2: NaNs
	// NaNs always convert to NaNs. For our case, we consider the converison
	// to be exact.
	if math.IsNaN(float64(input)) {
		return Bits(NaN), big.Exact, floatBit.Fits
	}

	// For the remaining cases, we need access to the underlying bits of the
	// float32 number. As such, we need to
	// first view the type as a uint32 format. This will allow us to
	// manipulate the underlying bits.
	asUint32 := math.Float32bits(input)

	// Special Case #3: Zeros
	if asUint32 == F32.PositiveZero {
		return Bits(PositiveZero), big.Exact, floatBit.Fits
	}
	if asUint32 == F32.NegativeZero {
		return Bits(NegativeZero), big.Exact, floatBit.Fits
	}

	// With the number interpreted as uint32, we can now extract the underlying
	// sign, exponent and mantissa bits.
	signBit := (asUint32 & F32.SignMask) >> 31
	exponentBits := (asUint32 & F32.ExponentMask) >> 23
	mantissaBits := asUint32 & F32.MantissaMask

	// Special Case #4: Input is subnormal (exponent bits == 0 && mantissa != 0)
	// If the input is subnormal in the source format, which has higher
	// precision, then it will underflow in the the format with the lower
	// precision. For these cases the input um [floatBit.UnderflowMode]
	// determines the result
	if exponentBits == 0 {
		return handleUnderflow(signBit, um)
	}

	// Special Case #5: Input exceeds the maximum normal value (in magnitude)
	// that can be represented in the float32 format. In this case, the input om
	// [floatBit.OverflowMode] determines the response.

	// First off, we calculate the actual value of the exponent. To do this,
	// we subtract the Exponent Bias from the bits that make up the exponent
	// in the input's bits
	actualExponent := int(exponentBits) - F32.ExponentBias

	if checkOverflow(actualExponent, mantissaBits) {
		return handleOverflow(signBit, om)
	}

	// Variables to store return values in
	var resultVal Bits
	var resultAcc big.Accuracy

	// To simplify the calculation, we need to calculate two quantities
	// 1. the aligned mantissa - For normal numbers this is the same
	// as the original, but for subnormals we need to adjust it, because
	// they don't have the implicit 1.0 addition
	// 2. the adjusted exponent - For normal numbers, all we need to do is
	// subtract the float32 bias and apply the float16 bias. But for subnormal
	// numbers this should exactly equal 0
	alignedMantissa := mantissaBits
	adjustedExponent := uint32(actualExponent + ExponentBias)

	// Before performing any rounding, we need to make sure this exponent
	// can actually be represented in the float32 format. If the exponent,
	// is smaller than the minimum exponent allowed in float32 (-126), this
	// either results in underflow, or it rounds up or trunc to some subnormal
	// number in the float32 format. We will need to take special care for
	// the cases where we truncate, because we might underflow and we need
	// to report that.
	if actualExponent < ExponentMin {

		// We start with the assumption that the number can be represented by
		// a subnormal number in float32. Since, subnormal numbers do not have
		// an implicit 1.0 addition (or an extra implicit precision bit,
		// however you want to interpret it). We add an implicit 1, to the
		// exponent LSB.
		const (
			float32ExponentLSB uint32 = 0b0_00000001_00000000000000000000000
		)

		// For subnormals, we need to appropriately calculate the aligned
		// mantissa to account for the deficit of the implcit 1.0 addition
		alignedMantissa = mantissaBits | float32ExponentLSB

		// And also, for the subnormal case, we need to set the adjusted
		// exponent bits to 0
		adjustedExponent = 0

		// For accurately determining underflow, we also need to check
		// if we shifted any 1s to the right
		lostPrecision := false

		// Now, to align the bits of the original format, with the mantissa
		// of the destination format as a subnormal, we have to shift right
		// until this implicit 1 addition, falls to the mantissa bit,
		// corresponding to the appropriate power of 2.
		//
		// To see what's going on, consider the example where, we have
		// f32_n = 2^(-16) * (1 + m_0/2 + m_1/4 + m_2/8 + ...)
		//       = 2^(-14) * (1 / 4 + m_0 / 8 + m_1 / 16 + m_2 / 32 ....)
		//
		// Clearly, this caused the mantissa bit corresponding to 1/2 turn to 0
		// and m_0 which originally would have been for 1/2, is now in the 1/8
		// power. This is equivalent to a right-shfit by 2 = (-14 - (-16))
		//
		// In general, by following the pattern, this shift amount is equal
		// to the difference between the minimum representable exponent (actual)
		// and the actual value of the exponent in float64.
		shiftAmount := uint32(ExponentMin - actualExponent)
		for ; shiftAmount > 0; shiftAmount-- {
			lastDigit := alignedMantissa & 0x1
			if lastDigit == 1 {
				lostPrecision = true
			}
			alignedMantissa >>= 1
		}

		// Now that we have the value for the mantissa, we can determine
		// the underflow case. There is underflow, in the case when the
		// part of the mantissa that has the precision that can be represented
		// in float32 is 0 (bits m52 to m30), but the rest of the mantissa has
		// atleast 1 bit set i.e. all of the precision in the number is higher
		// than that could be represented in float32. In this case, the response
		// is handled by the input um [floatBit.UnderflowMode]
		if checkUnderflow(alignedMantissa, lostPrecision) {
			return handleUnderflow(signBit, um)
		}

	}

	// Now that the mantissa bits are properly placed, and exponents are aligned
	// and the overfow, underflow case is handled, we can handle the general
	// normal, subnormal case by performing the correct rounding.

	// Finally, we get to the case, where the resulting float32 number is also
	// a normal number.
	switch rm {
	case floatBit.RoundTowardsZero:
		resultVal, resultAcc = roundTowardsZero(signBit,
			adjustedExponent, alignedMantissa)
	case floatBit.RoundTowardsNegativeInf:
		resultVal, resultAcc = roundTowardsNegativeInf(signBit,
			adjustedExponent, alignedMantissa)
	case floatBit.RoundTowardsPositiveInf:
		resultVal, resultAcc = roundTowardsPositiveInf(signBit,
			adjustedExponent, alignedMantissa)
	case floatBit.RoundHalfTowardsZero:
		resultVal, resultAcc = roundHalfTowardsZero(signBit, adjustedExponent,
			alignedMantissa)
	case floatBit.RoundHalfTowardsNegativeInf:
		resultVal, resultAcc = roundHalfTowardsNegativeInf(signBit,
			adjustedExponent, alignedMantissa)
	case floatBit.RoundHalfTowardsPositiveInf:
		resultVal, resultAcc = roundHalfTowardsPositiveInf(signBit,
			adjustedExponent, alignedMantissa)
	case floatBit.RoundNearestEven:
		resultVal, resultAcc = roundNearestEven(signBit, adjustedExponent,
			alignedMantissa)
	case floatBit.RoundNearestOdd:
		resultVal, resultAcc = roundNearestOdd(signBit, adjustedExponent,
			alignedMantissa)
	}

	return resultVal, resultAcc, floatBit.Fits
}

// Utility function to check if the number with the given exponent and mantissa
// bits would overflow when trying to represent it in a float16 value
// exponentBits should correspond to bits which are encoded with the float16
// bias in mind. mantissaBits should occupy the bits with the float32 format
// in mind.
func checkOverflow(actualExponent int, mantissaBits uint32) bool {
	// If the exponent is larger than the max, then it's overflow
	if actualExponent > ExponentMax {
		return true
	}

	// If the exponent is equal to the maximum exponent, all the
	// float32 mantissa bits are set, but there is
	// additional precision in the number than can be represented in float32
	// then it exceeds the maximum normal and overflows.
	if (actualExponent == ExponentMax) &&
		(mantissaBits&f32Float16MantissaMask == f32Float16MantissaMask) &&
		(mantissaBits&f32Float16HalfSubnormalMask > 0) {
		return true
	}
	return false
}

// Utility function to check if the number with the given exponent and mantissa
// bits would overflow when trying to represent it in a float16 value
// Subnormals require shifting the mantissa to align the exponents. This might
// cause loss of precision that cannot be detected by mantissaBits alone as
// they are already shifted. The lostPrecision parameter helps us with that.
// If it's true then there was precision lost when mantissa was being aligned
func checkUnderflow(mantissaBits uint32, lostPrecision bool) bool {
	// This assumes that the exponent is 0, so any extra precision in the
	// mantissa means underflow.
	f16PrecisionMantissa := mantissaBits & f32Float16MantissaMask
	f16ExtraPrecisionMantissa := mantissaBits & f32Float16HalfSubnormalMask
	if (f16PrecisionMantissa == 0) && (f16ExtraPrecisionMantissa != 0 || lostPrecision) {
		return true
	}
	return false
}

// Utility function that returns the result for the case when
// the converison results in underflow
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

// Utility function that returns the result for the case when the conversion
// results in overflow
func handleOverflow(signBit uint32, om floatBit.OverflowMode) (Bits,
	big.Accuracy, floatBit.Status) {
	switch om {
	case floatBit.SaturateInf:
		if signBit == 0 {
			// +Inf is greater than any other normal number in float64
			return Bits(PositiveInfinity), big.Above, floatBit.Overflow
		}
		return Bits(NegativeInfinity), big.Below, floatBit.Overflow
	case floatBit.MakeNaN:
		// The accuracy and status don't matter for this case.
		if signBit == 0 {
			return Bits(PositiveNaN), big.Above, floatBit.Overflow
		}
		return Bits(NegativeNaN), big.Below, floatBit.Overflow
	case floatBit.SaturateMax:
		if signBit == 0 {
			// The maximum normal in float32 is smaller than any number
			// this function will be invoked for
			return Bits(PositiveMaxNormal), big.Below, floatBit.Overflow
		}
		return Bits(NegativeMaxNormal), big.Above, floatBit.Overflow
	default:
		panic("Unsupported OverflowMode encountered")
	}
}

// ToFloatFormat converts the given Bits type representing the bits that make
// up a bfloat16 number into [floatBit.FloatBitFormat]
// Implements the FloatBitFormatter Interface
func (b *Bits) ToFloatFormat() floatBit.FloatBitFormat {
	// Iterate over the bits and construct the return Values

	asUint := uint16(*b)
	signBits := (asUint & SignMask) >> 15
	exponentBits := (asUint & ExponentMask) >> 7
	mantissaBits := asUint & MantissaMask

	// 1 Sign Bit
	signRetVal := make([]byte, 0, 1)
	if signBits == 0 {
		signRetVal = append(signRetVal, byte('0'))
	} else {
		signRetVal = append(signRetVal, byte('1'))
	}

	// 5 Exponent Bits
	exponentRetVal := make([]byte, 0, 5)
	for i := 0; i < 5; i++ {
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

	// 10 Mantissa Bits
	mantissaRetVal := make([]byte, 0, 10)
	for i := 0; i < 10; i++ {
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

	return floatBit.FloatBitFormat{Sign: signRetVal,
		Exponent: exponentRetVal, Mantissa: mantissaRetVal}
}

// Conversion error returns the difference between the input [big.Float]
// number and the float32 number represented by the bits in the [Bits] receiver
func (b *Bits) ConversionError(input *big.Float) (big.Float, error) {
	// If the receiver is a NaN then we return an error
	asFloat32 := b.ToFloat32()
	if math.IsNaN(float64(asFloat32)) {
		return *big.NewFloat(0.0), errors.New("NaN encountered")
	}

	// Positive Infinity == Positive Infinity
	if math.IsInf(float64(asFloat32), 1) &&
		(input.IsInf() && (input.Sign() > 0)) {
		return *big.NewFloat(0), nil
	}

	// Negative Infinity == Negative Infinity
	if math.IsInf(float64(asFloat32), -1) &&
		(input.IsInf() && (input.Sign() < 0)) {
		return *big.NewFloat(0), nil
	}

	asBigFloat := b.ToBigFloat()
	convDiff := asBigFloat.Sub(&asBigFloat, input)
	return *convDiff, nil
}
