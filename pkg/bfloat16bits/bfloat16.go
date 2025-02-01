package BF16

import (
	"errors"
	"math"
	"math/big"

	floatBit "github.com/shantanu-gontia/float-conv/pkg"
	F32 "github.com/shantanu-gontia/float-conv/pkg/float32bits"
)

// Some constants that will help with bit manipulation we'll need to perform
const (
	f32Bfloat16ExtraPrecisionMask uint32 = 0x0000_ffff
)

// Alias type for uint16. This is used to represent the bits that make up a
// BFloat16 number. This type also comes with utility methods to support
// Floating point conversions with different Rounding Modes and Out of Bounds
// responses.
type Bits uint16

// Convert the given [Bits] type to the floating point number it represents,
// inside a [float32] value. This is effectively a bit_cast to bfloat16,
// followed by a up-cast to [float32]
func (input Bits) ToFloat32() float32 {
	asUint32 := uint32(input)
	// Bfloat16 is just the lower 16-bits of float32 truncated
	asUint32 <<= 16
	return math.Float32frombits(asUint32)
}

// Convert the given [Bits] type to a [big.Float] arbitrary precision
// floating-point number
func (input Bits) ToBigFloat() big.Float {
	asFloat32 := input.ToFloat32()
	asBigFloat := *big.NewFloat(float64(asFloat32))
	return asBigFloat
}

// Convert the given [big.Float] arbitrary precision floating-point number
// to a [Bits] type representing the bits of a bfloat16 number. If the number
// cannot be represented in bfloat16 format exactly, then the rounding mode,
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
	// Float64() returns the float32 value that is closest to the input.
	// This might cause it to round up for some cases.
	// But, we need to get the value with extra precision truncated
	// Therefore, to get the truncated result, we need to subtract 1 ULP of
	// precision if the number is positive and the float32 is larger, or
	// if the number is negative and the float32 is smaller, or alternatively
	// if the .Float64() returns big.Above as the accuracy, because for
	// truncation this should always be big.Below
	// Note that however, we need to exempt, the case where the results
	// becomes infinity.
	if math.IsInf(float64(closestFloat32), 1) && fromBigFloatAcc == big.Above {
		// If input was greater than F32 Maximum Normal, then closestFloat32
		// would be +inf, and the accuracy returned would be big.Above
		// F32.PositiveMaxNormal will trigger overflow repsonse in BF16
		asFloat32 = math.Float32frombits(F32.PositiveMaxNormal)
		// We set the accuracy to exact, so that the BF16 methods alter them
		// correctly
		fromBigFloatAcc = big.Exact
	} else if math.IsInf(float64(closestFloat32), -1) && fromBigFloatAcc == big.Below {
		// Similarly,
		// for -inf case, it will be big.Below.
		// F32.NegativeMaxNormal will trigger overflow response in BF16
		asFloat32 = math.Float32frombits(F32.NegativeMaxNormal)
		fromBigFloatAcc = big.Exact
	} else if closestFloat32 == 0.0 && fromBigFloatAcc == big.Below {
		// We also need to do this for the cases, where closestFloat32 is smaller
		// than the minimum float32 subnormal, again, because we want to handle
		// the underflow response in the BF16 methods.
		// F32.PositiveMinSubnormal will trigger underflow response in BF16
		asFloat32 = math.Float32frombits(F32.PositiveMinSubnormal)
		fromBigFloatAcc = big.Exact
	} else if closestFloat32 == -0.0 && fromBigFloatAcc == big.Above {
		// And for the negative case
		// F32.NegativeMinSubnormal will trigger underflow response in BF16
		asFloat32 = math.Float32frombits(F32.NegativeMinSubnormal)
		fromBigFloatAcc = big.Exact
	} else if (input.Sign() > 0 && fromBigFloatAcc == big.Above) ||
		(input.Sign() < 0 && fromBigFloatAcc == big.Below) {
		// For positive numbers if the accuracy was big.Above, then Float32()
		// caused rounding away from zero. This is undesirable. To make it
		// truncation we need to subtract 1 ULP from the number
		closestFloat32Bits := math.Float32bits(closestFloat32)
		asFloat32 = math.Float32frombits(closestFloat32Bits - 1)
		// Since we made it truncation, ther result must now be smaller
		fromBigFloatAcc = big.Below
	} else {
		asFloat32 = closestFloat32
	}

	resultBits, resultAcc, resultStatus := FromFloat32(asFloat32, rm, om, um)

	// If conversion to float32 itself wasn't exact, then we use that as the
	// status, otherwise we use the status from the float32 -> bfloat16
	// conversion
	return resultBits, resultAcc, resultStatus

	// // There is underflow if the resulting float32 was 0, but the input wasn't
	// if input.Sign() != 0 &&
	// 	(math.Float32bits(asFloat32) == 0x8000_0000 ||
	// 		math.Float32bits(asFloat32) == 0x0) {
	// 	resultStatus = floatBit.Underflow
	// }

	// // There is overflow if the resulting float32 was +/- inf but the input
	// // wasn't
	// if !input.IsInf() && math.IsInf(float64(asFloat32), 0) {
	// 	resultStatus = floatBit.Overflow
	// }

	// return resultBits, fromBigFloatAcc, resultStatus
}

