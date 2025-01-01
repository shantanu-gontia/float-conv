package BFloat16

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"unsafe"

	floatBit "github.com/shantanu-gontia/float-conv/pkg"
	Float32 "github.com/shantanu-gontia/float-conv/pkg/float32"
	outOfBounds "github.com/shantanu-gontia/float-conv/pkg/oob"
)

// A type wrapper to store the underlying bits of the BFloat16 format
// BFloat16 is a popular floating-point format used in machine learning
type BFloat16 struct {
	Val uint16
}

// Some BFloat16 constants
const (
	BFloat16SignMask     = 0b1_00000000_0000000
	BFloat16ExponentMask = 0b0_11111111_0000000
	BFloat16MantissaMask = 0b0_00000000_1111111

	BFloat16PositiveInfinity = 0b0_11111111_0000000
	BFloat16NegativeInfinity = 0b1_11111111_0000000

	BFloat16PositiveZero = 0b0_00000000_0000000
	BFloat16NegativeZero = 0b1_00000000_0000000

	BFloat16PositiveMaxNormal = 0b0_11111110_1111111
	BFloat16NegativeMaxNormal = 0b1_11111110_1111111

	BFloat16PositiveMinNormal = 0b0_00000001_0000000
	BFloat16NegativeMinNormal = 0b1_00000001_0000000

	BFloat16PositiveMaxSubnormal = 0b0_00000000_1111111
	BFloat16NegativeMaxSubnormal = 0b1_00000000_1111111

	BFloat16PositiveMinSubnormal = 0b0_00000000_0000001
	BFloat16NegativeMinSubnormal = 0b1_00000000_0000001

	// There are multiple values of NaN, we just use this one consistently
	BFloat16PositiveNaN = 0b0_11111111_0000001
	BFloat16NegativeNaN = 0b1_11111111_0000001

	BFloat16ExponentBias = 127
)

func (bf BFloat16) FromBigFloat(input *big.Float, r floatBit.RoundingMode, o outOfBounds.OverflowMode,
	u outOfBounds.UnderflowMode) (BFloat16, big.Accuracy, outOfBounds.Status) {
	input.SetMode(r.ToBigRoundingMode())

	var retAcc big.Accuracy
	var retStatus outOfBounds.Status

	// We use an intermediate step where we convert to Float32 first since big.Float supports it natively
	Float32Val, acc, status := Float32.Float32{}.FromBigFloat(input, r)
	retAcc = acc
	retStatus = status

	BFloat16Val, acc, status := bf.FromFloat(Float32Val.ToFloat(), r, o, u)
	// We update the accuracy only if the Float32 was converted with big.Exact
	if retAcc == big.Exact {
		retAcc = acc
	}
	// We update the overflow status only if the Float32 was converted with outOfBounds.fits
	if retStatus == outOfBounds.Fits {
		retStatus = status
	}

	return BFloat16Val, retAcc, retStatus
}

// Convert the input float32 into BFloat16 type. Since BFloat16 has less precision,
// we use the provided RoundingMode to determine the tie breaking during rounding.
func (bf BFloat16) FromFloat(input float32, r floatBit.RoundingMode, o outOfBounds.OverflowMode,
	u outOfBounds.UnderflowMode) (BFloat16, big.Accuracy, outOfBounds.Status) {
	// Truncate the input to 7 mantissa precision (to make the number into BF16)
	// Combine with shifting left by 16 to fit in uint16
	// We need to reinterpret the float32 as bits before that
	switch r {
	case floatBit.RNE:
		return FromFloatRNE(input, o, u)
	case floatBit.RTZ:
		fallthrough
	default:
		return FromFloatRTZ(input, o, u)
	}
}

// Convert the bits the BFloat16 struct to a proper floating-point number type
func (bf BFloat16) ToFloat() float32 {
	// Upcast to a 32-bit container with 0-padding
	sourceToUint32 := uint32(bf.Val)
	// Shift the values to the left by 16 to align with Float32
	sourceAligned := sourceToUint32 << 16
	// Now that the bits are correctly aligned, perform an unsafe cast to float32 ptr and return the dereferenced value
	return *(*float32)(unsafe.Pointer(&sourceAligned))
}

