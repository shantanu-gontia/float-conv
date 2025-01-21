package F32

const (
	SignMask     uint32 = 0x8000_0000
	ExponentMask uint32 = 0x7f80_0000
	MantissaMask uint32 = 0x007f_ffff

	PositiveInfinity uint32 = 0x7f80_0000
	NegativeInfinity uint32 = 0xff80_0000

	PositiveZero uint32 = 0x0000_0000
	NegativeZero uint32 = 0x8000_0000

	PositiveMaxNormal uint32 = 0x7f7fffff
	NegativeMaxNormal uint32 = 0xff7fffff

	PositiveMinSubnormal uint32 = 0x00000001
	NegativeMinSubnormal uint32 = 0x80000001

	// In float32 format, all numbers with the exponent bits = 11111111
	// and, mantissa bits not all zero, constitute the special NaN value
	// Though, there technically are two types, we lump them together
	// into just a single NaN. And, whenver the result of an operation is a
	// NaN, we encode it with the same sign as that of the result and,
	// with the mantissa LSB=1, and rest of the mantissa bits=0
	NaN         uint32 = 0x7f800001
	PositiveNaN uint32 = 0x7f800001
	NegativeNaN uint32 = 0xff800001

	ExponentBias int = 127
	ExponentMin  int = -126
	ExponentMax  int = 127
)
