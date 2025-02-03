package F16

import (
	"math"
	"math/big"
	"testing"

	floatBit "github.com/shantanu-gontia/float-conv/pkg"
	F32 "github.com/shantanu-gontia/float-conv/pkg/float32bits"
)

func TestToFloat32(t *testing.T) {
	testCases := []struct {
		// Input
		input Bits
		// Output
		golden float32
	}{
		{
			input:  0b0_01010_0011100000,
			golden: math.Float32frombits(0x3d1c0000),
		},
		{
			input:  0b1_01010_0011100000,
			golden: math.Float32frombits(0xbd1c0000),
		},
		{
			input:  0b0_11111_0000000000,
			golden: float32(math.Inf(1)),
		},
		{
			input:  0b1_11111_0000000000,
			golden: float32(math.Inf(-1)),
		},
		{
			input:  0,
			golden: 0.0,
		},
		{
			input:  0b1_00000_0000000000,
			golden: -0.0,
		},
		{
			input:  0b0_00000_1111111111,
			golden: math.Float32frombits(0x387fc000),
		},
		{
			input:  0b1_00000_1111111111,
			golden: math.Float32frombits(0xb87fc000),
		},
		{
			input:  0b0_00000_0000000010,
			golden: math.Float32frombits(0x34000000),
		},
	}
	for _, tt := range testCases {
		t.Run("ToFloat32", func(t *testing.T) {
			result := tt.input.ToFloat32()
			if result != tt.golden {
				t.Logf("Failed Input Set:\n")
				t.Logf("Input: %0#16b (%0#4x)", tt.input, tt.input)
				t.Errorf("Expected Output: %f (%0#8x). Got: %f (%0#8x)", tt.golden, math.Float32bits(tt.golden), result, math.Float32bits(result))
			}
		})
	}
}

func TestHandleOverflow(t *testing.T) {
	testCases := []struct {
		// In
		signBit uint32
		om      floatBit.OverflowMode
		// Out
		goldenVal    Bits
		goldenAcc    big.Accuracy
		goldenStatus floatBit.Status
	}{
		{0, floatBit.SaturateInf, Bits(PositiveInfinity), big.Above, floatBit.Overflow},
		{1, floatBit.SaturateInf, Bits(NegativeInfinity), big.Below, floatBit.Overflow},
		{200, floatBit.SaturateInf, Bits(NegativeInfinity), big.Below, floatBit.Overflow},
		{0, floatBit.SaturateMax, Bits(PositiveMaxNormal), big.Below, floatBit.Overflow},
		{1, floatBit.SaturateMax, Bits(NegativeMaxNormal), big.Above, floatBit.Overflow},
		{200, floatBit.SaturateMax, Bits(NegativeMaxNormal), big.Above, floatBit.Overflow},
	}

	for _, tt := range testCases {
		t.Run("HandleOverflow", func(t *testing.T) {
			resultVal, resultAcc, resultStatus := handleOverflow(tt.signBit, tt.om)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) || (resultStatus != tt.goldenStatus) {
				t.Logf("Failed Input Set:\n")
				t.Logf("SignBit: %v\tOverflowMode: %v\n", tt.signBit, tt.om)
				t.Errorf("Expected Result: %0#8x, Got: %0#8x\n", tt.goldenVal, resultVal)
				t.Errorf("Expected Accuracy: %v, Got: %v\n", tt.goldenAcc, resultAcc)
				t.Errorf("Expected Status: %v, Got: %v\n", tt.goldenStatus, resultStatus)
			}
		})
	}
}

