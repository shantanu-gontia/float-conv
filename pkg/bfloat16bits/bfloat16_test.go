package BF16

import (
	"math"
	"math/big"
	"testing"

	floatBit "github.com/shantanu-gontia/float-conv/pkg"
	F32 "github.com/shantanu-gontia/float-conv/pkg/float32bits"
)

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
			mantissaBits: 0b0_00000000_0000001_0000000000000000,
			goldenVal:    Bits(0b0_00000000_0000001),
			goldenAcc:    big.Exact,
		},
		// Positive RTZ to below
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000001_0100000000000000,
			goldenVal:    Bits(0b0_00000000_0000001),
			goldenAcc:    big.Below,
		},
		// Negative RTZ to above
		{
			signBit:      0x1,
			exponentBits: 0x1,
			mantissaBits: 0b0_00000000_0000001_1111111111111111,
			goldenVal:    Bits(0b1_00000001_0000001),
			goldenAcc:    big.Above,
		},
	}

	for _, tt := range testCases {
		t.Run("RoundTowardsZero", func(t *testing.T) {
			resultVal, resultAcc := truncate(tt.signBit, tt.exponentBits, tt.mantissaBits)
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

	// Rounding towards negative infinity involves adding 1 if the number
	// is negative, otherwise truncating, so that the number is closer to +inf

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
			mantissaBits: 0b0_00000000_0000000_0000000000000000,
			goldenVal:    Bits(0b0_00000010_0000000),
			goldenAcc:    big.Exact,
		},
		// Normal, Rounds up (negative)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b1_00000010_0000001),
			goldenAcc:    big.Below,
		},
		// Normal, Truncates (positive)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b0_00000010_0000000),
			goldenAcc:    big.Below,
		},
		// Subnormal, Rounds up (negative)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1111111_0000000000000011,
			goldenVal:    Bits(0b1_00000001_0000000),
			goldenAcc:    big.Below,
		},
		// Subnormal, truncates (positive)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1111111_1111111111111111,
			goldenVal:    Bits(0b0_00000000_1111111),
			goldenAcc:    big.Below,
		},
	}
	for _, tt := range testCases {
		t.Run("RoundTowardsNegativeInf", func(t *testing.T) {
			resultVal, resultAcc := roundTowardsNegativeInf(tt.signBit, tt.exponentBits, tt.mantissaBits)
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
			mantissaBits: 0b0_00000000_0000000_0000000000000000,
			goldenVal:    Bits(0b0_00000010_0000000),
			goldenAcc:    big.Exact,
		},
		// Normal, Rounds up (positive)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b0_00000010_0000001),
			goldenAcc:    big.Above,
		},
		// Normal, Truncates (negative)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b1_00000010_0000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, Rounds up (positive)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1111111_0000000000000001,
			goldenVal:    Bits(0b0_00000001_0000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, truncates (negative)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1111111_1111111111111111,
			goldenVal:    Bits(0b1_00000000_1111111),
			goldenAcc:    big.Above,
		},
	}

	for _, tt := range testCases {
		t.Run("RoundTowardsPositiveInf", func(t *testing.T) {
			resultVal, resultAcc := roundTowardsPositiveInf(tt.signBit, tt.exponentBits, tt.mantissaBits)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#016x, exponentBits: %#016x, mantissaBits: %#016x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Errorf("Expected Result: %0#8x, Got: %0#8x\n", tt.goldenVal, resultVal)
				t.Errorf("Expected Accuracy: %v, Got: %v\n", tt.goldenAcc, resultAcc)
			}
		})
	}
}

