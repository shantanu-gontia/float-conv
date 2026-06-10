package main

import (
	"encoding/json"
	"math/big"
	"syscall/js"
	"strconv"

	floatBit "github.com/shantanu-gontia/float-conv/pkg"
	BF16 "github.com/shantanu-gontia/float-conv/pkg/bfloat16bits"
	F16 "github.com/shantanu-gontia/float-conv/pkg/float16bits"
	F32 "github.com/shantanu-gontia/float-conv/pkg/float32bits"
)

type ConversionResult struct {
	Format      string `json:"format"`
	Sign        string `json:"sign"`
	Exponent    string `json:"exponent"`
	Mantissa    string `json:"mantissa"`
	Decimal     string `json:"decimal"`
	Hexfloat    string `json:"hexfloat"`
	Binary      string `json:"binary"`
	Hexadecimal string `json:"hexadecimal"`
	Accuracy    string `json:"accuracy"`
	Error       string `json:"error"`
	Status      string `json:"status"`
}

func parseRoundingMode(mode string) floatBit.RoundingMode {
	switch mode {
	case "rtz":
		return floatBit.RoundTowardsZero
	case "rtposinf":
		return floatBit.RoundTowardsPositiveInf
	case "rtneginf":
		return floatBit.RoundTowardsNegativeInf
	case "rthalfzero":
		return floatBit.RoundHalfTowardsZero
	case "rthalfposinf":
		return floatBit.RoundHalfTowardsPositiveInf
	case "rthalfneginf":
		return floatBit.RoundHalfTowardsNegativeInf
	case "rne":
		return floatBit.RoundNearestEven
	case "rno":
		return floatBit.RoundNearestOdd
	default:
		return floatBit.RoundNearestEven
	}
}

func parseOverflowMode(mode string) floatBit.OverflowMode {
	switch mode {
	case "satinf":
		return floatBit.SaturateInf
	case "satmax":
		return floatBit.SaturateMax
	case "nan":
		return floatBit.MakeNaN
	default:
		return floatBit.SaturateMax
	}
}

func parseUnderflowMode(mode string) floatBit.UnderflowMode {
	switch mode {
	case "flushzero":
		return floatBit.FlushToZero
	case "satmin":
		return floatBit.SaturateMin
	default:
		return floatBit.SaturateMin
	}
}

func convertFloat32(inputStr string, rm, om, um string) ConversionResult {
	roundingMode := parseRoundingMode(rm)
	overflowMode := parseOverflowMode(om)
	underflowMode := parseUnderflowMode(um)

	val, _, err := big.ParseFloat(inputStr, 0, 53, roundingMode.ToBigRoundingMode())
	if err != nil {
		return ConversionResult{Format: "float32", Error: err.Error()}
	}

	floatVal, accuracy, status := F32.FromBigFloat(*val, roundingMode, overflowMode, underflowMode)

	format := floatVal.ToFloatFormat()
	asBigFloat := floatVal.ToBigFloat()

	result := ConversionResult{
		Format:      "Float32",
		Sign:        string(format.Sign),
		Exponent:    string(format.Exponent),
		Mantissa:    string(format.Mantissa),
		Decimal:     asBigFloat.Text('e', -1),
		Hexfloat:    strconv.FormatFloat(float64(floatVal.ToFloat32()), 'x', -1, 32),
		Binary:      formatBinary32(floatVal),
		Hexadecimal: formatHex32(floatVal),
		Accuracy:    accuracyString(accuracy),
		Status:      statusString(status),
	}

	convErr, _ := floatVal.ConversionError(val)
	result.Error = convErr.Text('e', -1)

	return result
}

func convertBFloat16(inputStr string, rm, om, um string) ConversionResult {
	roundingMode := parseRoundingMode(rm)
	overflowMode := parseOverflowMode(om)
	underflowMode := parseUnderflowMode(um)

	val, _, err := big.ParseFloat(inputStr, 0, 53, roundingMode.ToBigRoundingMode())
	if err != nil {
		return ConversionResult{Format: "bfloat16", Error: err.Error()}
	}

	floatVal, accuracy, status := BF16.FromBigFloat(*val, roundingMode, overflowMode, underflowMode)

	format := floatVal.ToFloatFormat()
	asBigFloat := floatVal.ToBigFloat()

	result := ConversionResult{
		Format:      "BFloat16",
		Sign:        string(format.Sign),
		Exponent:    string(format.Exponent),
		Mantissa:    string(format.Mantissa),
		Decimal:     asBigFloat.Text('e', -1),
		Hexfloat:    strconv.FormatFloat(float64(floatVal.ToFloat32()), 'x', -1, 32),
		Binary:      formatBinaryBF16(floatVal),
		Hexadecimal: formatHexBF16(floatVal),
		Accuracy:    accuracyString(accuracy),
		Status:      statusString(status),
	}

	convErr, _ := floatVal.ConversionError(val)
	result.Error = convErr.Text('e', -1)

	return result
}

