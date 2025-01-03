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
	Float32SignMask     uint32 = 0x8000_0000
	Float32ExponentMask uint32 = 0x7f80_0000
	Float32MantissaMask uint32 = 0x007f_ffff

	Float32PositiveInfinity uint32 = 0x7f80_0000
	Float32NegativeInfinity uint32 = 0xff80_0000

	Float32NegativeZero uint32 = 0x8000_0000
	Float32PositiveZero uint32 = 0x0000_0000

	Float32PositiveMaxNormal uint32 = 0x7f7fffff
	Float32NegativeMaxNormal uint32 = 0xff7fffff

	Float32PositiveMinSubnormal uint32 = 0x00000001
	Float32NegativeMinSubnormal uint32 = 0x80000001

	Float32ExponentBias int = 127
)

// Convert the input float32 into Float32 format by turning the float32 into its bits
func FromFloat(input float32) Float32 {
	asBits := math.Float32bits(input)
	return Float32{asBits}
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
func FromBigFloat(bigf *big.Float,
	r floatBit.RoundingMode, om outOfBounds.OverflowMode, um outOfBounds.UnderflowMode) (Float32, big.Accuracy, outOfBounds.Status) {
	asFloat, acc := bigf.Float32()

	retVal := FromFloat(asFloat).Val

	/** Check for overflow start **/
	maxNormalPositive := big.NewFloat(float64(math.Float32frombits(Float32PositiveMaxNormal)))
	maxNormalNegative := big.NewFloat(float64(math.Float32frombits(Float32NegativeMaxNormal)))

	// Create a copy of the original big.Float
	bigfCopy := big.Float{}
	bigfCopy.Copy(bigf)

	// Positive case, check if num > maximum normal (in magnitude)
	if !bigf.IsInf() && bigf.Sign() > 0 {
		bigfCopy.Sub(&bigfCopy, maxNormalPositive)
		if bigfCopy.Sign() > 0 { // overflow
			switch om {
			case outOfBounds.SaturateInf:
				return Float32{Float32PositiveInfinity}, big.Above, outOfBounds.Overflow
			case outOfBounds.SaturateMax:
				return Float32{Float32PositiveMaxNormal}, big.Below, outOfBounds.Overflow
			case outOfBounds.MakeNaN:
				return Float32{Float32PositiveInfinity + 1}, big.Exact, outOfBounds.Overflow
			}
		}
	}

	// Negative case, check if num < maximum normal (in magnitude)
	if !bigf.IsInf() && bigf.Sign() < 0 {
		bigfCopy.Sub(&bigfCopy, maxNormalNegative)
		if bigfCopy.Sign() < 0 {
			switch om {
			case outOfBounds.SaturateInf:
				return Float32{Float32NegativeInfinity}, big.Below, outOfBounds.Overflow
			case outOfBounds.SaturateMax:
				return Float32{Float32NegativeMaxNormal}, big.Above, outOfBounds.Overflow
			case outOfBounds.MakeNaN:
				return Float32{Float32NegativeInfinity + 1}, big.Exact, outOfBounds.Overflow
			}
		}
	}
	/** Check for overflow end  **/

	/** Check for underflow start **/
	// bigf.Sign() == 0 is only true if bigf is zero. So we use it as a equality check here
	// If big float was non-zero, but rounded value was zero => underflow
	if (bigf.Sign() != 0) && (retVal == Float32PositiveZero || retVal == Float32NegativeZero) {
		switch um {
		case outOfBounds.FlushToZero:
			if bigf.Sign() > 0 {
				retVal = Float32PositiveZero
				acc = big.Below
			} else {
				retVal = Float32NegativeZero
				acc = big.Above
			}
		case outOfBounds.SaturateMin:
			if bigf.Sign() > 0 {
				retVal = Float32PositiveMinSubnormal
				acc = big.Above
			} else {
				retVal = Float32NegativeMinSubnormal
				acc = big.Below
			}
		}
		return Float32{retVal}, acc, outOfBounds.Underflow
	}
	/** Check for underflow end **/

	return Float32{retVal}, acc, outOfBounds.Fits
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
	exponentRetVal := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		currentExponentBit := exponentBits & 0x1
		if currentExponentBit == 0 {
			exponentRetVal[i] = '0'
		} else {
			exponentRetVal[i] = '1'
		}
		exponentBits >>= 1
	}

	// 23 Mantissa Bits
	mantissaRetVal := make([]byte, 23)
	for i := 22; i >= 0; i-- {
		currentMantissaBit := mantissaBits & 0x1
		if currentMantissaBit == 0 {
			mantissaRetVal[i] = '0'
		} else {
			mantissaRetVal[i] = '1'
		}
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
	// and report the difference. big.Float.Sub stores the result in the
	// receiver
	asBigFloat := input.ToBigFloat()
	asBigFloat.Sub(&asBigFloat, bf)

	return asBigFloat, nil
}