func TestRoundHalfTowardsZero(t *testing.T) {

	// Rounding half towards zero involves rounding to the nearest
	// representable number, and breaking ties by rounding towards the
	// number closer to zero

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
			mantissaBits: 0b0_00000000_0000000_0000000000000000,
			goldenVal:    Bits(0b0_00000010_0000000),
			goldenAcc:    big.Exact,
		},
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0010100_0000000000000000,
			goldenVal:    Bits(0b0_00000010_0010100),
			goldenAcc:    big.Exact,
		},
		// Normal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b1_00000010_0000000),
			goldenAcc:    big.Above,
		},
		// Normal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_0000100000000001,
			goldenVal:    Bits(0b0_00000010_0000000),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b1_00000010_0000001),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b0_00000010_0000001),
			goldenAcc:    big.Above,
		},
		// Normal, truncate, because half-way (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_1000000000000000,
			goldenVal:    Bits(0b0_00000010_0000000),
			goldenAcc:    big.Below,
		},
		// Normal, truncate, because half-way (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_1000000000000000,
			goldenVal:    Bits(0b1_00000010_0000000),
			goldenAcc:    big.Above,
		},
		// Subormal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b1_00000000_0000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b0_00000000_0000000),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b1_00000000_0000001),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b0_00000000_0000001),
			goldenAcc:    big.Above,
		},
		// Subormal, truncate, because half-way (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000_1000000000000000,
			goldenVal:    Bits(0b0_00000000_1000000),
			goldenAcc:    big.Below,
		},
		// Subnormal, truncate, because half-way (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000_1000000000000000,
			goldenVal:    Bits(0b1_00000000_1000000),
			goldenAcc:    big.Above,
		},
	}

	for _, tt := range testCases {
		t.Run("RoundHalfTowardsZero", func(t *testing.T) {
			resultVal, resultAcc := roundHalfTowardsZero(tt.signBit, tt.exponentBits, tt.mantissaBits)
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

	// Rounding half towards negative infinity involves rounding to the nearest
	// representable number, and breaking ties by rounding towards the
	// number closer to -inf

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
			mantissaBits: 0b0_00000000_0000000_0000000000000000,
			goldenVal:    Bits(0b0_00000010_0000000),
			goldenAcc:    big.Exact,
		},
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0010100_0000000000000000,
			goldenVal:    Bits(0b0_00000010_0010100),
			goldenAcc:    big.Exact,
		},
		// Normal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b1_00000010_0000000),
			goldenAcc:    big.Above,
		},
		// Normal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_0000100000000001,
			goldenVal:    Bits(0b0_00000010_0000000),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b1_00000010_0000001),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b0_00000010_0000001),
			goldenAcc:    big.Above,
		},
		// Normal, truncate, because half-way (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_1000000000000000,
			goldenVal:    Bits(0b0_00000010_0000000),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because half-way (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_1000000000000000,
			goldenVal:    Bits(0b1_00000010_0000001),
			goldenAcc:    big.Below,
		},
		// Subormal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b1_00000000_0000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b0_00000000_0000000),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b1_00000000_0000001),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b0_00000000_0000001),
			goldenAcc:    big.Above,
		},
		// Subormal, truncate, because half-way (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000_1000000000000000,
			goldenVal:    Bits(0b0_00000000_1000000),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because half-way (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000_1000000000000000,
			goldenVal:    Bits(0b1_00000000_1000001),
			goldenAcc:    big.Below,
		},
	}

	for _, tt := range testCases {
		t.Run("RoundHalfTowardsNegativeInf", func(t *testing.T) {
			resultVal, resultAcc := roundHalfTowardsNegativeInf(tt.signBit, tt.exponentBits, tt.mantissaBits)
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
			mantissaBits: 0b0_00000000_0000000_0000000000000000,
			goldenVal:    Bits(0b0_00000010_0000000),
			goldenAcc:    big.Exact,
		},
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0010100_0000000000000000,
			goldenVal:    Bits(0b0_00000010_0010100),
			goldenAcc:    big.Exact,
		},
		// Normal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b1_00000010_0000000),
			goldenAcc:    big.Above,
		},
		// Normal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_0000100000000001,
			goldenVal:    Bits(0b0_00000010_0000000),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b1_00000010_0000001),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b0_00000010_0000001),
			goldenAcc:    big.Above,
		},
		// Normal, round up, because half-way (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_1000000000000000,
			goldenVal:    Bits(0b0_00000010_0000001),
			goldenAcc:    big.Above,
		},
		// Normal, truncate, because half-way (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_1000000000000000,
			goldenVal:    Bits(0b1_00000010_0000000),
			goldenAcc:    big.Above,
		},
		// Subormal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b1_00000000_0000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b0_00000000_0000000),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b1_00000000_0000001),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b0_00000000_0000001),
			goldenAcc:    big.Above,
		},
		// Subormal, round up, because half-way (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000_1000000000000000,
			goldenVal:    Bits(0b0_00000000_1000001),
			goldenAcc:    big.Above,
		},
		// Subnormal, truncate, because half-way (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000_1000000000000000,
			goldenVal:    Bits(0b1_00000000_1000000),
			goldenAcc:    big.Above,
		},
	}

	for _, tt := range testCases {
		t.Run("RoundHalfTowardsPositiveInf", func(t *testing.T) {
			resultVal, resultAcc := roundHalfTowardsPositiveInf(tt.signBit, tt.exponentBits, tt.mantissaBits)
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
	// Rounding towards nearest even involves rounding to the nearest
	// representable number, and breaking ties by rounding towards the
	// number with LSB as 0

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
			mantissaBits: 0b0_00000000_0000000_0000000000000000,
			goldenVal:    Bits(0b0_00000010_0000000),
			goldenAcc:    big.Exact,
		},
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000101_0000000000000000,
			goldenVal:    Bits(0b0_00000010_0000101),
			goldenAcc:    big.Exact,
		},
		// Normal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b1_00000010_0000000),
			goldenAcc:    big.Above,
		},
		// Normal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b0_00000010_0000000),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b1_00000010_0000001),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b0_00000010_0000001),
			goldenAcc:    big.Above,
		},
		// Normal, truncate, because half-way, bf16 LSB is zero (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_1110000_1000000000000000,
			goldenVal:    Bits(0b0_00000010_1110000),
			goldenAcc:    big.Below,
		},
		// Normal, truncate, because half-way, bf16 LSB is zero (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_1110000_1000000000000000,
			goldenVal:    Bits(0b1_00000010_1110000),
			goldenAcc:    big.Above,
		},
		// Normal, round up, because half-way, bf16 LSB is one (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_1110001_1000000000000000,
			goldenVal:    Bits(0b0_00000010_1110010),
			goldenAcc:    big.Above,
		},
		// Normal, round up, because half-way, bf16 LSB is one (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_1110001_1000000000000000,
			goldenVal:    Bits(0b1_00000010_1110010),
			goldenAcc:    big.Below,
		},
		// Subormal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_0111000000000000,
			goldenVal:    Bits(0b1_00000000_0000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b0_00000000_00000000),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b1_00000000_0000001),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b0_00000000_0000001),
			goldenAcc:    big.Above,
		},
		// Subnormal, truncate, because half-way, bf16 LSB is zero (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000_1000000000000000,
			goldenVal:    Bits(0b0_00000000_1000000),
			goldenAcc:    big.Below,
		},
		// Subnormal, truncate, because half-way, bf16 LSB is zero (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000_1000000000000000,
			goldenVal:    Bits(0b1_00000000_1000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, round up, because half-way, bf16 LSB is one (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1111111_1000000000000000,
			goldenVal:    Bits(0b0_00000001_0000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, round up, because half-way, bf16 LSB is one (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1111111_1000000000000000,
			goldenVal:    Bits(0b1_00000001_0000000),
			goldenAcc:    big.Below,
		},
		// Minimum Subnormal (Exact)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_0000000_0000001_0000000000000000,
			goldenVal:    Bits(0b0_00000000_0000001),
			goldenAcc:    big.Exact,
		},
	}

	for _, tt := range testCases {
		t.Run("RoundNearestEven", func(t *testing.T) {
			resultVal, resultAcc := roundNearestEven(tt.signBit, tt.exponentBits, tt.mantissaBits)
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
	// Rounding towards nearest odd involves rounding to the nearest
	// representable number, and breaking ties by rounding towards the
	// number resulting in the LSB as 1

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
			mantissaBits: 0b0_00000000_0000000_0000000000000000,
			goldenVal:    Bits(0b0_00000010_0000000),
			goldenAcc:    big.Exact,
		},
		// Exact
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000101_0000000000000000,
			goldenVal:    Bits(0b0_00000010_0000101),
			goldenAcc:    big.Exact,
		},
		// Normal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b1_00000010_0000000),
			goldenAcc:    big.Above,
		},
		// Normal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b0_00000010_0000000),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b1_00000010_0000001),
			goldenAcc:    big.Below,
		},
		// Normal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b0_00000010_0000001),
			goldenAcc:    big.Above,
		},
		// Normal, round up, because half-way, bf16 LSB is zero (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_1110000_1000000000000000,
			goldenVal:    Bits(0b0_00000010_1110001),
			goldenAcc:    big.Above,
		},
		// Normal, truncate, because half-way, bf16 LSB is zero (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_1110000_1000000000000000,
			goldenVal:    Bits(0b1_00000010_1110001),
			goldenAcc:    big.Below,
		},
		// Normal, truncate, because half-way, bf16 LSB is one (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_1110001_1000000000000000,
			goldenVal:    Bits(0b0_00000010_1110001),
			goldenAcc:    big.Below,
		},
		// Normal, truncate, because half-way, bf16 LSB is one (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x2,
			mantissaBits: 0b0_00000000_1110001_1000000000000000,
			goldenVal:    Bits(0b1_00000010_1110001),
			goldenAcc:    big.Above,
		},
		// Subormal, truncate, because closer to truncated value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_0111000000000000,
			goldenVal:    Bits(0b1_00000000_0000000),
			goldenAcc:    big.Above,
		},
		// Subnormal, truncate because closer to truncated value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_0000000000000001,
			goldenVal:    Bits(0b0_00000000_00000000),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b1_00000000_0000001),
			goldenAcc:    big.Below,
		},
		// Subnormal, round up, because closer to rounded up value (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_0000000_1000000000000001,
			goldenVal:    Bits(0b0_00000000_0000001),
			goldenAcc:    big.Above,
		},
		// Subnormal, round up, because half-way, bf16 LSB is zero (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000_1000000000000000,
			goldenVal:    Bits(0b0_00000000_1000001),
			goldenAcc:    big.Above,
		},
		// Subnormal, round up, because half-way, bf16 LSB is zero (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1000000_1000000000000000,
			goldenVal:    Bits(0b1_00000000_1000001),
			goldenAcc:    big.Below,
		},
		// Subnormal, truncate, because half-way, bf16 LSB is one (+ve)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1111111_1000000000000000,
			goldenVal:    Bits(0b0_00000000_1111111),
			goldenAcc:    big.Below,
		},
		// Subnormal, truncate, because half-way, bf16 LSB is one (-ve)
		{
			signBit:      0x1,
			exponentBits: 0x0,
			mantissaBits: 0b0_00000000_1111111_1000000000000000,
			goldenVal:    Bits(0b1_00000000_1111111),
			goldenAcc:    big.Above,
		},
		// Minimum Subnormal (Exact)
		{
			signBit:      0x0,
			exponentBits: 0x0,
			mantissaBits: 0b0_0000000_0000001_0000000000000000,
			goldenVal:    Bits(0b0_00000000_0000001),
			goldenAcc:    big.Exact,
		},
	}

	for _, tt := range testCases {
		t.Run("RoundNearestOdd", func(t *testing.T) {
			resultVal, resultAcc := roundNearestOdd(tt.signBit, tt.exponentBits, tt.mantissaBits)
			if (resultVal != tt.goldenVal) || (resultAcc != tt.goldenAcc) {
				t.Logf("Failed Input Set:\n")
				t.Logf("signBit: %#08x, exponentBits: %#08x, mantissaBits: %#08x", tt.signBit, tt.exponentBits, tt.mantissaBits)
				t.Errorf("Expected Result: %0#4x, Got: %0#4x\n", tt.goldenVal, resultVal)
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
			goldenVal:    Bits(0x3f9d),
			goldenAcc:    big.Below,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "NormalNumberRTPosInf",
			input:        *big.NewFloat(1.2323),
			rm:           floatBit.RoundTowardsPositiveInf,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateInf,
			goldenVal:    Bits(0x3f9e),
			goldenAcc:    big.Above,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "NormalNumberRTNegInf",
			input:        *big.NewFloat(1.2323),
			rm:           floatBit.RoundTowardsNegativeInf,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateInf,
			goldenVal:    Bits(0x3f9d),
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
			input:        *big.NewFloat(float64(math.Float32frombits(uint32(PositiveMinSubnormal) << 16))),
			rm:           floatBit.RoundNearestEven,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(PositiveMinSubnormal),
			goldenAcc:    big.Exact,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "SubnormalPositiveRTZ",
			input:        *big.NewFloat(7.46972e-40),
			rm:           floatBit.RoundTowardsZero,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(0x0008),
			goldenAcc:    big.Below,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "SubnormalNegativeRTZ",
			input:        *big.NewFloat(-7.46972e-40),
			rm:           floatBit.RoundTowardsZero,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(0x8008),
			goldenAcc:    big.Above,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "SubnormalPositiveRoundPosInf",
			input:        *big.NewFloat(7.46972e-40),
			rm:           floatBit.RoundTowardsPositiveInf,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(0x0009),
			goldenAcc:    big.Above,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "SubnormalNegativeRoundPosInf",
			input:        *big.NewFloat(-7.46972e-40),
			rm:           floatBit.RoundTowardsPositiveInf,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateMax,
			goldenVal:    Bits(0x8008),
			goldenAcc:    big.Above,
			goldenStatus: floatBit.Fits,
		},
		{
			name:         "SubnormalNegativeRoundNegInf",
			input:        *big.NewFloat(-7.46972e-40),
			rm:           floatBit.RoundTowardsNegativeInf,
			um:           floatBit.FlushToZero,
			om:           floatBit.SaturateInf,
			goldenVal:    Bits(0x8009),
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
