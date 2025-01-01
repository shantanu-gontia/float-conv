package Float32

import (
	"errors"
	"fmt"
	"math"
	"math/big"

	floatBit "github.com/shantanu-gontia/float-conv/pkg"
	outOfBounds "github.com/shantanu-gontia/float-conv/pkg/oob"
)

// A type wrapper to store the underlying bits of the Float32 format
type Float32 struct {
	Val uint32 // Underlying Value stored as float
}

// Some Float32 Constants
const (
	Float32SignMask     = 0x8000_0000
	Float32ExponentMask = 0x7f80_0000
	Float32MantissaMask = 0x007f_ffff

	Float32PositiveInfinity = 0x7f80_0000
	Float32NegativeInfinity = 0xff80_0000

	Float32NegativeZero = 0x8000_0000
	Float32PositiveZero = 0x0000_0000

	Float32PositiveMaxNormal = 0x7f7fffff
	Float32NegativeMaxNormal = 0xff7fffff

	Float32PositiveMinSubnormal = 0x00000001
	Float32NegativeMinSubnormal = 0x80000001

	Float32ExponentBias = 127
)

// Convert the input float32 into Float32 format by turning the float32 into its bits
func (f Float32) FromFloat(input float32) Float32 {
	f.Val = math.Float32bits(input)
	return f
}

// Convert the bit-fields in the Float32 struct to a proper floating-point number
func (input Float32) ToFloat() float32 {
	return math.Float32frombits(input.Val)
}

// Convert the bit-fields in the Float32 struct to an arbitrary-precision floating point number
// The precision of the returned big.Float is the default 53
func (input Float32) ToBigFloat() big.Float {
	asFloat := float64(input.ToFloat())
	return *big.NewFloat(asFloat)
}

// Convert the big.Float input to a float32 format and subsequently, to the Float32 Type.
// Use roundingMode to determine how to break ties when an exact match is not possible
// Returns the converted bits, and a big.Accuracy which represents the difference from the exact match.
func (input Float32) FromBigFloat(bigf *big.Float,
	r floatBit.RoundingMode, om outOfBounds.OverflowMode, um outOfBounds.UnderflowMode) (Float32, big.Accuracy, outOfBounds.Status) {
	bigf.SetMode(r.ToBigRoundingMode())

	asFloat, acc := bigf.Float32()
	input = input.FromFloat(asFloat)

	// Check for overflow
	if (input.Val == Float32PositiveInfinity || input.Val == Float32NegativeInfinity) && !bigf.IsInf() {
		switch om {
		case outOfBounds.SaturateInf:
			if asFloat > 0 {
				input.Val = Float32PositiveInfinity
				acc = big.Above
			} else {
				input.Val = Float32NegativeInfinity
				acc = big.Below
			}
		case outOfBounds.MakeNaN:
			if asFloat > 0 {
				input.Val = Float32PositiveInfinity + 1
			} else {
				input.Val = Float32NegativeInfinity + 1
			}
		case outOfBounds.SaturateMax:
			if asFloat > 0 {
				input.Val = Float32PositiveMaxNormal
				acc = big.Below
			} else {
				input.Val = Float32NegativeMaxNormal
				acc = big.Above
			}
		}
		return input, acc, outOfBounds.Overflow
	}

	// Check for underflow
	// bigf.Sign() == 0 is only true if bigf is zero. So we use it as a equality check here
	// If big float was non-zero, but rounded value was zero => underflow
	if (bigf.Sign() != 0) && (input.Val == Float32PositiveZero || input.Val == Float32NegativeZero) {
		switch um {
		case outOfBounds.FlushToZero:
			if asFloat > 0 {
				input.Val = Float32PositiveZero
				acc = big.Below
			} else {
				input.Val = Float32NegativeZero
				acc = big.Above
			}
		case outOfBounds.SaturateMin:
			if asFloat > 0 {
				input.Val = Float32PositiveMinSubnormal
				acc = big.Above
			} else {
				input.Val = Float32NegativeMinSubnormal
				acc = big.Below
			}
		}
		return input, acc, outOfBounds.Underflow
	}

	return input, acc, outOfBounds.Fits
}

// Implementation for the FloatBitFormat interface for IEEE-754 Float32 numbers.
// Returns a FloatFormat containing the byte characters for sign, exponent and mantissa bits.
func (input Float32) ToFloatFormat() floatBit.FloatBitFormat {

	// Iterate over the bits and construct the return Values

	signBits, exponentBits, mantissaBits := input.ExtractFields()

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

	return floatBit.FloatBitFormat{Sign: signRetVal, Exponent: exponentRetVal, Mantissa: mantissaRetVal}
}

// Extracts the sign, exponent and mantissa bits from a floating point number,
// and returns them in 3 uint32 values
// Note that the values are shifted, to make them occupy the LSB bits
func (input Float32) ExtractFields() (uint32, uint32, uint32) {
	asBits := input.Val

	// Extract Sign Bit
	signBits := (Float32SignMask & asBits) >> 31

	// Extract Exponent Bits
	exponentBits := (Float32ExponentMask & asBits) >> 23

	// Extract Mantissa Bits
	mantissaBits := Float32MantissaMask & asBits

	return signBits, exponentBits, mantissaBits
}

// Convert the given Float32 Bit representation to a Hexadecimal String
func (input Float32) ToHexStr() string {
	return fmt.Sprintf("%#x", input.Val)
}

// Return the difference between this float32 Value and a big.Float value
// where the big.Float value is supposed to be the number this value is
// rounded from. If the input is a NaN then returns an error.
// If either of the inputs is an infinity, returns infinity
func (input Float32) ConversionError(bf *big.Float) (big.Float, error) {
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