func TestHandleUnderflow(t *testing.T) {
	testCases := []struct {
		// In
		signBit uint32
		um      floatBit.UnderflowMode
		// Out
		goldenVal    Bits
		goldenAcc    big.Accuracy
		goldenStatus floatBit.Status
	}{
		{0, floatBit.FlushToZero, Bits(PositiveZero), big.Below, floatBit.Underflow},
		{1, floatBit.FlushToZero, Bits(NegativeZero), big.Above, floatBit.Underflow},
		{200, floatBit.FlushToZero, Bits(NegativeZero), big.Above, floatBit.Underflow},
		{0, floatBit.SaturateMin, Bits(PositiveMinSubnormal), big.Above, floatBit.Underflow},
		{1, floatBit.SaturateMin, Bits(NegativeMinSubnormal), big.Below, floatBit.Underflow},
		{200, floatBit.SaturateMin, Bits(NegativeMinSubnormal), big.Below, floatBit.Underflow},
	}

	for _, tt := range testCases {
		t.Run("HandleUnderflow", func(t *testing.T) {
			resultVal, resultAcc, resultStatus := handleUnderflow(tt.signBit, tt.um)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) || (resultStatus != tt.goldenStatus) {
				t.Logf("Failed Input Set:\n")
				t.Logf("SignBit: %v\tUnderflowMode: %v\n", tt.signBit, tt.um)
				t.Errorf("Expected Result: %0#8x, Got: %0#8x\n", tt.goldenVal, resultVal)
				t.Errorf("Expected Accuracy: %v, Got: %v\n", tt.goldenAcc, resultAcc)
				t.Errorf("Expected Status: %v, Got: %v\n", tt.goldenStatus, resultStatus)
			}
		})
	}
}

func TestCheckOverflow(t *testing.T) {
	testCases := []struct {
		// Inputs
		actualExponent int
		mantissaBits   uint32
		// Outputs
		golden bool
	}{
		{
			actualExponent: 12,
			mantissaBits:   0b0_00000000_1111111111_0000000000001,
			golden:         false,
		},
		{
			actualExponent: 15,
			mantissaBits:   0b0_00000000_1111111111_0100000000001,
			golden:         true,
		},
		{
			actualExponent: 15,
			mantissaBits:   0b0_00000000_1111111110_0100000000001,
			golden:         false,
		},
	}

	for _, tt := range testCases {
		result := checkOverflow(tt.actualExponent, tt.mantissaBits)
		if result != tt.golden {
			t.Logf("Failed Input Set:\n")
			t.Logf("Exponent: %d Mantissa Bits: %0#16x", tt.actualExponent, tt.mantissaBits)
			t.Errorf("Expected: %v, Got: %v", tt.golden, result)
		}
	}
}

func TestCheckUnderflow(t *testing.T) {
	testCases := []struct {
		// Inputs
		mantissaBits  uint32
		lostPrecision bool
		// Outputs
		golden bool
	}{
		{
			mantissaBits:  0b0_00000000_0000000000_0100000000001,
			lostPrecision: false,
			golden:        true,
		},
		{
			mantissaBits:  0b0_0000000000_0000000001_0000000000001,
			lostPrecision: false,
			golden:        false,
		},
		{
			mantissaBits:  0b0_00000000_0000000000_0000000000000,
			lostPrecision: false,
			golden:        false,
		},
		{
			mantissaBits:  0b0_00000000_0000000011_0000000000000,
			lostPrecision: false,
			golden:        false,
		},
		{
			mantissaBits:  0b0_00000000_0000000000_0100000000001,
			lostPrecision: true,
			golden:        true,
		},
		{
			mantissaBits:  0b0_00000000_0000000001_0000000000001,
			lostPrecision: true,
			golden:        false,
		},
		{
			mantissaBits:  0b0_00000000_0000000000_0000000000000,
			lostPrecision: true,
			golden:        true,
		},
		{
			mantissaBits:  0b0_00000000_0000000011_0000000000000,
			lostPrecision: true,
			golden:        false,
		},
	}

	for _, tt := range testCases {
		result := checkUnderflow(tt.mantissaBits, tt.lostPrecision)
		if result != tt.golden {
			t.Logf("Failed Input Set:\n")
			t.Logf("Mantissa Bits: %0#16x", tt.mantissaBits)
			t.Logf("Lost Precision?: %v", tt.lostPrecision)
			t.Errorf("Expected: %v, Got: %v", tt.golden, result)
		}
	}
}

