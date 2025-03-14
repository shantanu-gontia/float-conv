package F8E4M3

import (
	"math"
	"math/big"
	"testing"

	floatBit "github.com/shantanu-gontia/float-conv/pkg"
	F32 "github.com/shantanu-gontia/float-conv/pkg/float32bits"
)

func TestToFloat32(t *testing.T) {
    testCases := []struct{
        // Input
        input Bits
        scaleFactor ScaleFactor
        // Output
        golden float32
    }{
        {
            input: 0b0_0000_000,
            scaleFactor: 127,
            golden: math.Float32frombits(F32.PositiveZero),
        },
        {
            input: 0b1_0000_000,
            scaleFactor: 127,
            golden: math.Float32frombits(F32.NegativeZero),
        },
        {
            input: 0b1_0000_000,
            scaleFactor: 254,
            golden: math.Float32frombits(F32.NegativeZero),
        },
        {
            input: 0b0_0000_001,
            scaleFactor: 255,
            golden: math.Float32frombits(F32.PositiveNaN),
        },
        {
            input: 0b1_0000_100,
            scaleFactor: 255,
            golden: math.Float32frombits(F32.NegativeNaN),
        },
        {
            input: 0b0_0000_001,
            scaleFactor: 254,
            // 2^(118)
            golden: math.Float32frombits(0b0_11110101_00000000000000000000000),
        },
        {
            input: 0b1_0000_100,
            scaleFactor: 254,
            // -2^(-7) * 2^(127) = -2^(120)
            golden: math.Float32frombits(0b1_11110111_00000000000000000000000),
        },
        {
            input: 0b0_1001_000,
            scaleFactor: 254,
            // 2^(9-7) => 2 + (254 - 127) = 129 => Infinity
            golden: math.Float32frombits(F32.PositiveInfinity),
        },
        {
            input: 0b1_1001_000,
            scaleFactor: 254, 
            golden: math.Float32frombits(F32.NegativeInfinity),
        },
        {
            input: 0b0_0000_001,
            scaleFactor: 127,
            // 2^(-9) => -9 + 127 = 118
            golden: math.Float32frombits(0b0_01110110_00000000000000000000000),
        },
        {
            input: 0b1_0000_001,
            scaleFactor: 0,
            // 2^(-9) => -9 -127 + 127 = -9 => 0
            golden: math.Float32frombits(F32.NegativeZero),
        },
        {
            input: 0b0_0111_001,
            scaleFactor: 50,
            // 2^(7-7) => 0 + (50 - 127) + 127 = 50
            golden: math.Float32frombits(0b0_00110010_00100000000000000000000),
        },
        {
            input: Bits(PositiveNaN),
            scaleFactor: 127,
            golden: math.Float32frombits(F32.PositiveNaN),
        },
        {
            input: Bits(NegativeNaN),
            scaleFactor: 127,
            golden: math.Float32frombits(F32.NegativeNaN),
        },
        {
            input: Bits(PositiveMaxNormal),
            scaleFactor: 127,
            // Exponent bits -> 2^(15-7) => 8 + (127 - 127) Scale + 127 f32Bias = 135
            golden: math.Float32frombits(0b0_10000111_11000000000000000000000),
        },
        {
            input: Bits(NegativeMaxNormal),
            scaleFactor: 246,
            // Exponent bits -> 2^(15-7) => 8 + (254 - 127) Scale + 127 f32Bias = 254
            golden: math.Float32frombits(0b1_11111110_11000000000000000000000),
        },
    }

    for _, tt := range testCases {
        t.Run("ToFloat32", func(t* testing.T) {
            result := tt.input.ToFloat32(tt.scaleFactor)
            if math.Float32bits(result) != math.Float32bits(tt.golden) {
                t.Logf("Failed Input Set:\n")
                t.Logf("Input: %0#8b (%0#2x)", tt.input, tt.input)
                t.Logf("Scale Factor: %d (%d)", tt.scaleFactor, int(tt.scaleFactor) - 127)
                t.Errorf("Expected Output: %f (%0#8x). Got: %f (%0#8x)", tt.golden, math.Float32bits(tt.golden), result, math.Float32bits(result))
            }
        })
    }
}

