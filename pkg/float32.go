package bitVisualization

import (
	"unsafe"
)

type Float32 struct {
	val uint32 // Underlying Value stored as float
}

// Convert the bit-fields in the Float32 struct to a proper floating-point number
func (input Float32) toFloat() float32 {
	uintPtr := &input.val
	uintPtrAsFloatPtr := (*float32)(unsafe.Pointer(uintPtr))
	return *uintPtrAsFloatPtr
}

// Implementation for the FloatBitFormat interface for IEEE-754 Float32 numbers.
// Returns a FloatFormat containing the byte characters for sign, exponent and mantissa bits.
func (input Float32) toFloatFormat() FloatFormat {
	const (
		// Masks
		signBitMask     = 0x8000_0000
		exponentBitMask = 0x7f80_0000
		mantissaBitMask = 0x007f_ffff
	)

	// Cast to unsafe Pointer to get at the bits
	floatPtrUnsafe := unsafe.Pointer(&input.val)
	// Cast to uint32ptr to get access to bitwise operations
	inputAsBits := *(*uint32)(floatPtrUnsafe)

	// Construct Sign Bits
	signBits := (signBitMask & inputAsBits) >> 31

	// Construct Exponent Bits
	exponentBits := (exponentBitMask & inputAsBits) >> 23
	mantissaBits := mantissaBitMask & inputAsBits

	// Iterate over the bits and construct the return Values
	// 1 Sign Bit
	signRetVal := []byte{byte(signBits)}

	// 8 Exponent Bits
	exponentRetVal := make([]byte, 0, 8)
	for exponentBits != 0 {
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
		mantissaRetVal = append(mantissaRetVal, byte(currentMantissaBit))
		mantissaBits >>= 1
	}

	return FloatFormat{signRetVal, exponentRetVal, mantissaRetVal}
}
