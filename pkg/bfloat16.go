package bitVisualization

import (
	"fmt"
	"math"
	"unsafe"
)

type BFloat16 struct {
	Val uint16
}

// Convert the input float32 into Float32 format by turning the float32 into its bits
// and truncating the extra precision bits
func (bf BFloat16) FromFloat(input float32) BFloat16 {
	// Truncate the input to 7 mantissa precision (to make the number into BF16)
	// Combine with shifting left by 16 to fit in uint16
	// We need to reinterpret the float32 as bits before that
	inputAsUint32 := math.Float32bits(input)
	inputAsUint32 >>= 16
	bf.Val = uint16(inputAsUint32)
	return bf
}

// Convert the bit-fields in the BFloat16 struct to a proper floating-point number
func (bf BFloat16) ToFloat() float32 {
	// Upcast to a 32-bit container with 0-padding
	sourceToUint32 := uint32(bf.Val)
	// Shift the values to the left by 16 to align with Float32
	sourceAligned := sourceToUint32 << 16
	// Now that the bits are correctly aligned, perform an unsafe cast to float32 ptr and return the dereferenced value
	return *(*float32)(unsafe.Pointer(&sourceAligned))
}

// Implementation for the FloatBitFormat interface for BFloat16 numbers.
// Returns a FloatBitFormat containing the byte characters for sign, exponent and mantissa bits.
func (input BFloat16) ToFloatFormat() FloatBitFormat {
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

	return FloatBitFormat{signRetVal, exponentRetVal, mantissaRetVal}

}

func (input BFloat16) ToHexStr() string {
	return fmt.Sprintf("%#x", input.Val)
}