func TestCheckOverflow(t* testing.T) {
    testCases := []struct{
        // Inputs
        actualExponent int
        mantissaBits uint32
        scaleFactor ScaleFactor
        // Outputs
        golden bool
    }{
        {
            actualExponent: 8,
            mantissaBits: 0b0_00000000_11100000000000000000000,
            scaleFactor: 127,
            golden: true,
        },
        {
            actualExponent: 8,
            mantissaBits: 0b0_00000000_11000000000000000000000,
            scaleFactor: 127,
            golden: false,
        },
        {
            actualExponent: 127,
            mantissaBits: 0b0_00000000_11100000000000000000000,
            scaleFactor: 254,
            golden: false,
        },
        {
            actualExponent: 55,
            mantissaBits: 0b0_00000000_11000000000000000000000,
            scaleFactor: 55,
            golden: true,
        },
    }

    for _, tt := range testCases {
        result := checkOverflow(tt.actualExponent, tt.mantissaBits, tt.scaleFactor)
        if result != tt.golden {
            t.Logf("Failed Input Set:\n")
            t.Logf("Exponent: %d Mantissa Bits: %#08x, Scale Factor: %d (%d)", tt.actualExponent, tt.mantissaBits, tt.scaleFactor, int(tt.scaleFactor) - 127)
            t.Errorf("Expected: %v, Got: %v", tt.golden, result)
        }
    }
}


func TestCheckUnderflow(t* testing.T) {
    testCases := []struct{
        // Inputs
        mantissaBits uint32
        lostPrecision bool
        // Outputs
        golden bool
    }{
        {
            mantissaBits: 0b0_00000000_11100000000000000000000,
            lostPrecision: false,
            golden: false,
        },
        {
            mantissaBits: 0b0_00000000_00010000000000000000000,
            lostPrecision: false,
            golden: true,
        },
        {
            mantissaBits: 0b0_00000000_00100000000000000000000,
            lostPrecision: true,
            golden: false,
        },
        {
            mantissaBits: 0b0_00000000_00000000000000000000000,
            lostPrecision: true,
            golden: true,
        },
        {
            mantissaBits: 0b0_00000000_00000000000000000000001,
            lostPrecision: false,
            golden: true,
        },
    }

    for _, tt := range testCases {
        result := checkUnderflow(tt.mantissaBits, tt.lostPrecision)
        if result != tt.golden {
            t.Logf("Failed Input Set:\n")
            t.Logf("Mantissa Bits: %#016x, lostPrecison: %v", tt.mantissaBits, tt.lostPrecision)
            t.Errorf("Expected: %v, Got: %v", tt.golden, result)
        }
    }
}

func TestRoundTowardsPositiveInf(t *testing.T) {
	testCases := []struct {
		// Inputs
		signBit       uint32
		exponentBits  uint32
		mantissaBits  uint32
		lostPrecision bool
		// Outputs
		goldenVal Bits
		goldenAcc big.Accuracy
	}{
        // Exact
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_010_00000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_010),
            goldenAcc: big.Exact,
        },
        // Normal, Rounds Up (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_010_00000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_011),
            goldenAcc: big.Above,
        },
        // Normal, Truncates (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_010_00000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_010),
            goldenAcc: big.Above,
        },
        // Subnormal, rounds up (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_010_00000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_011),
            goldenAcc: big.Above,
        },
        // Subornmal, truncates (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_010_00000000000000000001,
            lostPrecision:  false,
            goldenVal: Bits(0b1_0000_010),
            goldenAcc: big.Above,
        },
        // Cases where precision was lost
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_00000000000000000000,
            lostPrecision: true,
            goldenVal: Bits(0b0_0000_010),
            goldenAcc: big.Above,
        },
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_00000000000000000000,
            lostPrecision: true,
            goldenVal: Bits(0b1_0000_001),
            goldenAcc: big.Above,
        },
    }

	for _, tt := range testCases {
		t.Run("RoundTowardsPositiveInf", func(t *testing.T) {
			resultVal, resultAcc := roundTowardsPositiveInf(tt.signBit, tt.exponentBits, tt.mantissaBits, tt.lostPrecision)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#08x, exponentBits: %#08x, mantissaBits: %#08x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Logf("lostPrecision: %v", tt.lostPrecision)
				t.Errorf("Expected Result: %0#2x, Got: %0#2x\n", tt.goldenVal, resultVal)
				t.Errorf("Expected Accuracy: %v, Got: %v\n", tt.goldenAcc, resultAcc)
			}
		})
	}

}