func TestRoundTowardsZero(t *testing.T) {
	testCases := []struct {
		// Inputs
		signBit      uint32
		exponentBits uint32
		mantissaBits uint32
		// Outputs
		goldenVal Bits
		goldenAcc big.Accuracy
	}{
		// +Zero -> +Zero
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0x0,
			goldenVal:    Bits(0x0),
			goldenAcc:    big.Exact,
		},
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000001_0000000000000,
			goldenVal:    Bits(0b0_00000_0000000001),
			goldenAcc:    big.Exact,
		},
		// Positive RTZ to below
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000001_0100000000001,
			goldenVal:    Bits(0b0_00000_0000000001),
			goldenAcc:    big.Below,
		},
		// Negative RTZ to above
		{
			signBit:      0x1,
			exponentBits: 0x1,
			mantissaBits: 0b0_000000000_0000101111_1100000000001,
			goldenVal:    Bits(0b1_00001_0000101111),
			goldenAcc:    big.Above,
		},
	}

	for _, tt := range testCases {
		t.Run("RoundTowardsZero", func(t *testing.T) {
			resultVal, resultAcc := truncate(tt.signBit, tt.exponentBits, tt.mantissaBits, false)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#08x, exponentBits: %#08x, mantissaBits: %#08x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Errorf("Expected Result: %0#4x, Got: %0#4x\n", tt.goldenVal, resultVal)
				t.Errorf("Expected Accuracy: %v, Got: %v\n", tt.goldenAcc, resultAcc)
			}
		})
	}
}

func TestRoundTowardsPositiveInf(t *testing.T) {

	// Rounding towards positive infinity involves adding 1 if the number
	// is positive, otherwise truncating, so that the number is closer to +inf

	testCases := []struct {
		// Inputs
		signBit      uint32
		exponentBits uint32
		mantissaBits uint32
		// Outputs
		goldenVal Bits
		goldenAcc big.Accuracy
	}{
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_000000000_0000000000000_0000000000000,
			goldenVal:    Bits(0b0_00010_0000000000),
			goldenAcc:    big.Exact,
		},
		// Normal, Rounds up (positive)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000000_0000000000001,
			goldenVal:    Bits(0b0_00010_0000000001),
			goldenAcc:    big.Above,
		},
		// Normal, Truncates (negative)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b1_00010_0000000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, Rounds up (positive)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1111111111_0000000000001,
			goldenVal:    Bits(0b0_00001_0000000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, truncates (negative)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1111111111_0000000000001,
			goldenVal:    Bits(0b1_00000_1111111111),
			goldenAcc:    big.Above,
		},
	}

	for _, tt := range testCases {
		t.Run("RoundTowardsPositiveInf", func(t *testing.T) {
			resultVal, resultAcc := roundTowardsPositiveInf(tt.signBit, tt.exponentBits, tt.mantissaBits, false)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#08x, exponentBits: %#08x, mantissaBits: %#08x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Errorf("Expected Result: %0#4x, Got: %0#4x\n", tt.goldenVal, resultVal)
				t.Errorf("Expected Accuracy: %v, Got: %v\n", tt.goldenAcc, resultAcc)
			}
		})
	}
}

