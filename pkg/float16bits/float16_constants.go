package F16

const (
	SignMask     uint16 = 0b1_00000_0000000000
	ExponentMask uint16 = 0b0_11111_0000000000
	MantissaMask uint16 = 0b0_00000_1111111111

	PositiveInfinity uint16 = 0b0_11111_0000000000
	NegativeInfinity uint16 = 0b1_11111_0000000000

	PositiveMaxNormal uint16 = 0b0_11110_1111111111
	NegativeMaxNormal uint16 = 0b0_11110_1111111111

	PositiveZero uint16 = 0b0_00000_0000000000
	NegativeZero uint16 = 0b1_00000_0000000000

	PositiveMinSubnormal uint16 = 0b0_00000_0000000001
	NegativeMinSubnormal uint16 = 0b1_00000_0000000001

	// In float16 format, all numbers with the exponent bits = 11111
	// and mantissa bits not all zero, constitute the special NaN value
	// We just use the first of these values as the flag NaN value, whenever
	// we want to return one. When parsing, all these values will be treated
	// as NaN
	NaN         uint16 = 0b0_11111_0000000001
	PositiveNaN uint16 = 0b0_11111_0000000001
	NegativeNaN uint16 = 0b1_11111_0000000001

	ExponentBias int = 15
	ExponentMin  int = -14
	ExponentMax  int = 15
)
