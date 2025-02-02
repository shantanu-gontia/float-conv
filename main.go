package main

import (
	"errors"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strings"

	floatBit "github.com/shantanu-gontia/float-conv/pkg"
	BF16 "github.com/shantanu-gontia/float-conv/pkg/bfloat16bits"
	F32 "github.com/shantanu-gontia/float-conv/pkg/float32bits"
)

type ProgramInputs struct {
	input  big.Float
	format string
	rm     floatBit.RoundingMode
	om     floatBit.OverflowMode
	um     floatBit.UnderflowMode
}

func main() {
	// Declare cmdline flags
	valStrPtr := flag.String("num", "nil", "Input floating point number. Required.")
	formatStrPtr := flag.String("format", "float32",
		"Target floating point format (Supported values are float32)")
	rouningModeStrPtr := flag.String("round-mode", "rne", "Rounding Mode to use (Supported values are rne, "+
		"rno, rtz, rtposinf, rtneginf, rthalfzero, rthalfposinf, rthalfneginf)")
	overflowModeStrPtr := flag.String("overflow-mode", "satmax",
		"Overflow behavior (Supported values are satmax, satinf, nan)")
	underflowModeStrPtr := flag.String("underflow-mode", "satmin",
		"Overflow behavior (Supported values are satmin, flushzero)")
	precisionPtr := flag.Uint("precision", 53, "Precision to use for the input floating point")

	// Parse the flags
	flag.Parse()

	// Parse the rounding mode
	roundingMode, err := parseRoundingMode(rouningModeStrPtr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Parse the overflow mode
	overflowMode, err := parseOverflowMode(overflowModeStrPtr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Pares the underflow mode
	underflowMode, err := parseUnderflowMode(underflowModeStrPtr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Input Value
	val, _, err := big.ParseFloat(*valStrPtr, 0, *precisionPtr, roundingMode.ToBigRoundingMode())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Call the appropriate handlers
	switch strings.ToLower(*formatStrPtr) {
	case "float32":
		fallthrough
	case "fp32":
		handleFloat32(val, roundingMode, overflowMode, underflowMode)
	case "bfloat16":
		fallthrough
	case "bf16":
		handleBFloat16(val, roundingMode, overflowMode, underflowMode)
	}

}

// Call the appropriate functions and methods required to put together the information to print for FP32
func handleFloat32(bf *big.Float, rm floatBit.RoundingMode, om floatBit.OverflowMode, um floatBit.UnderflowMode) {
	// First we print the type
	fmt.Println("Float32")

	// Get the Float32 Value
	floatVal, accuracy, status := F32.FromBigFloat(*bf, rm, om, um)

	// Print the bits in a table
	fmt.Print(floatVal.ToFloatFormat().AsTable())

	// Print the decimal value
	asBigFloat := floatVal.ToBigFloat()
	fmt.Printf("Decimal: %s\n", asBigFloat.Text('e', -1))

	// Print the hexfloat value
	fmt.Printf("Hexfloat: %x\n", floatVal.ToFloat32())

	// Print the conversion error
	conv, err := floatVal.ConversionError(bf)
	var convStr string
	if err == nil {
		convStr = conv.Text('e', -1)
	} else {
		convStr = "NaN"
	}
	fmt.Printf("Conversion Error: %s (%s)\n", convStr, accuracy)

	// Print the bits in binary
	fmt.Printf("Binary: %0#32b\n", floatVal)

	// Print the bits in hexadecimal
	fmt.Printf("Hexadecimal: %0#8x\n", floatVal)

	if status != floatBit.Fits {
		fmt.Printf("%s\n", strings.ToUpper(status.String()))
	}
}

// Call the appropriate functions and methods required to put together the information to print for BF16
func handleBFloat16(bf *big.Float, rm floatBit.RoundingMode, om floatBit.OverflowMode, um floatBit.UnderflowMode) {
	// First we print the type
	fmt.Println("BFloat16")

	// Get the BFloat16 Value
	floatVal, accuracy, status := BF16.FromBigFloat(*bf, rm, om, um)

	// Print the bits in a table
	fmt.Print(floatVal.ToFloatFormat().AsTable())

	// Print the decimal value
	asBigFloat := floatVal.ToBigFloat()
	fmt.Printf("Decimal: %s\n", asBigFloat.Text('e', -1))

	// Print the hexfloat value
	fmt.Printf("Hexfloat: %x\n", floatVal.ToFloat32())

	// Print the conversion error
	conv, err := floatVal.ConversionError(bf)
	var convStr string
	if err == nil {
		convStr = conv.Text('e', -1)
	} else {
		convStr = "NaN"
	}
	fmt.Printf("Conversion Error: %s (%s)\n", convStr, accuracy)

	// Print the bits in binary
	fmt.Printf("Binary: %0#16b\n", floatVal)

	// Print the bits in hexadecimal
	fmt.Printf("Hexadecimal: %0#4x\n", floatVal)

	if status != floatBit.Fits {
		fmt.Printf("%s\n", strings.ToUpper(status.String()))
	}
}

// Underflow mode to use
func parseUnderflowMode(underflowModeStrPtr *string) (floatBit.UnderflowMode, error) {
	var underflowMode floatBit.UnderflowMode
	switch strings.ToLower(*underflowModeStrPtr) {
	case "satmin":
		underflowMode = floatBit.SaturateMin
	case "flushzero":
		underflowMode = floatBit.FlushToZero
	default:
		return underflowMode, errors.New("Unsupported UnderflowMode " + *underflowModeStrPtr)
	}
	return underflowMode, nil
}

// Overflow mode to use
func parseOverflowMode(overflowModeStrPtr *string) (floatBit.OverflowMode, error) {
	var overflowMode floatBit.OverflowMode
	switch strings.ToLower(*overflowModeStrPtr) {
	case "nan":
		overflowMode = floatBit.MakeNaN
	case "satmax":
		overflowMode = floatBit.SaturateMax
	case "satinf":
		overflowMode = floatBit.SaturateInf
	default:
		return overflowMode, errors.New("Unsupported overflow mode " + *overflowModeStrPtr)
	}
	return overflowMode, nil
}

// Rounding mode to use
func parseRoundingMode(roundingModeStrPtr *string) (floatBit.RoundingMode, error) {
	var roundMode floatBit.RoundingMode
	switch strings.ToLower(*roundingModeStrPtr) {
	case "rne":
		roundMode = floatBit.RoundNearestEven
	case "rno":
		roundMode = floatBit.RoundNearestOdd
	case "rtz":
		roundMode = floatBit.RoundTowardsZero
	case "rtposinf":
		roundMode = floatBit.RoundTowardsPositiveInf
	case "rtneginf":
		roundMode = floatBit.RoundTowardsNegativeInf
	case "rthalfzero":
		roundMode = floatBit.RoundHalfTowardsZero
	case "rthalfposinf":
		roundMode = floatBit.RoundHalfTowardsPositiveInf
	case "rthalfneginf":
		roundMode = floatBit.RoundHalfTowardsNegativeInf
	default:
		return roundMode, errors.New("Unsupported rounding mode " + *roundingModeStrPtr)
	}
	return roundMode, nil
}