func TestRoundTowardsNegativeInf(t *testing.T) {

	// Rounding towards positive infinity involves adding 1 if the number
	// is positive, otherwise truncating, so that the number is closer to +inf

	testCases := []struct {
		// Inputs
		signBit      uint32
		exponentBits uint32
		mantissaBits uint32
		// Outputs
		goldenVal Bits
		goldenAcc big.Accuracy
	}{
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_000000000_0000000000000_0000000000000,
			goldenVal:    Bits(0b0_00010_0000000000),
			goldenAcc:    big.Exact,
		},
		// Normal, truncates (positive)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000000_0000000000001,
			goldenVal:    Bits(0b0_00010_0000000000),
			goldenAcc:    big.Below,
		},
		// Normal, rounds up (negative)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b1_00010_0000000001),
			goldenAcc:    big.Below,
		},
		// Subnormal, truncates (positive)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1111111111_0000000000001,
			goldenVal:    Bits(0b0_00000_1111111111),
			goldenAcc:    big.Below,
		},
		// Subnormal, rounds up (negative)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1111111111_0000000000001,
			goldenVal:    Bits(0b1_00001_0000000000),
			goldenAcc:    big.Below,
		},
	}

	for _, tt := range testCases {
		t.Run("RoundTowardsNegativeInf", func(t *testing.T) {
			resultVal, resultAcc := roundTowardsNegativeInf(tt.signBit, tt.exponentBits, tt.mantissaBits, false)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#08x, exponentBits: %#08x, mantissaBits: %#08x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Errorf("Expected Result: %0#4x, Got: %0#4x\n", tt.goldenVal, resultVal)
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
		signBit      uint32
		exponentBits uint32
		mantissaBits uint32
		// Outputs
		goldenVal Bits
		goldenAcc big.Accuracy
	}{
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_0000000000000,
			goldenVal:    Bits(0b0_00010_0000000000),
			goldenAcc:    big.Exact,
		},
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000100100_0000000000000,
			goldenVal:    Bits(0b0_00010_0000100100),
			goldenAcc:    big.Exact,
		},
		// Normal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b1_00010_0000000000),
			goldenAcc:    big.Above,
		},
		// Normal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b0_00010_0000000000),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b1_00010_0000000001),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b0_00010_0000000001),
			goldenAcc:    big.Above,
		},
		// Normal, truncate, because half-way (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_1000000000000,
			goldenVal:    Bits(0b0_00010_0000000000),
			goldenAcc:    big.Below,
		},
		// Normal, truncate, because half-way (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_1000000000000,
			goldenVal:    Bits(0b1_00010_0000000000),
			goldenAcc:    big.Above,
		},
		// Subormal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b1_00000_0000000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b0_00000_0000000000),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b1_00000_0000000001),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b0_00000_0000000001),
			goldenAcc:    big.Above,
		},
		// Subormal, truncate, because half-way (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000000_1000000000000,
			goldenVal:    Bits(0b0_00000_1000000000),
			goldenAcc:    big.Below,
		},
		// Subnormal, truncate, because half-way (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000000_1000000000000,
			goldenVal:    Bits(0b1_00000_1000000000),
			goldenAcc:    big.Above,
		},
	}

	for _, tt := range testCases {
		t.Run("RoundHalfTowardsZero", func(t *testing.T) {
			resultVal, resultAcc := roundHalfTowardsZero(tt.signBit, tt.exponentBits, tt.mantissaBits, false)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#08x, exponentBits: %#08x, mantissaBits: %#08x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Errorf("Expected Result: %0#4x, Got: %0#4x\n", tt.goldenVal, resultVal)
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
		signBit      uint32
		exponentBits uint32
		mantissaBits uint32
		// Outputs
		goldenVal Bits
		goldenAcc big.Accuracy
	}{
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_0000000000000,
			goldenVal:    Bits(0b0_00010_0000000000),
			goldenAcc:    big.Exact,
		},
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000100100_0000000000000,
			goldenVal:    Bits(0b0_00010_0000100100),
			goldenAcc:    big.Exact,
		},
		// Normal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b1_00010_0000000000),
			goldenAcc:    big.Above,
		},
		// Normal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b0_00010_0000000000),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b1_00010_0000000001),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b0_00010_0000000001),
			goldenAcc:    big.Above,
		},
		// Normal, round up, because half-way (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_1000000000000,
			goldenVal:    Bits(0b0_00010_0000000001),
			goldenAcc:    big.Above,
		},
		// Normal, truncate, because half-way (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_1000000000000,
			goldenVal:    Bits(0b1_00010_0000000000),
			goldenAcc:    big.Above,
		},
		// Subormal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b1_00000_0000000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b0_00000_0000000000),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b1_00000_0000000001),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b0_00000_0000000001),
			goldenAcc:    big.Above,
		},
		// Subormal, round up, because half-way (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_1000000000000,
			goldenVal:    Bits(0b0_00000_0000000001),
			goldenAcc:    big.Above,
		},
		// Subnormal, truncate, because half-way (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_1000000000000,
			goldenVal:    Bits(0b1_00000_0000000000),
			goldenAcc:    big.Above,
		},
	}

	for _, tt := range testCases {
		t.Run("RoundHalfTowardsPositiveInf", func(t *testing.T) {
			resultVal, resultAcc := roundHalfTowardsPositiveInf(tt.signBit, tt.exponentBits, tt.mantissaBits, false)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#08x, exponentBits: %#08x, mantissaBits: %#08x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Errorf("Expected Result: %0#4x, Got: %0#4x\n", tt.goldenVal, resultVal)
				t.Errorf("Expected Accuracy: %v, Got: %v\n", tt.goldenAcc, resultAcc)
			}
		})
	}

}