// Convert the bit-fields in the BFloat16 struct to an arbitrary-precision floating point number
// The precision of the returned big.Float is the default 53
func (bf BFloat16) ToBigFloat() big.Float {
	// Convert to float64 first, which is supported by big.Float
	asFloat := float64(bf.ToFloat())
	return *big.NewFloat(asFloat)
}

// Implementation for the FloatBitFormat interface for BFloat16 numbers.
// Returns a FloatBitFormat containing the byte characters for sign, exponent and mantissa bits.
func (input BFloat16) ToFloatFormat() floatBit.FloatBitFormat {
	const (
		// Masks
		signBitMask     = 0x8000
		exponentBitMask = 0x7f80
		mantissaBitMask = 0x007f
	)

	inputAsBits := input.Val

	// Extract Sign Bit
	signBits := (inputAsBits & signBitMask) >> 15

	// Extract Exponent Bits
	exponentBits := (inputAsBits & exponentBitMask) >> 7

	// Extract Mantissa Bits
	mantissaBits := inputAsBits & mantissaBitMask

	// Iterate over the bits and construct the return value

	// 1 Sign Bit
	signRetVal := make([]byte, 1)
	if signBits == 0 {
		signRetVal[0] = '0'
	} else {
		signRetVal[0] = '1'
	}

	// 8 Exponent Bits
	exponentRetVal := make([]byte, 0, 8)
	for i := 0; i < 8; i++ {
		currentExponentBit := exponentBits & 0x1
		if currentExponentBit == 0 {
			exponentRetVal = append(exponentRetVal, byte('0'))
		} else {
			exponentRetVal = append(exponentRetVal, byte('1'))
		}
		exponentBits >>= 1
	}

	// 7 Mantissa Bits
	mantissaRetVal := make([]byte, 0, 7)
	for i := 0; i < 7; i++ {
		currentMantissaBit := mantissaBits & 0x1
		if currentMantissaBit == 0 {
			mantissaRetVal = append(mantissaRetVal, byte('0'))
		} else {
			mantissaRetVal = append(mantissaRetVal, byte('1'))
		}
		mantissaBits >>= 1
	}

	return floatBit.FloatBitFormat{Sign: signRetVal, Exponent: exponentRetVal, Mantissa: mantissaRetVal}

}

func (input BFloat16) ToHexStr() string {
	return fmt.Sprintf("%#x", input.Val)
}

