package F8E4M3

const (
    SignMask uint8 = 0b1_0000_000
    ExponentMask uint8 = 0b0_1111_000
    MantissaMask uint8 = 0b0_0000_111

    // Float8E4M3 doesn't support Infinities

    PositiveMaxNormal uint8 = 0b0_1111_110
    NegativeMaxNormal uint8 = 0b1_1111_110

    PositiveZero uint8 = 0b0_0000_000
    NegativeZero uint8 = 0b1_0000_000

    PositiveMinSubnormal uint8 = 0b0_0000_001
    NegativeMinSubnormal uint8 = 0b1_0000_001

    NaN uint8 = 0b0_1111_111
    PositiveNaN uint8 = 0b0_1111_111
    NegativeNaN uint8 = 0b1_1111_111

    ExponentBias int = 7
    ExponentMin int = -6
    ExponentMax int = 8
)
