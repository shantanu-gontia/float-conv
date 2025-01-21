package BF16

const (
	SignMask     uint16 = 0x8000
	ExponentMask uint16 = 0x7f80
	MantissaMask uint16 = 0x007f

	PositiveInfinity uint16 = 0x7f80
	NegativeInfinity uint16 = 0xff80

	PositiveZero uint16 = 0x0000
	NegativeZero uint16 = 0x8000

	PositiveMaxNormal uint16 = 0x7f7f
	NegativeMaxNormal uint16 = 0xff7f

	PositiveMinSubnormal uint16 = 0x0001
	NegativeMinSubnormal uint16 = 0x8001

	// In bfloat16 format, all numbers with the exponent bits = 11111111
	// and, mantissa bits not all zero, constitute the special NaN value
	// Though, there technically are two types, we lump them together
	// into just a single NaN. And, whenver the result of an operation is a
	// NaN, we encode it with the same sign as that of the result and,
	// with the mantissa LSB=1, and rest of the mantissa bits=0
	NaN         uint16 = 0x7f81
	PositiveNaN uint16 = 0x7f81
	NegativeNaN uint16 = 0xff81

	ExponentBias int = 127
	ExponentMin  int = -126
	ExponentMax  int = 127
)