func TestRoundTowardsNegativeInf(t *testing.T) {
	testCases := []struct {
		// Inputs
		signBit       uint32
		exponentBits  uint32
		mantissaBits  uint32
		lostPrecision bool
		// Outputs
		goldenVal Bits
		goldenAcc big.Accuracy
	}{
        // Exact
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_000_00000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_000),
            goldenAcc: big.Exact,
        },
        // Normal, Truncate (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_001_00000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_001),
            goldenAcc: big.Below,
        },
        // Normal, Rounds Up (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_001_00000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_010),
            goldenAcc: big.Below,
        },
        // Subnormal, Truncate (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_00000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_001),
            goldenAcc: big.Below,
        },
        // Subnormal, Rounds Up (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_00000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_010),
            goldenAcc: big.Below,
        },
        // Cases where precision was lost
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_00000000000000000000,
            lostPrecision: true,
            goldenVal: Bits(0b0_0000_001),
            goldenAcc: big.Below,
        },
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_00000000000000000000,
            lostPrecision: true,
            goldenVal: Bits(0b1_0000_010),
            goldenAcc: big.Below,
        },
    }

	for _, tt := range testCases {
		t.Run("RoundTowardsNegativeInf", func(t *testing.T) {
			resultVal, resultAcc := roundTowardsNegativeInf(tt.signBit, tt.exponentBits, tt.mantissaBits, tt.lostPrecision)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#08x, exponentBits: %#08x, mantissaBits: %#08x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Logf("lostPrecision: %v\n", tt.lostPrecision)
				t.Errorf("Expected Result: %0#2x, Got: %0#2x\n", tt.goldenVal, resultVal)
				t.Errorf("Expected Accuracy: %v, Got: %v\n", tt.goldenAcc, resultAcc)
			}
		})
	}
}

func TestRoundTowardsZero(t *testing.T) {
	testCases := []struct {
		// Inputs
		signBit       uint32
		exponentBits  uint32
		mantissaBits  uint32
		lostPrecision bool
		// Outputs
		goldenVal Bits
		goldenAcc big.Accuracy
    }{
        // Zero -> Zero
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0x0,
            lostPrecision: false,
            goldenVal: Bits(0x0),
            goldenAcc: big.Exact,
        },
        // Exact
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_00000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_001),
            goldenAcc: big.Exact,
        },
        // Positive RTZ to below
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_01000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_001),
            goldenAcc: big.Below,
        },
        // Negative RTZ to above
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_011_01000000000000100001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_011),
            goldenAcc: big.Above,
        },
        // Lost precision before got pased into func
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_011_00000000000000000000,
            lostPrecision: true,
            goldenVal: Bits(0b1_0000_011),
            goldenAcc: big.Above,
        },
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_011_00000000000000000000,
            lostPrecision: true,
            goldenVal: Bits(0b0_0000_011),
            goldenAcc: big.Below,
        },
    }

	for _, tt := range testCases {
		t.Run("RoundTowardsZero", func(t *testing.T) {
			resultVal, resultAcc := roundTowardsZero(tt.signBit, tt.exponentBits, tt.mantissaBits, tt.lostPrecision)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#08x, exponentBits: %#08x, mantissaBits: %#08x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Logf("lostPrecision: %v\n", tt.lostPrecision)
				t.Errorf("Expected Result: %0#2x, Got: %0#2x\n", tt.goldenVal, resultVal)
				t.Errorf("Expected Accuracy: %v, Got: %v\n", tt.goldenAcc, resultAcc)
			}
		})
	}
}

