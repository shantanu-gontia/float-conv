package F32

import (
	"errors"
	"math"
	"math/big"

	floatBit "github.com/shantanu-gontia/float-conv/pkg"
)

// Some constants that will help with the bit manipulation we'll need to
// perform
const (
	// Mantissa bits retained in float32
	f64float32MantissaMask uint64 = 0x000f_ffff_e000_0000
	// Mantissa bits not retained in float32
	f64float32HalfSubnormalMask uint64 = 0x0000_0000_1fff_ffff
	// LSB of float32 and rest of the extra precision
	f64float32SubnormalMask uint64 = 0x0000_0000_3fff_ffff
	// LSB of float32
	f64float32SubnormalLSB uint64 = 0x0000_0000_2000_0000
	// Most significant bit not retained in float32
	f64float32HalfSubnormalLSB uint64 = 0x0000_0000_1000_0000

	f64SignBitMask      uint64 = 0x8000_0000_0000_0000
	f64ExponentBitsMask uint64 = 0x7ff0_0000_0000_0000
	f64MantissaBitsMask uint64 = 0x000f_ffff_ffff_ffff

	f64ExponentBias int = 1023
)

// Alias type for uint32. This is used to represent the bits which make up
// a IEEE-754 binary32 number. This type also comes with additional utility
// methods to support Floating point conversions with different Rounding
// Modes and Out of Bounds Responses.
type Bits uint32

// Convert a float32 number to its constituent bits
// This is effectively just a bit_cast from float32 to uint32
func FromFloat32(input float32) Bits {
	asBits := math.Float32bits(input)
	return Bits(asBits)
}

// Convert the given [Bits] type to the [float32] number it represents
// This is effectively just a bit_cast from [uint32] to [float32]
func (input Bits) ToFloat32() float32 {
	return math.Float32frombits(uint32(input))
}

// Convert the given [Bits] type to a [big.Float] arbitrary precision
// floating-point number
func (input Bits) ToBigFloat() big.Float {
	asBigFloat := *big.NewFloat(float64(input.ToFloat32()))
	return asBigFloat
}

// Convert the given [big.Float] arbitrary precision floating-point number
// to a [Bits] type representing the bits of a [float32] number. If the number
// cannot be represented in [float32] format exactly, then the rounding mode,
// overflow mode and underflow mode decide the result. Returns the result
// [Bits], a [big.Accuracy] which encodes whether the result value was the same,
// larger () or smaller than the input, and a [floatBit.Status] which encodes,
// whether the result fit in the [Bits], caused overflow or underflow.
func FromBigFloat(input big.Float, rm floatBit.RoundingMode,
	om floatBit.OverflowMode, um floatBit.UnderflowMode) (Bits,
	big.Accuracy, floatBit.Status) {

	// We need to get the value with extra precision truncated
	// But, Float64() returns the float64 value that is closest to the input.
	// Therefore, to get the truncated result, we need to subtract 1 ULP of
	// precision if the number is positive and the float64 is larger, or
	// if the number is negative and the float64 is smaller
	// Since the [big] package's methods do not support rounding modes for
	// direct conversion to float32 with all the other metadata we want.
	// We convert to an intermediate float64 number and use our custom
	// conversion functions [FromFloat32] to convert to [Bits]
	closestFloat64, fromBigFloatAcc := input.Float64()

	var asFloat64 float64
	// Float64() returns the float64 value that is closest to the input.
	// This might cause it to round up for some cases.
	// But, we need to get the value with extra precision truncated
	// Therefore, to get the truncated result, we need to subtract 1 ULP of
	// precision if the number is positive and the float64 is larger, or
	// if the number is negative and the float64 is smaller, or alternatively
	// if the .Float64() returns big.Above as the accuracy, because for
	// truncation this should always be big.Below
	// Note that however, we need to exempt, the case where the results
	// becomes infinity.
	if math.IsInf(closestFloat64, 1) && fromBigFloatAcc == big.Above {
		// If input was greater than float64 Maximum Normal, then closestFloat64
		// would be +inf, and the accuracy returned would be big.Above
		// MaxFloat64 will trigger overflow repsonse in F32
		asFloat64 = math.MaxFloat64
	} else if math.IsInf(closestFloat64, -1) && fromBigFloatAcc == big.Below {
		// Similarly,
		// for -inf case, it will be big.Below.
		// -MaxFloat64 will trigger overflow response in F32
		asFloat64 = -math.MaxFloat64
	} else if closestFloat64 == 0.0 && fromBigFloatAcc == big.Below {
		// We also need to do this for the cases, where closestFloat64 is smaller
		// than the minimum float64 subnormal, again, because we want to handle
		// the underflow response in the F32 methods.
		// SmallestNonzeroFloat64 will trigger underflow response in F32
		asFloat64 = math.SmallestNonzeroFloat64
	} else if closestFloat64 == -0.0 && fromBigFloatAcc == big.Above {
		// And for the negative case
		// F32.NegativeMinSubnormal will trigger underflow response in BF16
		asFloat64 = -math.SmallestNonzeroFloat64
	} else if (input.Sign() > 0 && fromBigFloatAcc == big.Above) ||
		(input.Sign() < 0 && fromBigFloatAcc == big.Below) {
		// For positive numbers if the accuracy was big.Above, then Float32()
		// caused rounding away from zero. This is undesirable. To make it
		// truncation we need to subtract 1 ULP from the number
		closestFloat64Bits := math.Float64bits(closestFloat64)
		asFloat64 = math.Float64frombits(closestFloat64Bits - 1)
	} else {
		asFloat64 = closestFloat64
	}

	resultBits, resultAcc, resultStatus := FromFloat64(asFloat64, rm, om, um)
	return resultBits, resultAcc, resultStatus
}

