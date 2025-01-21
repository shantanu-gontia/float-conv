package BF16

import (
	"math/big"
	"testing"

	floatBit "github.com/shantanu-gontia/float-conv/pkg"
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
		signBit      uint64
		exponentBits uint64
		mantissaBits uint64
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
		signBit      uint64
		exponentBits uint64
		mantissaBits uint64
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
