package floatBit

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
// SaturateInf: If the result overflows, then the result is Inf. If the destination
// format supports signed infinities then the sign is retained
const (
	MakeNaN     OverflowMode = 0
	SaturateMax OverflowMode = 1
	SaturateInf OverflowMode = 2
)

// Stringer interface for OverflowMode
func (o OverflowMode) String() string {
	switch o {
	case MakeNaN:
		return "MakeNaN"
	case SaturateMax:
		return "SaturateMax"
	case SaturateInf:
		return "SaturateInf"
	default:
		return ""
	}
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
	FlushToZero UnderflowMode = 1
)

// Stringer interface for UnderflowMode
func (u UnderflowMode) String() string {
	switch u {
	case SaturateMin:
		return "SaturateMin"
	case FlushToZero:
		return "FlushToZero"
	default:
		return ""
	}
}

// Status represents the status of the operation, with regards to overflow/underflow.
type Status byte

// Status
//
// Fits: The result fits in the destination format. Doesn't overflow or underflow
//
// Overflow: Result doesn't fit in the destination format. Causes overflow.
//
// Underflow: Result doesn't fit in the destination format. Causes underflow.
//
// NoEncoding: Result doesn't have a value in the destination format. This should
// only be used for special cases. Like when the result is a negative infinity,
// but the destination format does not support signed infinites.
const (
	Fits       Status = 0
	Overflow   Status = 1
	Underflow  Status = 2
	NoEncoding Status = 3
)

func (s Status) String() string {
	switch s {
	case Fits:
		return "fits"
	case Overflow:
		return "overflow"
	case Underflow:
		return "underflow"
	case NoEncoding:
		return "no_encoding"
	}
	return ""
}