func convertFloat16(inputStr string, rm, om, um string) ConversionResult {
	roundingMode := parseRoundingMode(rm)
	overflowMode := parseOverflowMode(om)
	underflowMode := parseUnderflowMode(um)

	val, _, err := big.ParseFloat(inputStr, 0, 53, roundingMode.ToBigRoundingMode())
	if err != nil {
		return ConversionResult{Format: "float16", Error: err.Error()}
	}

	floatVal, accuracy, status := F16.FromBigFloat(*val, roundingMode, overflowMode, underflowMode)

	format := floatVal.ToFloatFormat()
	asBigFloat := floatVal.ToBigFloat()

	result := ConversionResult{
		Format:      "Float16",
		Sign:        string(format.Sign),
		Exponent:    string(format.Exponent),
		Mantissa:    string(format.Mantissa),
		Decimal:     asBigFloat.Text('e', -1),
		Hexfloat:    strconv.FormatFloat(float64(floatVal.ToFloat32()), 'x', -1, 32),
		Binary:      formatBinaryF16(floatVal),
		Hexadecimal: formatHexF16(floatVal),
		Accuracy:    accuracyString(accuracy),
		Status:      statusString(status),
	}

	convErr, _ := floatVal.ConversionError(val)
	result.Error = convErr.Text('e', -1)

	return result
}

func convert(this js.Value, args []js.Value) interface{} {
	if len(args) < 5 {
		return "Error: insufficient arguments"
	}

	input := args[0].String()
	format := args[1].String()
	rm := args[2].String()
	om := args[3].String()
	um := args[4].String()

	var result ConversionResult
	switch format {
	case "float32", "fp32":
		result = convertFloat32(input, rm, om, um)
	case "bfloat16", "bf16":
		result = convertBFloat16(input, rm, om, um)
	case "float16", "fp16":
		result = convertFloat16(input, rm, om, um)
	default:
		result = convertFloat32(input, rm, om, um)
	}

	jsonBytes, _ := json.Marshal(result)
	return string(jsonBytes)
}

func accuracyString(acc big.Accuracy) string {
	switch acc {
	case big.Exact:
		return "Exact"
	case big.Above:
		return "Above"
	case big.Below:
		return "Below"
	default:
		return "Unknown"
	}
}

func statusString(status floatBit.Status) string {
	switch status {
	case floatBit.Fits:
		return "Fits"
	case floatBit.Overflow:
		return "Overflow"
	case floatBit.Underflow:
		return "Underflow"
	default:
		return "Unknown"
	}
}

func formatBinary32(val F32.Bits) string {
	v := uint32(val)
	result := "0b"
	for i := 31; i >= 0; i-- {
		if v&(1<<i) != 0 {
			result += "1"
		} else {
			result += "0"
		}
	}
	return result
}

func formatHex32(val F32.Bits) string {
	return "0x" + strconv.FormatUint(uint64(val), 16)
}

func formatBinaryBF16(val BF16.Bits) string {
	v := uint16(val)
	result := "0b"
	for i := 15; i >= 0; i-- {
		if v&(1<<i) != 0 {
			result += "1"
		} else {
			result += "0"
		}
	}
	return result
}

func formatHexBF16(val BF16.Bits) string {
	return "0x" + strconv.FormatUint(uint64(val), 16)
}

func formatBinaryF16(val F16.Bits) string {
	v := uint16(val)
	result := "0b"
	for i := 15; i >= 0; i-- {
		if v&(1<<i) != 0 {
			result += "1"
		} else {
			result += "0"
		}
	}
	return result
}

func formatHexF16(val F16.Bits) string {
	return "0x" + strconv.FormatUint(uint64(val), 16)
}

func main() {
	js.Global().Set("floatConvConvert", js.FuncOf(convert))
	<-make(chan struct{})
}