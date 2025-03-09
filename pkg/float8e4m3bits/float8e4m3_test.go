package F8E4M3

import (
    "testing"
    "math"
    F32 "github.com/shantanu-gontia/float-conv/pkg/float32bits"
)

func TestToFloat32(t *testing.T) {
    testCases := []struct{
        // Input
        input Bits
        scaleFactor int8
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
        }
    }

    for _, tt := range testCases {
        t.Run("ToFloat32", func(t* testing.T) {
            result := tt.input.ToFloat32()
            if result != tt.golden {
                t.Logf("Failed Input Set:\n")
                t.Logf("Input: %0#16b (%0#4x)", tt.input, tt.input)
                t.Errorf("Expected Output: %f (%0#8x). Got: %f (%0#8x)", tt.golden, math.Float32bits(tt.golden), result, math.Float32bits(result))
            }
        })
    }
}