// Convert the input float32 into BFloat16 type. Break ties while rounding
// by Rounding to the nearest even. Returns the rounded number, accuracy [above (rounded is larger),
// exact, or below (rounded is smaller)], and status which indicates Overflow/Underflow
// Out of Bounds behavior is controlled using om for Overflow and um for Underflow
func FromFloatRNE(input float32, om outOfBounds.OverflowMode, um outOfBounds.UnderflowMode) (BFloat16, big.Accuracy, outOfBounds.Status) {

	// First we convert the number to its bits, so we can manipulate at the
	// bit-level
	floatBits := math.Float32bits(input)
	signBits, _, mantissaBits := Float32.Float32{}.FromFloat(input).ExtractFields()

	// Handle Special Cases

	// 1. +Inf => +Inf
	if floatBits == Float32.Float32PositiveInfinity {
		return BFloat16{BFloat16PositiveInfinity}, big.Exact, outOfBounds.Fits
	}

	// 2. -Inf => -Inf
	if floatBits == Float32.Float32NegativeInfinity {
		return BFloat16{BFloat16NegativeInfinity}, big.Exact, outOfBounds.Fits
	}

	// 3. NaN => NaN (Note this doesn't error out)
	if math.IsNaN(float64(input)) {
		if signBits > 0 {
			return BFloat16{BFloat16NegativeNaN}, big.Exact, outOfBounds.Fits
		}
		return BFloat16{BFloat16PositiveNaN}, big.Exact, outOfBounds.Fits
	}

	// Now, for the rest of the cases, we only need to check the extra precision bits in the mantissa.
	// This is because BFloat16 shares the same exponents as Float32
	// BFloat16 has 7-bits of precision in the mantissa. bit-8 from the mantissa MSB starts to count
	// as the extra precision bit. For RNE, this bit (along with the rest of the mantissa-bits)
	// decides whether we round up or down.

	// Consider the 23-bit float32 mantissa (we've stored it in uint32, so first 9 bits are zeros)
	// x x x x x x x x x m22 m21 m20 m19 m18 m17 m16 m15 m14 m13 m12 m11 m10  m9  m8  m7  m6  m5  m4  m3  m2  m1  m0
	// x x x x x x x x x ___ ___ ___ ___ ___ ___ ___ xxx xxx xxx xxx xxx xxx xxx xxx xxx xxx xxx xxx xxx xxx xxx xxx

	m15 := mantissaBits & 0x00008000
	m14ToM0 := mantissaBits & 0x00007fff
	m16 := mantissaBits & 0x00010000

	// 1. if m15 is zero, that means we're below the half-way point, in which case we round down (just truncate)
	// 2. if m15 is 1, and atleast one of m14-m0 is non-zero, means we're above the half-way point
	// so we round up (add 1)
	// 3. if m15 is 1, and all of m14-m0 are zero. That means we're exactly at the half-way point. In this case,
	// if m16 is zero (even), we round down (truncate). And if m6 is one (odd), we round up (add 1) to make sure the
	// resulting m16 (which will be the LSB for BFloat16) is even

	if m15 == 0 {
		// Case 1
		return roundDown(input, um)
	} else if m15 == 1 && m14ToM0 != 0 {
		// Case 2
		return roundUp(input, om)
	}
	// Remaining - Case 3 m15 == 1 && m14ToM0 == 0
	if m16 == 0 {
		return roundDown(input, um)
	}
	return roundUp(input, om)
}

// Convert the input float32 into BFloat16 type. Break ties while rounding
// by Rounding towards Zero. Returns the rounded number, accuracy [above (rounded is larger),
// exact, or below (rounded is smaller)], and an error if conversion is not possible
// Out of Bounds behavior is controlled using om for Overflow and um for Underflow
func FromFloatRTZ(input float32, om outOfBounds.OverflowMode, um outOfBounds.UnderflowMode) (BFloat16, big.Accuracy, outOfBounds.Status) {

	// First we convert the number to its bits, so we can manipulate at the
	// bit-level
	floatBits := math.Float32bits(input)

	// Handle Special Cases

	// 1. +Inf => +Inf
	if floatBits == Float32.Float32PositiveInfinity {
		return BFloat16{BFloat16PositiveInfinity}, big.Exact, outOfBounds.Fits
	}

	// 2. -Inf => -Inf
	if floatBits == Float32.Float32NegativeInfinity {
		return BFloat16{BFloat16NegativeInfinity}, big.Exact, outOfBounds.Fits
	}

	// 3. NaN => NaN
	if math.IsNaN(float64(input)) {
		if floatBits&Float32.Float32SignMask > 0 {
			return BFloat16{BFloat16NegativeNaN}, big.Exact, outOfBounds.Fits
		}
		return BFloat16{BFloat16PositiveNaN}, big.Exact, outOfBounds.Fits
	}

	// When rounding towards zero, other than handling the special values like Inf, Nan etc.
	// We just need to truncate the input to 7 bits of mantissa and align the result with BFloat16
	return roundDown(input, um)
}