func TestRoundHalfTowardsPositiveInf(t *testing.T) {

	// Rounding half towards positive infinity involves rounding to the nearest
	// representable number, and breaking ties by rounding towards the
	// number closer to +inf

	testCases := []struct {
		// Inputs
		signBit       uint32
		exponentBits  uint32
		mantissaBits  uint32
		lostPrecision bool
		// Outputs
		goldenVal Bits
		goldenAcc big.Accuracy
	}{
        // Exact
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_010_00000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_010),
            goldenAcc: big.Exact,
        },
        // Normal, truncate, because closer to truncated value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_000_00000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_000),
            goldenAcc: big.Above,
        },
        // Normal, truncate, because closer to truncated value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_000_00000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_000),
            goldenAcc: big.Below,
        },
        // Normal, round up, because closer to rounded up value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2, 
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_001),
            goldenAcc: big.Below,
        },
        // Normal, round up, because closer to rounded up value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2, 
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_001),
            goldenAcc: big.Above,
        },
        // Normal, round up, because half-way (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_000_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_001),
            goldenAcc: big.Above,
        },
        // Normal, truncate, because half-way (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_000_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_000),
            goldenAcc: big.Above,
        },
        // Subnormal, truncate, because closer to truncated value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_01000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_000),
            goldenAcc: big.Above,
        },
        // Subnormal, truncate, because closer to truncated value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_01000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_000),
            goldenAcc: big.Below,
        },
        // Subnormal, round up, because closer to rounded up value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_001),
            goldenAcc: big.Below,
        },
        // Subnormal, round up, because closr to rounded up value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_001),
            goldenAcc: big.Above,
        },
        // Subnormal, round up, because half-way (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_001),
            goldenAcc: big.Above,
        },
        // Subnormal, truncate, because half-way (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_000),
            goldenAcc: big.Above,
        },
        // Lost Precision
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_10000000000000000000,
            lostPrecision: true,
            goldenVal: Bits(0b1_0000_010),
            goldenAcc: big.Below,
        },
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_10000000000000000000,
            lostPrecision: true,
            goldenVal: Bits(0b0_0000_010),
            goldenAcc: big.Above,
        },
    }

	for _, tt := range testCases {
		t.Run("RoundHalfTowardsPositiveInf", func(t *testing.T) {
			resultVal, resultAcc := roundHalfTowardsPositiveInf(tt.signBit, tt.exponentBits, tt.mantissaBits, tt.lostPrecision)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#08x, exponentBits: %#08x, mantissaBits: %#08x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Logf("lostPrecision: %v", tt.lostPrecision)
				t.Errorf("Expected Result: %0#2x, Got: %0#2x\n", tt.goldenVal, resultVal)
				t.Errorf("Expected Accuracy: %v, Got: %v\n", tt.goldenAcc, resultAcc)
			}
		})
	}
}

func TestRoundHalfTowardsNegativeInf(t *testing.T) {

	// Rounding half towards negative infinity involves rounding to the nearest
	// representable number, and breaking ties by rounding towards the
	// number closer to -inf

	testCases := []struct {
		// Inputs
		signBit       uint32
		exponentBits  uint32
		mantissaBits  uint32
		lostPrecision bool
		// Outputs
		goldenVal Bits
		goldenAcc big.Accuracy
	}{
        // Exact
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_010_00000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_010),
            goldenAcc: big.Exact,
        },
        // Normal, truncate, because closer to truncated value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_000_00000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_000),
            goldenAcc: big.Above,
        },
        // Normal, truncate, because closer to truncated value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_000_00000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_000),
            goldenAcc: big.Below,
        },
        // Normal, round up, because closer to rounded up value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2, 
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_001),
            goldenAcc: big.Below,
        },
        // Normal, round up, because closer to rounded up value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2, 
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_001),
            goldenAcc: big.Above,
        },
        // Normal, round up, because half-way (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_000_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_001),
            goldenAcc: big.Below,
        },
        // Normal, truncate, because half-way (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_000_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_000),
            goldenAcc: big.Below,
        },
        // Subnormal, truncate, because closer to truncated value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_01000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_000),
            goldenAcc: big.Above,
        },
        // Subnormal, truncate, because closer to truncated value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_01000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_000),
            goldenAcc: big.Below,
        },
        // Subnormal, round up, because closer to rounded up value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_001),
            goldenAcc: big.Below,
        },
        // Subnormal, round up, because closer to rounded up value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_001),
            goldenAcc: big.Above,
        },
        // Subnormal, round up, because half-way (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_001),
            goldenAcc: big.Below,
        },
        // Subnormal, truncate, because half-way (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_000),
            goldenAcc: big.Below,
        },
        // Lost Precision
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_10000000000000000000,
            lostPrecision: true,
            goldenVal: Bits(0b1_0000_010),
            goldenAcc: big.Below,
        },
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_10000000000000000000,
            lostPrecision: true,
            goldenVal: Bits(0b0_0000_010),
            goldenAcc: big.Above,
        },
    }

	for _, tt := range testCases {
		t.Run("RoundHalfTowardsNegativeInf", func(t *testing.T) {
			resultVal, resultAcc := roundHalfTowardsNegativeInf(tt.signBit, tt.exponentBits, tt.mantissaBits, tt.lostPrecision)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#08x, exponentBits: %#08x, mantissaBits: %#08x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Logf("lostPrecision: %v", tt.lostPrecision)
				t.Errorf("Expected Result: %0#2x, Got: %0#2x\n", tt.goldenVal, resultVal)
				t.Errorf("Expected Accuracy: %v, Got: %v\n", tt.goldenAcc, resultAcc)
			}
		})
	}
}