func TestRoundHalfTowardsNegativeInf(t *testing.T) {

	// Rounding half towards positive infinity involves rounding to the nearest
	// representable number, and breaking ties by rounding towards the
	// number closer to +inf

	testCases := []struct {
		// Inputs
		signBit      uint32
		exponentBits uint32
		mantissaBits uint32
		// Outputs
		goldenVal Bits
		goldenAcc big.Accuracy
	}{
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_0000000000000,
			goldenVal:    Bits(0b0_00010_0000000000),
			goldenAcc:    big.Exact,
		},
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000100100_0000000000000,
			goldenVal:    Bits(0b0_00010_0000100100),
			goldenAcc:    big.Exact,
		},
		// Normal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b1_00010_0000000000),
			goldenAcc:    big.Above,
		},
		// Normal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b0_00010_0000000000),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b1_00010_0000000001),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b0_00010_0000000001),
			goldenAcc:    big.Above,
		},
		// Normal, truncate, because half-way (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_1000000000000,
			goldenVal:    Bits(0b0_00010_0000000000),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because half-way (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_1000000000000,
			goldenVal:    Bits(0b1_00010_0000000001),
			goldenAcc:    big.Below,
		},
		// Subormal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_0111000000001,
			goldenVal:    Bits(0b1_00000_0000000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b0_00000_0000000000),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b1_00000_0000000001),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b0_00000_0000000001),
			goldenAcc:    big.Above,
		},
		// Subnormal, truncate, because half-way (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000001_1000000000000,
			goldenVal:    Bits(0b0_00000_1000000001),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because half-way (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000001_1000000000000,
			goldenVal:    Bits(0b1_00000_1000000010),
			goldenAcc:    big.Below,
		},
	}

	for _, tt := range testCases {
		t.Run("RoundHalfTowardsNegativeInf", func(t *testing.T) {
			resultVal, resultAcc := roundHalfTowardsNegativeInf(tt.signBit, tt.exponentBits, tt.mantissaBits, false)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#08x, exponentBits: %#08x, mantissaBits: %#08x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Errorf("Expected Result: %0#4x, Got: %0#4x\n", tt.goldenVal, resultVal)
				t.Errorf("Expected Accuracy: %v, Got: %v\n", tt.goldenAcc, resultAcc)
			}
		})
	}
}

