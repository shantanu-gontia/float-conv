package Float32

import (
	"math"
	"math/big"
	"testing"

	floatBit "github.com/shantanu-gontia/float-conv/pkg"
	outOfBounds "github.com/shantanu-gontia/float-conv/pkg/oob"
)

type InputVals struct {
	input big.Float
	r     floatBit.RoundingMode
	o     outOfBounds.OverflowMode
	u     outOfBounds.UnderflowMode
}

type OutputVals struct {
	output Float32
	acc    big.Accuracy
	status outOfBounds.Status
}

type InputOutputPair struct {
	ival InputVals
	oval OutputVals
}

type TestCase struct {
	name   string
	ioVals []InputOutputPair
}

func TestFromBigFloat(t *testing.T) {
	testCases := []TestCase{
		{
			"PosInfInput",
			[]InputOutputPair{
				{
					InputVals{*big.NewFloat(math.Inf(1)),
						floatBit.RNE, outOfBounds.SaturateMax, outOfBounds.SaturateMin},
					OutputVals{Float32{Float32PositiveInfinity}, big.Exact, outOfBounds.Fits},
				},
				{
					InputVals{*big.NewFloat(math.Inf(1)), floatBit.RTZ, outOfBounds.MakeNaN, outOfBounds.SaturateMin},
					OutputVals{Float32{Float32PositiveInfinity}, big.Exact, outOfBounds.Fits},
				},
			},
		},
		{
			"NegInfInput",
			[]InputOutputPair{
				{
					InputVals{*big.NewFloat(math.Inf(-1)),
						floatBit.RNE, outOfBounds.SaturateMax, outOfBounds.SaturateMin},
					OutputVals{Float32{Float32NegativeInfinity}, big.Exact, outOfBounds.Fits},
				},
				{
					InputVals{*big.NewFloat(math.Inf(-1)), floatBit.RTZ, outOfBounds.MakeNaN, outOfBounds.SaturateMin},
					OutputVals{Float32{Float32NegativeInfinity}, big.Exact, outOfBounds.Fits},
				},
			},
		},
		{
			"PosZeroInput",
			[]InputOutputPair{
				{
					InputVals{*big.NewFloat(0.0), floatBit.RNE, outOfBounds.SaturateMax, outOfBounds.SaturateMin},
					OutputVals{Float32{Float32PositiveZero}, big.Exact, outOfBounds.Fits},
				},
			},
		},
		{
			"NegZeroInput",
			[]InputOutputPair{
				{
					InputVals{*big.NewFloat(float64(math.Float32frombits(Float32NegativeZero))),
						floatBit.RNE, outOfBounds.SaturateMax, outOfBounds.SaturateMin},
					OutputVals{Float32{Float32NegativeZero}, big.Exact, outOfBounds.Fits},
				},
			},
		},
		{
			"NormalNumbers",
			[]InputOutputPair{
				{
					InputVals{*big.NewFloat(1.2323), floatBit.RNE, outOfBounds.SaturateMax, outOfBounds.SaturateMin},
					OutputVals{Float32{0x3f9dbc02}, big.Above, outOfBounds.Fits},
				},
				{
					InputVals{*big.NewFloat(1.2323), floatBit.RTZ, outOfBounds.SaturateMax, outOfBounds.SaturateMin},
					OutputVals{Float32{0x3f9dbc02}, big.Above, outOfBounds.Fits},
				},
				{
					InputVals{*big.NewFloat(-0.1), floatBit.RNE, outOfBounds.SaturateMax, outOfBounds.SaturateMin},
					OutputVals{Float32{0xbdcccccd}, big.Below, outOfBounds.Fits},
				},
				{
					InputVals{*big.NewFloat(-0.1), floatBit.RTZ, outOfBounds.SaturateMax, outOfBounds.SaturateMin},
					OutputVals{Float32{0xbdcccccd}, big.Below, outOfBounds.Fits},
				},
				// Overflow
				{
					InputVals{*big.NewFloat(3.4028235e38), floatBit.RNE, outOfBounds.SaturateMax, outOfBounds.SaturateMin},
					OutputVals{Float32{Float32PositiveMaxNormal}, big.Below, outOfBounds.Overflow},
				},
				{
					InputVals{*big.NewFloat(3.4028235e38), floatBit.RNE, outOfBounds.SaturateInf, outOfBounds.SaturateMin},
					OutputVals{Float32{Float32PositiveInfinity}, big.Above, outOfBounds.Overflow},
				},
				{
					InputVals{*big.NewFloat(-3.4028235e38), floatBit.RNE, outOfBounds.SaturateMax, outOfBounds.SaturateMin},
					OutputVals{Float32{Float32NegativeMaxNormal}, big.Above, outOfBounds.Overflow},
				},
				{
					InputVals{*big.NewFloat(-3.4028235e38), floatBit.RNE, outOfBounds.SaturateInf, outOfBounds.SaturateMin},
					OutputVals{Float32{Float32NegativeInfinity}, big.Below, outOfBounds.Overflow},
				},
			},
		},
		{
			"SubnormalNumbers",
			[]InputOutputPair{
				// Smallest Float32 Subnormal
				{
					InputVals{*big.NewFloat(float64(math.Float32frombits(0x0000_0001))), floatBit.RNE, outOfBounds.SaturateMax, outOfBounds.SaturateMin},
					OutputVals{Float32{0x0000_0001}, big.Exact, outOfBounds.Fits},
				},
				// More precision than Float32 has
				{
					InputVals{*big.NewFloat(1.2288e-41), floatBit.RNE, outOfBounds.SaturateMax, outOfBounds.SaturateMin},
					OutputVals{Float32{0x0000_2241}, big.Below, outOfBounds.Fits},
				},
				{
					InputVals{*big.NewFloat(-1.2288e-41), floatBit.RNE, outOfBounds.SaturateMax, outOfBounds.SaturateMin},
					OutputVals{Float32{0x8000_2241}, big.Above, outOfBounds.Fits},
				},
				// Underflow
				{
					InputVals{*big.NewFloat(1e-46), floatBit.RNE, outOfBounds.SaturateMax, outOfBounds.SaturateMin},
					OutputVals{Float32{Float32PositiveMinSubnormal}, big.Above, outOfBounds.Underflow},
				},
				{
					InputVals{*big.NewFloat(1e-46), floatBit.RNE, outOfBounds.SaturateMax, outOfBounds.FlushToZero},
					OutputVals{Float32{Float32PositiveZero}, big.Below, outOfBounds.Underflow},
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			for _, iopair := range tt.ioVals {
				resultVal, resultAcc, resultStatus := FromBigFloat(&iopair.ival.input, iopair.ival.r, iopair.ival.o, iopair.ival.u)
				wantVal := iopair.oval.output
				wantAcc := iopair.oval.acc
				wantStatus := iopair.oval.status

				if (resultVal != wantVal) || (resultAcc != wantAcc) || (resultStatus != wantStatus) {
					t.Logf("Failed Input Set:\n")
					t.Logf("Value: %s", iopair.ival.input.String())
					t.Logf("Rounding Mode: %v", iopair.ival.r)
					t.Logf("Overflow Mode: %v", iopair.ival.o)
					t.Logf("Underflow Mode: %v", iopair.ival.u)
					t.Errorf("Expected result: %.10e (%0#8x), Got: %.10e (%0#8x)", wantVal.ToFloat(), wantVal, resultVal.ToFloat(), resultVal)
					t.Errorf("Expect accuracy: %v, Got: %v", wantAcc, resultAcc)
					t.Errorf("Expected status: %v, Got: %v", wantStatus, resultStatus)
				}
			}
		})
	}
}
