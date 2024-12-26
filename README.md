# float-conv

Go application to convert a number to a floating-point format and print details like the hex/bit representation of the
converted number, conversion error etc.

* IEEE-754 Float32
* BFloat16

## Usage

```bash
float-conv <number> [-f=<format>] [-t=<as_table>] [-p=<precision>]
```

* The first and only positional argument is the input number. The number can be input as either decimal, scientific or
in hexfloat formats.
* The `-f` option is used to specify which floating point format to convert to. Valid values are `float32` (default), `bfloat16`. If unspecified, the default value is 'float32'
* The `-p` option is used to specify the precision to use for parsing the input value. If unspecified, the default value is 53.
* The `-t` option controls whether the bitfields are printed as a Table or a single line string (See example below). If unspecified, the default value is true.

## Example

```bash
$ float-conv 0.125 -f=float32
Sign: 0, Exponent: 00111110, Mantissa: 00000000000000000000000

$ float-conv 0.125 -f=float32 -t=true
|Sign|Exponent|               Mantissa|
|   0|00111110|00000000000000000000000|
```

## Roadmap

* Add proper argument parsing support
* Add CLI arguments that allow outputting specific formatting (like just the binary representation or just the 
hexadecimal representation), so this can be used for piping to other commands.
* Add support for OCP Microscaling 8-bit Formats
* When converting input string to floating-point formats, allow specification of the rounding modes. Currently, the plan
is to support Rounding to Nearest Even (RNE), Rounding Towards Zero (RTZ), Truncation (default, one currently used)