# float-conv

Go application to convert a number to a floating-point format and print details like the hex/bit representation of the
converted number, conversion error etc.

* IEEE-754 Float32
* BFloat16

## Usage

```bash
float-conv --num=<number> [--format=<format>] [--round-mode=<rounding mode>] [--overflow-mode=<overflow mode>] [--underflow-mode=<underflow-mode>]
```

* The option `--num` is used to provide the input number. The number can be input as either decimal, scientific or
in hexfloat formats.
* The `--format` option is used to specify which floating point format to convert to. Valid values are `float32` [*Default*], `bfloat16`.
* The `--round-mode` option is used to specify which rounding mode to use, if the number cannot be exactly represented in the desired format. Supported options are
  * `rtz`: Round Towards Zero
  * `rtposinf`: Round Towards Positive Infinity
  * `rtneginf`: Round Towards Negative Infinity
  * `rthalfzero`: Round to the closest number, break ties by rounding towards zero
  * `rthalfposinf`: Round to the closest number, break ties by rounding towards positive infinity
  * `rthalfneginf`: Round to the closest number, break ties by rounding towards negative infinity
  * `rne`: Round towards the nearest even number (LSB is 0) [*Default*]
  * `rno`: Round towards the nearest odd number (LSB is 1)
* The `--overflow-mode` option is used to specify the response if the number (in magnitude) is larger than the maximum representable (in magnitude) in the target format. Supported options are
  * `satinf`: Saturate the number to infinity with the same sign as the input
  * `satmax`: Saturate the number to the maximum possible in the format, with the same sign as the input [*Default*]
* The `--underflow-mode` option is used to specify the response if the number (in magnitude) is smaller than the minimum representable (in magnitude) in the target format. Supported options are
  * `flushzero`: Flush the number to 0. If the target format supports signed zeros, then the sign is same as that of the input
  * `satmin`: Saturates the number to the minimum representable, with the same sign as the input [*Default*]
* The `--precision` flag is used to augment the precision to use when parsing the input. The default is 53.

## Example

```bash
$ float-conv --num=0.125 --format=float32
Float32
|Sign|Exponent|               Mantissa|
|   0|00111110|00000000000000000000000|
Decimal: 1.25e-01
Hexfloat: 0x1p-03
Conversion Error: 0e+00 (Exact)
Binary: 0b00111110000000000000000000000000
Hexadecimal: 0x3e000000

$ float-conv --num=1e-256 --format=bfloat16 --underflow-mode=flushzero
BFloat16
|Sign|Exponent|Mantissa|
|   0|00000000| 0000000|
Decimal: 0e+00
Hexfloat: 0x0p+00
Conversion Error: -1e-256 (Below)
Binary: 0b0000000000000000
Hexadecimal: 0x0000
UNDERFLOW
```

## Roadmap

* Add support for Half-Precision floating points (float16/binary16)
* Add support for OCP Microscaling 8-bit Formats
