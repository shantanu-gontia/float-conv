package F8E8M0

import (
	"math"

	F32 "github.com/shantanu-gontia/float-conv/pkg/float32bits"
)

// Alias type for uint8. This is used to represent the FP8E8M0 format which
// is used to specify the scale factor for a OCP MXFP8 floating point format.
// The actual floating point represented by an OCP MXPF8 number is the
// value encoded in the bits scaled by 2^(scale_factor) when scale_factor
// is interpreted as an fp8e8m0 number. For example, to apply no scaling,
// the number passed must be 127, because the actual scale that is
// multiplied is 2^(scale_factor - 127). If the exponent in the resulting
// number is out of the range supported by the format, then the result is
// undefined per the spec. In our case, we clamp to Infinity of the same sign
// as the original number
type ScaleFactor uint8

// Apply scale factor to the given float32 number
func ApplyScaleFactor(input float32, scaleFactor ScaleFactor) float32 {
    inputAsBits := math.Float32bits(input)
    inputAsBitsNoExponent := (inputAsBits & F32.SignMask)  | 
        (inputAsBits & F32.MantissaMask)
    inputAsBitsMaskedExponent := inputAsBits & F32.ExponentMask
    exponentBits := (inputAsBitsMaskedExponent) >> 20
    actualExponent := int(exponentBits) - 127
    scaleFactorAsInt := int(scaleFactor) - 127
    resultExponent := actualExponent - scaleFactorAsInt
    resultExponentBits := (uint32(resultExponent) + 127) << 20
    resultBits := inputAsBitsNoExponent | resultExponentBits
    return math.Float32frombits(resultBits)
}

const (
    NaN ScaleFactor = 255
    ExponentBias int = 127
)