func TestRoundNearestEven(t *testing.T) {
	// Rounding half towards positive infinity involves rounding to the nearest
	// representable number, and breaking ties by rounding towards the
	// number closer to +inf

	testCases := []struct {
		// Inputs
		signBit      uint32
		exponentBits uint32
		mantissaBits uint32
		// Outputs
		goldenVal Bits
		goldenAcc big.Accuracy
	}{
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_00000000000000,
			goldenVal:    Bits(0b0_00010_0000000000),
			goldenAcc:    big.Exact,
		},
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000100100_0000000000000,
			goldenVal:    Bits(0b0_00010_0000100100),
			goldenAcc:    big.Exact,
		},
		// Normal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b1_00010_0000000000),
			goldenAcc:    big.Above,
		},
		// Normal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b0_00010_0000000000),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b1_00010_0000000001),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b0_00010_0000000001),
			goldenAcc:    big.Above,
		},
		// Normal, truncate, because half-way, f32 LSB is zero (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_1110000000_1000000000000,
			goldenVal:    Bits(0b0_00010_1110000000),
			goldenAcc:    big.Below,
		},
		// Normal, truncate, because half-way, f32 LSB is zero (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_1110000000_1000000000000,
			goldenVal:    Bits(0b1_00010_1110000000),
			goldenAcc:    big.Above,
		},
		// Normal, round up, because half-way, f32 LSB is one (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_1110000001_1000000000000,
			goldenVal:    Bits(0b0_00010_1110000010),
			goldenAcc:    big.Above,
		},
		// Normal, round up, because half-way, f32 LSB is one (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_1110000001_1000000000000,
			goldenVal:    Bits(0b1_00010_1110000010),
			goldenAcc:    big.Below,
		},
		// Subormal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_0111000000001,
			goldenVal:    Bits(0b1_00000_0000000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b0_00000_0000000000),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b1_00000_0000000001),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b0_00000_0000000001),
			goldenAcc:    big.Above,
		},
		// Subnormal, truncate, because half-way, f32 LSB is zero (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000000_1000000000000,
			goldenVal:    Bits(0b0_00000_1000000000),
			goldenAcc:    big.Below,
		},
		// Subnormal, truncate, because half-way, f32 LSB is zero (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000000_1000000000000,
			goldenVal:    Bits(0b1_00000_1000000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, round up, because half-way, f32 LSB is one (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1111111111_1000000000000,
			goldenVal:    Bits(0b0_00001_0000000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, round up, because half-way, f32 LSB is one (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1111111111_1000000000000,
			goldenVal:    Bits(0b1_00001_0000000000),
			goldenAcc:    big.Below,
		},
		// Minimum Subnormal (Exact)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000001_0000000000000,
			goldenVal:    Bits(0b0_00000_00000001),
			goldenAcc:    big.Exact,
		},
	}

	for _, tt := range testCases {
		t.Run("RoundHalfTowardsNegativeInf", func(t *testing.T) {
			resultVal, resultAcc := roundNearestEven(tt.signBit, tt.exponentBits, tt.mantissaBits, false)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#08x, exponentBits: %#08x, mantissaBits: %#08x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Errorf("Expected Result: %0#4x, Got: %0#4x\n", tt.goldenVal, resultVal)
				t.Errorf("Expected Accuracy: %v, Got: %v\n", tt.goldenAcc, resultAcc)
			}
		})
	}
}

func TestRoundNearestOdd(t *testing.T) {
	// Rounding half towards positive infinity involves rounding to the nearest
	// representable number, and breaking ties by rounding towards the
	// number closer to +inf

	testCases := []struct {
		// Inputs
		signBit      uint32
		exponentBits uint32
		mantissaBits uint32
		// Outputs
		goldenVal Bits
		goldenAcc big.Accuracy
	}{
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_0000000000000,
			goldenVal:    Bits(0b0_00010_0000000000),
			goldenAcc:    big.Exact,
		},
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000100100_0000000000000,
			goldenVal:    Bits(0b0_00010_0000100100),
			goldenAcc:    big.Exact,
		},
		// Normal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b1_00010_0000000000),
			goldenAcc:    big.Above,
		},
		// Normal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b0_00010_0000000000),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b1_00010_0000000001),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b0_00010_0000000001),
			goldenAcc:    big.Above,
		},
		// Normal, round up, because half-way, f32 LSB is zero (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_1110000000_1000000000000,
			goldenVal:    Bits(0b0_00010_1110000001),
			goldenAcc:    big.Above,
		},
		// Normal, round up, because half-way, f32 LSB is zero (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_1110000000_1000000000000,
			goldenVal:    Bits(0b1_00010_1110000001),
			goldenAcc:    big.Below,
		},
		// Normal, truncate, because half-way, f32 LSB is one (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_1110000001_1000000000000,
			goldenVal:    Bits(0b0_00010_1110000001),
			goldenAcc:    big.Below,
		},
		// Normal, truncate, because half-way, f32 LSB is one (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_1110000001_1000000000000,
			goldenVal:    Bits(0b1_00010_1110000001),
			goldenAcc:    big.Above,
		},
		// Subormal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_0111000000001,
			goldenVal:    Bits(0b1_00000_0000000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_0000000000001,
			goldenVal:    Bits(0b0_00000_0000000000),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b1_00000_0000000001),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000000_1000000000001,
			goldenVal:    Bits(0b0_00000_0000000001),
			goldenAcc:    big.Above,
		},
		// Subnormal, round up, because half-way, f32 LSB is zero (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000000_1000000000000,
			goldenVal:    Bits(0b0_00000_1000000001),
			goldenAcc:    big.Above,
		},
		// Subnormal, round up, because half-way, f32 LSB is zero (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000000_1000000000000,
			goldenVal:    Bits(0b1_00000_1000000001),
			goldenAcc:    big.Below,
		},
		// Subnormal, truncate, because half-way, f32 LSB is one (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1111111111_1000000000000,
			goldenVal:    Bits(0b0_00000_1111111111),
			goldenAcc:    big.Below,
		},
		// Subnormal, truncate, because half-way, f32 LSB is one (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1111111111_1000000000000,
			goldenVal:    Bits(0b1_00000_1111111111),
			goldenAcc:    big.Above,
		},
	}

	for _, tt := range testCases {
		t.Run("RoundHalfTowardsNegativeInf", func(t *testing.T) {
			resultVal, resultAcc := roundNearestOdd(tt.signBit, tt.exponentBits, tt.mantissaBits, false)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#016x, exponentBits: %#016x, mantissaBits: %#016x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Errorf("Expected Result: %0#8x, Got: %0#8x\n", tt.goldenVal, resultVal)
				t.Errorf("Expected Accuracy: %v, Got: %v\n", tt.goldenAcc, resultAcc)
			}
		})
	}
}

