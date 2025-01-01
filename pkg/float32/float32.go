package Float32

import (
	"fmt"
	"math"
	"math/big"

	floatBit "github.com/shantanu-gontia/float-conv/pkg"
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
func (input Float32) FromBigFloat(bigf big.Float, r floatBit.RoundingMode) (Float32, big.Accuracy) {
	bigf.SetMode(r.ToBigRoundingMode())
	asFloat, acc := bigf.Float32()
	input = input.FromFloat(asFloat)
	return input, acc
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