func TestRoundHalfTowardsZero(t *testing.T) {

	// Rounding half towards zero involves rounding to the nearest
	// representable number, and breaking ties by rounding towards the
	// number closer to 0

	testCases := []struct {
		// Inputs
		signBit       uint32
		exponentBits  uint32
		mantissaBits  uint32
		lostPrecision bool
		// Outputs
		goldenVal Bits
		goldenAcc big.Accuracy
	}{
        // Exact
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_010_00000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_010),
            goldenAcc: big.Exact,
        },
        // Normal, truncate, because closer to truncated value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_000_00000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_000),
            goldenAcc: big.Above,
        },
        // Normal, truncate, because closer to truncated value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_000_00000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_000),
            goldenAcc: big.Below,
        },
        // Normal, round up, because closer to rounded up value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2, 
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_001),
            goldenAcc: big.Below,
        },
        // Normal, round up, because closer to rounded up value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2, 
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_001),
            goldenAcc: big.Above,
        },
        // Normal, truncate, because half-way (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_000_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_000),
            goldenAcc: big.Above,
        },
        // Normal, truncate, because half-way (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_000_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_000),
            goldenAcc: big.Below,
        },
        // Subnormal, truncate, because closer to truncated value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_01000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_000),
            goldenAcc: big.Above,
        },
        // Subnormal, truncate, because closer to truncated value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_01000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_000),
            goldenAcc: big.Below,
        },
        // Subnormal, round up, because closer to rounded up value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_001),
            goldenAcc: big.Below,
        },
        // Subnormal, round up, because closer to rounded up value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_001),
            goldenAcc: big.Above,
        },
        // Subnormal, truncate, because half-way (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_000),
            goldenAcc: big.Above,
        },
        // Subnormal, truncate, because half-way (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_000),
            goldenAcc: big.Below,
        },
        // Lost Precision
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_10000000000000000000,
            lostPrecision: true,
            goldenVal: Bits(0b1_0000_010),
            goldenAcc: big.Below,
        },
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_10000000000000000000,
            lostPrecision: true,
            goldenVal: Bits(0b0_0000_010),
            goldenAcc: big.Above,
        },
    }

	for _, tt := range testCases {
		t.Run("RoundHalfTowardsZero", func(t *testing.T) {
			resultVal, resultAcc := roundHalfTowardsZero(tt.signBit, tt.exponentBits, tt.mantissaBits, tt.lostPrecision)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#08x, exponentBits: %#08x, mantissaBits: %#08x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Logf("lostPrecision: %v", tt.lostPrecision)
				t.Errorf("Expected Result: %0#2x, Got: %0#2x\n", tt.goldenVal, resultVal)
				t.Errorf("Expected Accuracy: %v, Got: %v\n", tt.goldenAcc, resultAcc)
			}
		})
	}
}