// Convert the given [float64] number to a [Bits] type which represents the bits
// of a [float32] number. Signature and usage is identical to [FromBigFloat],
// except the parameter input for this function is [float64]
func FromFloat64(input float64, rm floatBit.RoundingMode,
	om floatBit.OverflowMode, um floatBit.UnderflowMode) (Bits,
	big.Accuracy, floatBit.Status) {

	// There are some special cases we need to handle, particularly, the special
	// values like Inf, NaN which have special encodings in the two formats

	// Special Case #1: Infinities
	// Both float64 and float32 have representations for positive and negative
	// infinities. Both of them convert to their counterparts exactly.
	if math.IsInf(input, 1) {
		return Bits(PositiveInfinity), big.Exact, floatBit.Fits
	}
	if math.IsInf(input, -1) {
		return Bits(NegativeInfinity), big.Exact, floatBit.Fits
	}

	// Special Case #2: NaNs
	// NaNs always convert to NaNs. For our case, we consider the converison
	// to be exact.
	if math.IsNaN(input) {
		return Bits(NaN), big.Exact, floatBit.Fits
	}

	// For the remaining cases, we need access to the underlying bits of the
	// float64 number. As such, we need to
	// first view the type as a uint64 format. This will allow us to
	// manipulate the underlying bits.
	asUint64 := math.Float64bits(input)

	// Special Case #3: Zeros
	// Both Negative and Positive Zeros convert to their counterparts exactly
	if asUint64 == 0x8000_0000_0000_0000 {
		return Bits(0x8000_0000), big.Exact, floatBit.Fits
	}
	if asUint64 == 0x0000_0000_0000_0000 {
		return Bits(0x0000_0000), big.Exact, floatBit.Fits
	}

	// With the number interpreted as uint64, we can now extract the underlying
	// sign, exponent and mantissa bits.
	signBit := (asUint64 & f64SignBitMask) >> 63
	exponentBits := (asUint64 & f64ExponentBitsMask) >> 52
	mantissaBits := asUint64 & f64MantissaBitsMask

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
	actualExponent := int(exponentBits) - f64ExponentBias

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
	// subtract the float64 bias and apply the float32 bias. But for subnormal
	// numbers this should exactly equal 0
	alignedMantissa := mantissaBits
	adjustedExponent := uint64(actualExponent + ExponentBias)

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
			Float64ExponentLSB uint64 = 0x0010_0000_0000_0000
		)

		// For subnormals, we need to appropriately calculate the aligned
		// mantissa to account for the deficit of the implcit 1.0 addition
		alignedMantissa = mantissaBits | Float64ExponentLSB

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
		// f64_n = 2^(-128) * (1 + m_0/2 + m_1/4 + m_2/8 + ...)
		//       = 2^(-126) * (1 / 4 + m_0 / 8 + m_1 / 16 + m_2 / 32 ....)
		//
		// Clearly, this caused the mantissa bit corresponding to 1/2 turn to 0
		// and m_0 which originally would have been there is now in the 1/8
		// power. This is equivalent to a right-shfit by 2 = (-126 - (-128))
		//
		// In general, by following the pattern, this shift amount is equal
		// to the difference between the minimum representable exponent (actual)
		// and the actual value of the exponent in float64.
		shiftAmount := uint64(ExponentMin - actualExponent)
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
// bits would overflow when trying to represent it in a float32 value
// exponentBits should correspond to bits which are encoded with the float32
// bias in mind. mantissaBits should occupy the bits with the float64 format
// in mind.
func checkOverflow(actualExponent int, mantissaBits uint64) bool {
	// If the exponent is larger than the max, then it's overflow
	if actualExponent > ExponentMax {
		return true
	}

	// If the exponent is equal to the maximum exponent, all the
	// float32 mantissa bits are set, but there is
	// additional precision in the number than can be represented in float32
	// then it exceeds the maximum normal and overflows.
	if (actualExponent == ExponentMax) &&
		(mantissaBits&f64float32MantissaMask == f64float32MantissaMask) &&
		(mantissaBits&f64float32HalfSubnormalMask > 0) {
		return true
	}
	return false
}