func TestFromBigFloat(t *testing.T) {
	testCases := []struct {
		name string
		// Inputs
		input big.Float
		rm    floatBit.RoundingMode
		um    floatBit.UnderflowMode
		om    floatBit.OverflowMode
		// Outputs
		goldenVal    Bits
		goldenAcc    big.Accuracy
		goldenStatus floatBit.Status
	}{
		{
			name:         "PosInfInput",
			input:        *big.NewFloat(math.Inf(1)),
			rm:           floatBit.RoundNearestEven,
			um:           floatBit.SaturateMin,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(PositiveInfinity),
			goldenAcc:    big.Exact,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "NegInfInput",
			input:        *big.NewFloat(math.Inf(-1)),
			rm:           floatBit.RoundHalfTowardsPositiveInf,
			um:           floatBit.SaturateMin,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(NegativeInfinity),
			goldenAcc:    big.Exact,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "PosZeroInput",
			input:        *big.NewFloat(0.0),
			rm:           floatBit.RoundHalfTowardsNegativeInf,
			um:           floatBit.SaturateMin,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(PositiveZero),
			goldenAcc:    big.Exact,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "NegZeroInput",
			input:        *big.NewFloat(float64(math.Float32frombits(F32.NegativeZero))),
			rm:           floatBit.RoundHalfTowardsNegativeInf,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateInf,
			goldenVal:    Bits(NegativeZero),
			goldenAcc:    big.Exact,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "NormalNumberRTZ",
			input:        *big.NewFloat(1.2323),
			rm:           floatBit.RoundTowardsZero,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateInf,
			goldenVal:    Bits(0x3ced),
			goldenAcc:    big.Below,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "NormalNumberRTPosInf",
			input:        *big.NewFloat(1.2323),
			rm:           floatBit.RoundTowardsPositiveInf,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateInf,
			goldenVal:    Bits(0x3cee),
			goldenAcc:    big.Above,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "NormalNumberRTNegInf",
			input:        *big.NewFloat(1.2323),
			rm:           floatBit.RoundTowardsNegativeInf,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateInf,
			goldenVal:    Bits(0x3ced),
			goldenAcc:    big.Below,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "PositiveOverflowToInf",
			input:        *big.NewFloat(3.4028235e38),
			rm:           floatBit.RoundHalfTowardsPositiveInf,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateInf,
			goldenVal:    Bits(PositiveInfinity),
			goldenAcc:    big.Above,
			goldenStatus: floatBit.Overflow,
		},
		{
			name:         "NegativeOverflowToInf",
			input:        *big.NewFloat(-3.4028235e38),
			rm:           floatBit.RoundHalfTowardsPositiveInf,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateInf,
			goldenVal:    Bits(NegativeInfinity),
			goldenAcc:    big.Below,
			goldenStatus: floatBit.Overflow,
		},
		{
			name:         "PositiveOverflowToMax",
			input:        *big.NewFloat(3.4028235e38),
			rm:           floatBit.RoundNearestEven,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(PositiveMaxNormal),
			goldenAcc:    big.Below,
			goldenStatus: floatBit.Overflow,
		},
		{
			name:         "PositiveOverflowToInf",
			input:        *big.NewFloat(-3.4028235e38),
			rm:           floatBit.RoundNearestEven,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(NegativeMaxNormal),
			goldenAcc:    big.Above,
			goldenStatus: floatBit.Overflow,
		},
		{
			name:         "SubnormalPositiveSmallestExact",
			input:        *big.NewFloat(float64(Bits(0b0_00000_0000000001).ToFloat32())),
			rm:           floatBit.RoundNearestEven,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(PositiveMinSubnormal),
			goldenAcc:    big.Exact,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "SubnormalPositiveRTZ",
			input:        *big.NewFloat(float64(math.Float32frombits(0x37980001))),
			rm:           floatBit.RoundTowardsZero,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(0x0130),
			goldenAcc:    big.Below,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "SubnormalNegativeRTZ",
			input:        *big.NewFloat(-0.000018119814),
			rm:           floatBit.RoundTowardsZero,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(0x8130),
			goldenAcc:    big.Above,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "SubnormalPositiveRoundPosInf",
			input:        *big.NewFloat(0.000018119814),
			rm:           floatBit.RoundTowardsPositiveInf,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(0x0131),
			goldenAcc:    big.Above,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "SubnormalNegativeRoundPosInf",
			input:        *big.NewFloat(-0.000018119814),
			rm:           floatBit.RoundTowardsPositiveInf,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(0x8130),
			goldenAcc:    big.Above,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "SubnormalNegativeRoundNegInf",
			input:        *big.NewFloat(-0.000018119814),
			rm:           floatBit.RoundTowardsNegativeInf,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateInf,
			goldenVal:    Bits(0x8131),
			goldenAcc:    big.Below,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "PositiveUnderflowSatMin",
			input:        *big.NewFloat(1e-46),
			rm:           floatBit.RoundHalfTowardsNegativeInf,
			um:           floatBit.SaturateMin,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(PositiveMinSubnormal),
			goldenAcc:    big.Above,
			goldenStatus: floatBit.Underflow,
		},
		{
			name:         "NegativeUnderflowSatMin",
			input:        *big.NewFloat(-1e-46),
			rm:           floatBit.RoundNearestEven,
			um:           floatBit.SaturateMin,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(NegativeMinSubnormal),
			goldenAcc:    big.Below,
			goldenStatus: floatBit.Underflow,
		},
		{
			name:         "PositiveUnderflowFlushZero",
			input:        *big.NewFloat(1e-46),
			rm:           floatBit.RoundHalfTowardsNegativeInf,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(PositiveZero),
			goldenAcc:    big.Below,
			goldenStatus: floatBit.Underflow,
		},
		{
			name:         "NegativeUnderflowFlushZero",
			input:        *big.NewFloat(-1e-46),
			rm:           floatBit.RoundNearestEven,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(NegativeZero),
			goldenAcc:    big.Above,
			goldenStatus: floatBit.Underflow,
		},
	}

	for _, tt := range testCases {
		resultVal, resultAcc, resultStatus := FromBigFloat(tt.input, tt.rm, tt.om, tt.um)
		if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) || (resultStatus != tt.goldenStatus) {
			t.Logf("Failed Input Set:\n")
			t.Logf("Name: %s", tt.name)
			t.Logf("Value: %s", tt.input.String())
			t.Logf("Rounding Mode: %v", tt.rm)
			t.Logf("Overflow Mode: %v", tt.om)
			t.Logf("Underflow Mode: %v", tt.um)
			t.Errorf("Expected result: %.10e (%0#4x), Got: %.10e (%0#4x)", tt.goldenVal.ToFloat32(), tt.goldenVal, resultVal.ToFloat32(), resultVal)
			t.Errorf("Expect accuracy: %v, Got: %v", tt.goldenAcc, resultAcc)
			t.Errorf("Expected status: %v, Got: %v", tt.goldenStatus, resultStatus)
		}
	}
}