// Utility function to performing rounding down (truncation) a BFloat16 Number
func roundDown(input float32, um outOfBounds.UnderflowMode) (BFloat16, big.Accuracy, outOfBounds.Status) {
	var retVal uint16
	var retAcc big.Accuracy = big.Exact
	var retStatus outOfBounds.Status = outOfBounds.Fits

	floatBits := math.Float32bits(input)
	retVal = uint16(floatBits >> 16)

	// To check if the result was exact, we need to see if any extra mantissa precision bits were set. If any of
	// them were set, the result was not exact, and since we always truncate, this means for negative numbers,
	// the result is larger, and for positive numbers, the result is smaller
	if floatBits&0x0000_ffff != 0 {
		if floatBits&Float32.Float32SignMask != 0 {
			retAcc = big.Above
		} else {
			retAcc = big.Below
		}
	}

	// If the number is a subnormal number, it is possible that the result underflows. This is the case when
	// all the set mantissa bits are shaved off during truncation. This corresponds to the Float32 number being a
	// subnormal that is smaller than the smallest denormal (in magnitude) that can be represented in BFloat16
	// In this case, we have to report underflow and the value of the result will depend on the underflow mode
	if floatBits&0x7fff_0000 == 0 && floatBits&0x0000_ffff != 0 {
		// Underflow Case
		retStatus = outOfBounds.Underflow

		switch um {
		case outOfBounds.SaturateMin:
			if input > 0 {
				retVal = BFloat16PositiveMinSubnormal
			} else {
				retVal = BFloat16PositiveMinSubnormal
			}
		case outOfBounds.FlushToZero:
			if input > 0 {
				retVal = BFloat16PositiveZero
			} else {
				retVal = BFloat16NegativeZero
			}
		}
	}

	return BFloat16{retVal}, retAcc, retStatus
}

// Utility function to performing rounding up (add 1) a BFloat16 Number
func roundUp(input float32, om outOfBounds.OverflowMode) (BFloat16, big.Accuracy, outOfBounds.Status) {
	var retVal uint16
	var retAcc big.Accuracy = big.Exact
	var retStatus outOfBounds.Status = outOfBounds.Fits

	floatBits := math.Float32bits(input)
	truncatedVal := uint16(floatBits >> 16)

	// To check if the result was exact, we need to see if any extra mantissa precision bits were set.
	// Since we're rounding up, the result would be larger for the positive case, and smaller for the negative case
	if floatBits&0x0000_ffff != 0 {
		if floatBits&Float32.Float32SignMask != 0 {
			retAcc = big.Below
		} else {
			retAcc = big.Above
		}
	}

	// To round up, we add 1 to the truncated value
	// There is possibility for overflow. We can use the sign-bit to detect this. As such, we remove the sign bit
	// from the truncated value. We'll add this later
	truncatedValNoSign := truncatedVal & 0x7fff
	truncatedValNoSignPlusOne := truncatedValNoSign + 1
	// It is possible that this addition, might have spilled over to the exponent. The spillover could also have
	// caused the exponent to overflow in which case there will be a 1 in the sign bit.
	// If the exponent did overflow, then we consult the overflow mode for the result
	if truncatedValNoSignPlusOne&0x8000 != 0 {
		// Overflow Case
		switch om {
		case outOfBounds.MakeNaN:
			if input > 0 {
				retVal = BFloat16PositiveNaN
			} else {
				retVal = BFloat16NegativeNaN
			}
		case outOfBounds.SaturateMax:
			if input > 0 {
				retVal = BFloat16PositiveMaxNormal
			} else {
				retVal = BFloat16NegativeMaxNormal
			}
		case outOfBounds.SaturateInf:
			if input > 0 {
				retVal = BFloat16PositiveInfinity
			} else {
				retVal = BFloat16NegativeInfinity
			}
		}
	}

	return BFloat16{retVal}, retAcc, retStatus

}

// Return the difference between this BFloat16 Value and a big.Float value
// where the big.Float value is supposed to be the number this value is
// rounded from. If the input is a NaN then returns an error.
// If either of the inputs is an infinity, returns infinity
func (input BFloat16) ConversionError(bf *big.Float) (big.Float, error) {
	asFloat := input.ToFloat()

	// NaN reports an error. It is the only error case for conversion
	if math.IsNaN(float64(asFloat)) {
		return big.Float{}, errors.New("NaN not supported")
	}

	// If either number is Inf, the difference is +Inf
	if math.IsInf(float64(asFloat), 0) || bf.IsInf() {
		return *big.NewFloat(math.Inf(1)), nil
	}

	// Finally, regular numbers, we just convert the input to big.Float
	// and report the difference
	asBigFloat := input.ToBigFloat()
	asBigFloat.Sub(&asBigFloat, bf)

	return asBigFloat, nil
}