func TestRoundNearestOdd(t *testing.T) {
	// Rounding to nearest even involves rounding to the nearest
	// representable number, and breaking ties by rounding towards the
	// number with LSB = 1

	testCases := []struct {
		// Inputs
		signBit       uint32
		exponentBits  uint32
		mantissaBits  uint32
		lostPrecision bool
		// Outputs
		goldenVal Bits
		goldenAcc big.Accuracy
	}{
        // Exact
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_010_00000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_010),
            goldenAcc: big.Exact,
        },
        // Normal, truncate, because closer to truncated value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_000_00000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_000),
            goldenAcc: big.Above,
        },
        // Normal, truncate, because closer to truncated value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_000_00000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_000),
            goldenAcc: big.Below,
        },
        // Normal, round up, because closer to rounded up value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2, 
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_001),
            goldenAcc: big.Below,
        },
        // Normal, round up, because closer to rounded up value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2, 
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_001),
            goldenAcc: big.Above,
        },
        // Normal, round up, because half-way, f32 LSB is zero (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_110_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_111),
            goldenAcc: big.Above,
        },
        // Normal, round up, because half-way, f32 LSB is zero (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_110_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_111),
            goldenAcc: big.Below,
        },
        // Normal, truncate, because half-way, f32 LSB is 1 (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_101_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_101),
            goldenAcc: big.Below,
        },
        // Normal, truncate, because half-way, f32 LSB is 1 (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_101_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_101),
            goldenAcc: big.Above,
        },
        // Subnormal, truncate, because closer to truncated value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_01000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_000),
            goldenAcc: big.Above,
        },
        // Subnormal, truncate, because closer to truncated value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_01000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_000),
            goldenAcc: big.Below,
        },
        // Subnormal, round up, because closer to rounded up value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_001),
            goldenAcc: big.Below,
        },
        // Subnormal, round up, because closer to rounded up value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_001),
            goldenAcc: big.Above,
        },
        // Subnormal, round up, because half-way, f32 LSB is zero (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_100_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_101),
            goldenAcc: big.Above,
        },
        // Subnormal, round up, because half-way, f32 LSB is zero (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_100_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_101),
            goldenAcc: big.Below,
        },
        // Subnormal, truncate, because half-way, f32 LSB is one (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_101_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_101),
            goldenAcc: big.Below,
        },
        // Subnormal, truncate, because half-way, f32 LSB is one (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_101_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_101),
            goldenAcc: big.Above,
        },
        // Lost Precision
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_10000000000000000000,
            lostPrecision: true,
            goldenVal: Bits(0b1_0000_010),
            goldenAcc: big.Below,
        },
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_10000000000000000000,
            lostPrecision: true,
            goldenVal: Bits(0b0_0000_010),
            goldenAcc: big.Above,
        },
    }

	for _, tt := range testCases {
		t.Run("RoundNearestOdd", func(t *testing.T) {
			resultVal, resultAcc := roundNearestOdd(tt.signBit, tt.exponentBits, tt.mantissaBits, tt.lostPrecision)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#08x, exponentBits: %#08x, mantissaBits: %#08x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Logf("lostPrecision: %v", tt.lostPrecision)
				t.Errorf("Expected Result: %0#2x, Got: %0#2x\n", tt.goldenVal, resultVal)
				t.Errorf("Expected Accuracy: %v, Got: %v\n", tt.goldenAcc, resultAcc)
			}
		})
	}
}

