package main

import (
	"fmt"
	"os"

	floatBitPrint "github.com/shantanu-gontia/float-conv/pkg"
)

func init() {

	const (
		name  = "float-conv"
		usage = "Converts the given floating-point string literal to the specified format." +
			"Prints the bit and hex representations of the bits of the target format value and the conversion error"
	)
}

func main() {
	// Parse CLI Arguments
	parsedArgs, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		return
	}

	// Setup the initial value for the arguments
	input := parsedArgs.input
	format := parsedArgs.format
	asTable := parsedArgs.asTable

	switch format {
	case "float32":
		floatVal, _ := input.Float32()
		floatBitVal := floatBitPrint.Float32{}.FromFloat(floatVal)
		if asTable {
			fmt.Printf("%s", floatBitVal.ToFloatFormat().AsTable())
		} else {
			fmt.Println(floatBitVal.ToFloatFormat())
		}
	case "bfloat16":
		floatVal, _ := input.Float32()
		floatBitVal := floatBitPrint.BFloat16{}.FromFloat(floatVal)
		if asTable {
			fmt.Printf("%s", floatBitVal.ToFloatFormat().AsTable())
		} else {
			fmt.Println(floatBitVal.ToFloatFormat())
		}
	default:
		panic("Unsupported argument encountered for format")
	}

}
