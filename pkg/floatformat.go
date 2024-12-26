package bitVisualization

import "fmt"

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

// Stringer interface for the FloatBitFormat type pointer
func (f FloatBitFormat) String() string {
	return fmt.Sprintf("Sign: %s, Exponent: %s, Mantissa: %s", string(f.Sign), string(f.Exponent), string(f.Mantissa))
}
