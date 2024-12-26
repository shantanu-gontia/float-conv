package bitVisualization

import "fmt"

// FloatFormat type represents a generalized unpacked bit-representation of a floating-point format
// Each of the three fields Sign, Exponent and Mantissa are represented as byte slices which contain
// either '0' or '1' as their entries
type FloatFormat struct {
	Sign     []byte
	Exponent []byte
	Mantissa []byte
}

// FloatFormatter interface supports conversion of the input type to a FloatFormat type.
type FloatFormatter interface {
	toFloatFormat() FloatFormat
}

// Stringer interface for the FloatFormat type
func (f *FloatFormat) String() string {
	return fmt.Sprintf("Sign: %s, Exponent: %s, Mantissa: %s", string(f.Sign), string(f.Exponent), string(f.Mantissa))
}
