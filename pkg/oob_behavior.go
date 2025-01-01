package floatBit

import "fmt"

// Overflow mode controls the behavior of rounding for the cases, where
// after rounding, the result overflows (i.e. the result is larger than the
// maximum value in magnitude that can be represented in the destination format)
type OverflowMode uint8

// OverflowMode
//
// MakeNaN: If the result overflows, then the result is NaN
//
// SaturateMax: If the result overflows, then the result is Max Normal value
// in the result format (sign is retained)
//
// Overflow: If the result overflows, then the result is Inf. If the destination
// format supports signed infinities then the sign is retained
const (
	MakeNaN     OverflowMode = 0
	SaturateMax OverflowMode = 1
	Overflow    OverflowMode = 2
)

// OverflowError is used to report whether an arithmetic operation (in our case we're limited to rounding)
// causes an overflow (in the destination format)
type OverflowError struct {
	base  string
	value string
}

func (e OverflowError) Error() string {
	return fmt.Sprintf("Overflow Error. Value %s could not be represented in %s.", e.value, e.base)
}

// Underflow mode controls the behavior of rounding for the cases, where
// after rounding, the result underflows (i.e. the result is smaller than the
// minimum value in magnitude that can be represented in the destination format)
type UnderflowMode uint8

// UnderflowMode
//
// SaturateMin: If the result underflows, then sets the result to the minimum
// subnormal value in the target format (sign is retained)
//
// Underflow: If the result underflows, then flushes the result to 0
// If the format supports signed zeros, then the sign is retained.
const (
	SaturateMin UnderflowMode = 0
	Underflow   UnderflowMode = 1
)

// UnderflowError is used to report whether an arithmetic operation (in our case we're limited to rounding)
// causes an underflow (in the destination format)
type UnderflowError struct {
	base  string
	value string
}

func (e UnderflowError) Error() string {
	return fmt.Sprintf("Underflow Error. Value %s could not be represented in %s.", e.value, e.base)
}