// Convert the given [float64] number to a [Bits] type which represents the bits
// of a [float32] number. Signature and usage is identical to [FromBigFloat],
// except the parameter input for this function is [float64]
func FromFloat32(input float32, rm floatBit.RoundingMode,
	om floatBit.OverflowMode, um floatBit.UnderflowMode) (Bits,
	big.Accuracy, floatBit.Status) {

	// There are some special cases we need to handle, particularly, those of
	// values like Inf, NaN which have special encodings in the target formats

	// Since bfloat16 is essentially float32 with 16 less mantissa bits,
	// conversion is simpler
	inputAsBits := math.Float32bits(input)

	// Special Case #1: Infinities
	// Both float32 and bfloat16 have representations for positive and negative
	// infinites.
	if inputAsBits == F32.PositiveInfinity {
		return Bits(PositiveInfinity), big.Exact, floatBit.Fits
	}
	if inputAsBits == F32.NegativeInfinity {
		return Bits(NegativeInfinity), big.Exact, floatBit.Fits
	}

	// Special Case #2: NaNs
	// NaNs always convert to NaNs. For our case, we consider the conversion
	// to be exact
	if math.IsNaN(float64(input)) {
		return Bits(NaN), big.Exact, floatBit.Fits
	}

	// Special Case #3: Zeros
	// Both Negative and Positive Zeros convert to their counterparts exactly
	if inputAsBits == F32.PositiveZero {
		return Bits(PositiveZero), big.Exact, floatBit.Fits
	}
	if inputAsBits == F32.NegativeZero {
		return Bits(NegativeZero), big.Exact, floatBit.Fits
	}

	// With the number interpreted as uint32, we can extract the underlying
	// sign, exponent and mantissa bits.
	signBit := (inputAsBits & F32.SignMask) >> 31
	exponentBits := (inputAsBits & F32.ExponentMask) >> 23
	mantissaBits := (inputAsBits & F32.MantissaMask)

	// Special Case #4: Input is subnormal (exponent bits == 0 && mantissa != 0)
	// If the input is subnormal in the source format, which has more precision
	// bits, it is possible to underflow in the destination format with lower
	// precision bits. For float32 to bfloat16 conversion, this is only valid
	// for the cases where the exponent bits are 0, and the only mantissa bits
	// set are the extra precision bits.
	// The exponent range for bfloat16 and float32 is exactly the same, so,
	// we do not any explicit exponent and mantissa alignment to perform and
	// can handle the rounding and this is the only case for underflow.
	if exponentBits == 0 &&
		(mantissaBits&F32.MantissaMask) <= f32Bfloat16ExtraPrecisionMask {
		return handleUnderflow(signBit, um)
	}

	// Special Case #5: Input exceeds the maximum normal number that is
	// representable (in magnitude) in bfloat16. This constitutes overflow.
	// In this case, the result is determined by the
	// input om [floatBit.OverflowMode]
	if (inputAsBits & 0x7f7f_ffff) > 0x7f7f_0000 {
		return handleOverflow(signBit, om)
	}

	var resultVal Bits
	var resultAcc big.Accuracy

	switch rm {
	case floatBit.RoundTowardsZero:
		resultVal, resultAcc =
			roundTowardsZero(signBit, exponentBits, mantissaBits)
	case floatBit.RoundTowardsNegativeInf:
		resultVal, resultAcc =
			roundDown(signBit, exponentBits, mantissaBits)
	case floatBit.RoundTowardsPositiveInf:
		resultVal, resultAcc =
			roundUp(signBit, exponentBits, mantissaBits)
	case floatBit.RoundHalfTowardsZero:
		resultVal, resultAcc =
			roundHalfTowardsZero(signBit, exponentBits, mantissaBits)
	case floatBit.RoundHalfTowardsNegativeInf:
		resultVal, resultAcc =
			roundHalfTowardsNegativeInf(signBit, exponentBits, mantissaBits)
	case floatBit.RoundHalfTowardsPositiveInf:
		resultVal, resultAcc =
			roundHalfTowardsPositiveInf(signBit, exponentBits, mantissaBits)
	case floatBit.RoundNearestEven:
		resultVal, resultAcc =
			roundNearestEven(signBit, exponentBits, mantissaBits)
	case floatBit.RoundNearestOdd:
		resultVal, resultAcc =
			roundNearestOdd(signBit, exponentBits, mantissaBits)
	}

	return resultVal, resultAcc, floatBit.Fits
}

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
			// Min subnormal is larger than any float32 subnormal that results
			// in underflow
			return Bits(PositiveMinSubnormal), big.Above, floatBit.Underflow
		}
		return Bits(NegativeMinSubnormal), big.Below, floatBit.Underflow
	default:
		panic("Unsupported UnderflowMode encountered")
	}
}

func handleOverflow(signBit uint32, om floatBit.OverflowMode) (Bits,
	big.Accuracy, floatBit.Status) {
	switch om {
	case floatBit.SaturateInf:
		if signBit == 0 {
			// +Inf is greater than any other normal number in float32
			return Bits(PositiveInfinity), big.Above, floatBit.Overflow
		}
		return Bits(NegativeInfinity), big.Below, floatBit.Overflow
	case floatBit.MakeNaN:
		if signBit == 0 {
			// The accuracy doesn't matter for this case
			return Bits(PositiveNaN), big.Above, floatBit.Overflow
		}
		return Bits(NegativeNaN), big.Below, floatBit.Overflow
	case floatBit.SaturateMax:
		if signBit == 0 {
			// The maximum normal in bfloat16 is smaller than any number this
			// function will be invoked for
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

	// 7 Mantissa Bits
	mantissaRetVal := make([]byte, 0, 7)
	for i := 0; i < 7; i++ {
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