func TestRoundNearestEven(t *testing.T) {
	// Rounding to nearest even involves rounding to the nearest
	// representable number, and breaking ties by rounding towards the
	// number with LSB = 0

	testCases := []struct {
		// Inputs
		signBit       uint32
		exponentBits  uint32
		mantissaBits  uint32
		lostPrecision bool
		// Outputs
		goldenVal Bits
		goldenAcc big.Accuracy
	}{
        // Exact
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_010_00000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_010),
            goldenAcc: big.Exact,
        },
        // Normal, truncate, because closer to truncated value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_000_00000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_000),
            goldenAcc: big.Above,
        },
        // Normal, truncate, because closer to truncated value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_000_00000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_000),
            goldenAcc: big.Below,
        },
        // Normal, round up, because closer to rounded up value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2, 
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_001),
            goldenAcc: big.Below,
        },
        // Normal, round up, because closer to rounded up value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2, 
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_001),
            goldenAcc: big.Above,
        },
        // Normal, truncate, because half-way, f32 LSB is zero (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_110_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_110),
            goldenAcc: big.Below,
        },
        // Normal, truncate, because half-way, f32 LSB is zero (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_110_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_110),
            goldenAcc: big.Above,
        },
        // Normal, round up, because half-way, f32 LSB is 1 (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_101_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0010_110),
            goldenAcc: big.Above,
        },
        // Normal, round up, because half-way, f32 LSB is 1 (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x2,
            mantissaBits: 0b0_00000000_101_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0010_110),
            goldenAcc: big.Below,
        },
        // Subnormal, truncate, because closer to truncated value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_01000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_000),
            goldenAcc: big.Above,
        },
        // Subnormal, truncate, because closer to truncated value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_01000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_000),
            goldenAcc: big.Below,
        },
        // Subnormal, round up, because closer to rounded up value (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_001),
            goldenAcc: big.Below,
        },
        // Subnormal, round up, because closer to rounded up value (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_000_10000000000000000001,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_001),
            goldenAcc: big.Above,
        },
        // Subnormal, truncate, because half-way, f32 LSB is zero (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_100_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_100),
            goldenAcc: big.Below,
        },
        // Subnormal, truncate, because half-way, f32 LSB is zero (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_100_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_100),
            goldenAcc: big.Above,
        },
        // Subnormal, round up, because half-way, f32 LSB is one (+ve)
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_101_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b0_0000_110),
            goldenAcc: big.Above,
        },
        // Subnormal, roudn up, because half-way, f32 LSB is one (-ve)
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_101_10000000000000000000,
            lostPrecision: false,
            goldenVal: Bits(0b1_0000_110),
            goldenAcc: big.Below,
        },
        // Lost Precision
        {
            signBit: 0x1,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_10000000000000000000,
            lostPrecision: true,
            goldenVal: Bits(0b1_0000_010),
            goldenAcc: big.Below,
        },
        {
            signBit: 0x0,
            exponentBits: 0x0,
            mantissaBits: 0b0_00000000_001_10000000000000000000,
            lostPrecision: true,
            goldenVal: Bits(0b0_0000_010),
            goldenAcc: big.Above,
        },
    }

	for _, tt := range testCases {
		t.Run("RoundNearestEven", func(t *testing.T) {
			resultVal, resultAcc := roundNearestEven(tt.signBit, tt.exponentBits, tt.mantissaBits, tt.lostPrecision)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#08x, exponentBits: %#08x, mantissaBits: %#08x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Logf("lostPrecision: %v", tt.lostPrecision)
				t.Errorf("Expected Result: %0#2x, Got: %0#2x\n", tt.goldenVal, resultVal)
				t.Errorf("Expected Accuracy: %v, Got: %v\n", tt.goldenAcc, resultAcc)
			}
		})
	}
}

func TestFromBigFloat(t* testing.T) {
	testCases := []struct {
		name string
		// Inputs
		input big.Float
        scaleFactor ScaleFactor
		rm    floatBit.RoundingMode
		um    floatBit.UnderflowMode
		om    floatBit.OverflowMode
		// Outputs
		goldenVal    Bits
		goldenAcc    big.Accuracy
		goldenStatus floatBit.Status
	}{
    }

    for _, tt := range testCases {
        resultVal, resultAcc, resultStatus := FromBigFloat(tt.input, tt.scaleFactor, tt.rm, tt.om, tt.um)
        if (resultVal != tt.goldenVal || resultAcc != tt.goldenAcc || resultStatus != tt.goldenStatus) {
			t.Logf("Failed Input Set:\n")
			t.Logf("Name: %s", tt.name)
			t.Logf("Value: %s", tt.input.String())
            t.Logf("Scale Factor: %d (%d)", tt.scaleFactor, int(tt.scaleFactor) - 127)
			t.Logf("Rounding Mode: %v", tt.rm)
			t.Logf("Overflow Mode: %v", tt.om)
			t.Logf("Underflow Mode: %v", tt.um)
			t.Errorf("Expected result: %.10e (%0#2x), Got: %.10e (%0#2x)", tt.goldenVal.ToFloat32(tt.scaleFactor), tt.goldenVal, resultVal.ToFloat32(tt.scaleFactor), resultVal)
			t.Errorf("Expect accuracy: %v, Got: %v", tt.goldenAcc, resultAcc)
			t.Errorf("Expected status: %v, Got: %v", tt.goldenStatus, resultStatus)
        }
    }
}