// Utility function to check if the number with the given exponent and mantissa
// bits would overflow when trying to represent it in a float32 value
// Subnormals require shifting the mantissa to align the exponents. This might
// cause loss of precision that cannot be detected by mantissaBits alone as
// they are already shifted. The lostPrecision parameter helps us with that.
// If it's true then there was precision lost when mantissa was being aligned
func checkUnderflow(mantissaBits uint64, lostPrecision bool) bool {
	// This assumes that the exponent is 0, so any extra precision in the
	// mantissa means underflow.
	f32PrecisionMantissa := mantissaBits & f64float32MantissaMask
	f32ExtraPrecisionMantissa := mantissaBits & f64float32HalfSubnormalMask
	if (f32PrecisionMantissa == 0) && (f32ExtraPrecisionMantissa != 0 || lostPrecision) {
		return true
	}
	return false
}

// Utility function that returns the result for the case when
// the converison results in underflow
func handleUnderflow(signBit uint64, um floatBit.UnderflowMode) (Bits,
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
func handleOverflow(signBit uint64, om floatBit.OverflowMode) (Bits,
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
// up a float32 number into [floatBit.FloatBitFormat]
// Implements the FloatBitFormatter Interface
func (b *Bits) ToFloatFormat() floatBit.FloatBitFormat {
	// Iterate over the bits and construct the return Values

	asUint := uint32(*b)
	signBits := (asUint & SignMask) >> 31
	exponentBits := (asUint & ExponentMask) >> 23
	mantissaBits := asUint & MantissaMask

	// 1 Sign Bit
	signRetVal := make([]byte, 0, 1)
	if signBits == 0 {
		signRetVal = append(signRetVal, byte('0'))
	} else {
		signRetVal = append(signRetVal, byte('1'))
	}

	// 8 Exponent Bits
	exponentRetVal := make([]byte, 0, 8)
	for i := 0; i < 8; i++ {
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

	// 23 Mantissa Bits
	mantissaRetVal := make([]byte, 0, 23)
	for i := 0; i < 23; i++ {
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
