package cmd

const (
	longUsageStr = `
float-conv is a tool to convert the input floating point number into different formats 
	like binary32 (IEEE-754, FP32), binary16 (half-float, FP16), BFloat16, OCP FP8 formats. Inputs can be decimal
	or hexadecimal floating point numbers.

	Other information like conversion error can also be printed, along with status flags like OVERFLOW/UNDERFLOW.

	Formats Supported (formats are case-insensitive in cmdline flags):
	* binary32 (Float32, FP32)
	* binary16 (Float16, FP16)
	* bfloat16 (BF16)
	* MXFP8E4M3 (fp8e4m3)
	* MXFP8E5M2 (fp8e5m2)
	
	Example usage:
	$ float-conv 1.23
	Float32
	sign| exponent| mantissa
		0   000000  00000000
	
	decimal: 1.23
	conversion error: 0
	binary: 00000000000000
	hexadecimal 0xabcde
	
	Float16
	sign| exponent| mantissa
		0   000000  00000000
	
	decimal: 1.23
	conversion error: 0
	binary: 00000000000000
	hexadecimal: 0xabcde
`
)
