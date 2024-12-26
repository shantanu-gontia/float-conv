package bitVisualization

import (
	"fmt"
	"unsafe"
)

type Float32 struct {
	Val uint32 // Underlying Value stored as float
}

func (f Float32) FromFloat(input float32) Float32 {
	inputPtr := unsafe.Pointer(&input)
	f.Val = *(*uint32)(inputPtr)
	return f
}

// Convert the bit-fields in the Float32 struct to a proper floating-point number
func (input Float32) ToFloat() float32 {
	uintPtr := &input.Val
	uintPtrAsFloatPtr := (*float32)(unsafe.Pointer(uintPtr))
	return *uintPtrAsFloatPtr
}

// Implementation for the FloatBitFormat interface for IEEE-754 Float32 numbers.
// Returns a FloatFormat containing the byte characters for sign, exponent and mantissa bits.
func (input Float32) ToFloatFormat() FloatBitFormat {
	const (
		// Masks
		signBitMask     = 0x8000_0000
		exponentBitMask = 0x7f80_0000
		mantissaBitMask = 0x007f_ffff
	)

	inputAsBits := input.Val

	// Extract Sign Bit
	signBits := (signBitMask & inputAsBits) >> 31

	// Extract Exponent Bits
	exponentBits := (exponentBitMask & inputAsBits) >> 23

	// Extract Mantissa Bits
	mantissaBits := mantissaBitMask & inputAsBits

	// Iterate over the bits and construct the return Values

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
