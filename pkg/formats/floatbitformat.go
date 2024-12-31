package floatBit

import (
	"fmt"
	"strings"
	"text/tabwriter"
)

// FloatBitFormat type represents a generalized unpacked bit-representation of a floating-point format
// Each of the three fields Sign, Exponent and Mantissa are represented as byte slices which contain
// either '0' or '1' as their entries
type FloatBitFormat struct {
	Sign     []byte
	Exponent []byte
	Mantissa []byte
}

// FloatBitFormatter interface supports conversion of the input type to a FloatFormat type.
type FloatBitFormatter interface {
	ToFloatFormat() FloatBitFormat
}

// Default Stringer interface for the FloatBitFormat type pointer, to print out as Sign: <>, Exponent: <>, Mantissa: <>
func (f FloatBitFormat) String() string {
	return fmt.Sprintf("Sign: %s, Exponent: %s, Mantissa: %s", string(f.Sign), string(f.Exponent), string(f.Mantissa))
}

// Return the FloatBitFormat as a string with the bits sequentially printed out as if it were a floating point
// format
func (f FloatBitFormat) toBitStr() string {
	return fmt.Sprintf("%s%s%s", string(f.Sign), string(f.Exponent), string(f.Mantissa))
}

// Returns a string with the FloatBitFormat formatted as a Table inside
// The Table has Sign, Exponent, and Mantissa as the header row
// Example:
//
//	Sign|  Exponent|    Mantissa|
//	   0|  00000000|     0000000|
func (f FloatBitFormat) AsTable() string {

	const (
		signStr     = "Sign"
		exponentStr = "Exponent"
		mantissaStr = "Mantissa"
	)

	sb := strings.Builder{}

	writer := tabwriter.NewWriter(&sb, 0, 0, 0, ' ', tabwriter.AlignRight|tabwriter.Debug)
	fmt.Fprintf(writer, "\t%s\t%s\t%s\t\n", signStr, exponentStr, mantissaStr)
	fmt.Fprintf(writer, "\t%s\t%s\t%s\t\n", string(f.Sign), string(f.Exponent), string(f.Mantissa))
	writer.Flush()

	return sb.String()
}
