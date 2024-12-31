package floatBit

import (
	"fmt"
	"math"
)

// A type wrapper to store the underlying bits of the Float32 format
type Float32 struct {
	Val uint32 // Underlying Value stored as float
}

// Extracts the sign, exponent and mantissa bits from a floating point number,
// and returns them in 3 uint32 values
// Note that the values are shifted, to make them occupy the LSB bits
func extractFloat32SignExponentMantissaBits(input float32) (uint32, uint32, uint32) {
	const (
		// Masks
		signBitMask     = 0x8000_0000
		exponentBitMask = 0x7f80_0000
		mantissaBitMask = 0x007f_ffff
	)

	inputAsBits := math.Float32bits(input)

	// Extract Sign Bit
	signBits := (signBitMask & inputAsBits) >> 31

	// Extract Exponent Bits
	exponentBits := (exponentBitMask & inputAsBits) >> 23

	// Extract Mantissa Bits
	mantissaBits := mantissaBitMask & inputAsBits

	return signBits, exponentBits, mantissaBits
}

// Convert the input float32 into Float32 format by turning the float32 into its bits
func (f Float32) FromFloat(input float32) Float32 {
	f.Val = math.Float32bits(input)
	return f
}

// Convert the bit-fields in the Float32 struct to a proper floating-point number
func (input Float32) ToFloat() float32 {
	return math.Float32frombits(input.Val)
}

// Implementation for the FloatBitFormat interface for IEEE-754 Float32 numbers.
// Returns a FloatFormat containing the byte characters for sign, exponent and mantissa bits.
func (input Float32) ToFloatFormat() FloatBitFormat {

	// Iterate over the bits and construct the return Values

	signBits, exponentBits, mantissaBits := extractFloat32SignExponentMantissaBits(input.ToFloat())

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

	return FloatBitFormat{signRetVal, exponentRetVal, mantissaRetVal}
}

// Convert the given Float32 Bit representation to a Hexadecimal String
func (input Float32) ToHexStr() string {
	return fmt.Sprintf("%#x", input.Val)
}
