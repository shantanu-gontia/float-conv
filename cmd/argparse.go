package main

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type Args struct {
	input     big.Float
	precision uint
	format    string
	asTable   bool
}

type noInputError struct{}

type invalidArgumentError struct {
	arg string
}

type unspecifiedArgumentError struct {
	currentlyParsing byte
}

type invalidArgumentValueError struct {
	currentlyParsing byte
	valueParsed      string
}

func (n noInputError) Error() string {
	return "No Input. Please provide an input number."
}

func (e invalidArgumentError) Error() string {
	return fmt.Sprintf("Invalid argument: %s", e.arg)
}

func (e unspecifiedArgumentError) Error() string {
	return fmt.Sprintf("Unspecified argument: %s", string(e.currentlyParsing))
}

func (e invalidArgumentValueError) Error() string {
	parsingString := string(e.currentlyParsing)
	if e.currentlyParsing == 0 {
		parsingString = "Input"
	}
	return fmt.Sprintf("Invalid value for argument %s: %s", parsingString, e.valueParsed)
}

func parseArgs(args []string) (Args, error) {
	// Construct default set of arguments
	parsedArgs := Args{
		precision: 53,
		format:    "float32",
		asTable:   false,
	}

	if len(args) == 0 {
		return Args{}, noInputError{}
	}

	// First, we parse the positional input
	// The first positional input must be the number to convert
	x, _, err := big.ParseFloat(args[0], 0, 53, big.ToZero)
	if err != nil {
		return parsedArgs, invalidArgumentValueError{byte(0), args[0]}
	}

	// If no other arguments are provided, then we are done
	if len(args) == 1 {
		return parsedArgs, nil
	}

	// Right now, we have a separate state for every different valid parameter :(
	const (
		reset                = 0
		encounteredFirstDash = 1
		parsedArgChar        = 2
		startParsingValue    = 3
		keepParsingValue     = 4
		doneParsingValue     = 5
		unspecifiedFlag      = 253
		invalidFlag          = 254
		done                 = 255
	)

	stateString := make(map[int]string)
	stateString[reset] = "reset"
	stateString[encounteredFirstDash] = "encounteredFirstDash"
	stateString[parsedArgChar] = "parsedArgChar"
	stateString[startParsingValue] = "startParsingValue"
	stateString[keepParsingValue] = "keepParsingValue"
	stateString[doneParsingValue] = "doneParsingValue"
	stateString[unspecifiedFlag] = "unspecifiedFlag"
	stateString[invalidFlag] = "invalidFlag"
	stateString[done] = "done"

	// We use a state-machine to parse the rest of the arguments
	joinedArgs := strings.Join(args[1:], " ")

	// Make a buffer where we store the flag we're currently in the process of parsing. This buffer is managed
	// by the state machine. This will store the invalid flag to print, if we reach the invalid state
	parsedBuffer := make([]byte, 0)
	currentlyParsing := byte(0)

	// Start in the reset state and parse characters one by one
	state := reset
	currentIndex := 0

	for state != done {
		// fmt.Println("State:", stateString[state])
		// fmt.Println("Parse Buffer:", string(parsedBuffer))
		// fmt.Println("Currently Parsing:", string(currentlyParsing))
		// fmt.Println("Current Index:", currentIndex, "/", len(joinedArgs))
		// fmt.Println("----")
		if state == reset {
			// Reset the state
			parsedBuffer = make([]byte, 0)
			currentlyParsing = byte(0)
			// Special Case, if we're at the end of parsing in this state, that means we were successful
			if currentIndex == len(joinedArgs) {
				state = done
				continue
			}
			// The reset state implies either we're at the beginning of parsing, or we just finished parsing
			// some other flag. At this point '-' is the only allowed character. Anything else means invalid argument
			currentChar := joinedArgs[currentIndex]
			parsedBuffer = append(parsedBuffer, currentChar)
			if currentChar == '-' {
				state = encounteredFirstDash
			} else if currentChar == ' ' {
				// Ignore spaces between successful parses
				continue
			} else {
				state = invalidFlag
			}

		} else if state == encounteredFirstDash {
			if currentIndex == len(joinedArgs) {
				state = invalidFlag
				continue
			}

			// After encountering the first dash, the only supported values are the single-character
			// flags that we suppport. Anything else results in the invalid state.
			currentChar := joinedArgs[currentIndex]
			parsedBuffer = append(parsedBuffer, currentChar)

			// Add a check for every supported flag
			state = parsedArgChar
			if currentChar == 'p' {
				// Update the flag we're currently Parsing
				currentlyParsing = 'p'
			} else if currentChar == 'f' {
				currentlyParsing = 'f'
			} else if currentChar == 't' {
				currentlyParsing = 't'
			} else {
				state = invalidFlag
			}
		} else if state == parsedArgChar {
			if currentIndex == len(joinedArgs) {
				state = unspecifiedFlag
				continue
			}

			// We've parsed the single character to represent the flag. The next character must be a space
			// or a equals sign
			currentChar := joinedArgs[currentIndex]
			parsedBuffer = append(parsedBuffer, currentChar)

			// We only support a single space (multiple whitespaces is invalid)
			if currentChar == ' ' || currentChar == '=' {
				// After space we start parsing the value itself
				state = startParsingValue
			} else {
				state = invalidFlag
			}
		} else if state == startParsingValue {
			if currentIndex == len(joinedArgs) {
				state = unspecifiedFlag
				continue
			}
			// At this point we've passed string for a supported argument and a space or an equals sign. What follows
			// now till the next space or EOF is the value to parse

			// Clear the parsedBuffer to start parsing the value
			parsedBuffer = make([]byte, 0)

			currentChar := joinedArgs[currentIndex]
			// If we encounter space, then we keep ourselves in the same state. value parsing doesn't properly start
			// until the first non-space character is encountered
			if currentChar == ' ' {
				continue
			} else {
				parsedBuffer = append(parsedBuffer, currentChar)
				state = keepParsingValue
			}
		} else if state == keepParsingValue {
			// If EOF while parsing, then we're done parsing
			if currentIndex == len(joinedArgs) {
				state = doneParsingValue
				continue
			}

			currentChar := joinedArgs[currentIndex]

			// If space encountered while parsing, we're done parsing this argument
			if currentChar == ' ' {
				state = doneParsingValue
			} else {
				// Otherwise, keep filling the parseBuffer and staying in this state
				parsedBuffer = append(parsedBuffer, currentChar)
			}
		} else if state == doneParsingValue {
			// Parse the actual value
			switch currentlyParsing {
			case 'p':
				{
					val, err := parseP(string(parsedBuffer))
					if err != nil {
						return parsedArgs, invalidArgumentValueError{currentlyParsing, string(parsedBuffer)}
					}
					parsedArgs.precision = val
				}
			case 'f':
				{
					val, err := parseF(string(parsedBuffer))
					if err != nil {
						return parsedArgs, invalidArgumentValueError{currentlyParsing, string(parsedBuffer)}
					}
					parsedArgs.format = val
				}
			case 't':
				{
					val, err := parseT(string(parsedBuffer))
					if err != nil {
						return parsedArgs, invalidArgumentValueError{currentlyParsing, string(parsedBuffer)}
					}
					parsedArgs.asTable = val
				}
			default:
				{
					// If currentlyParsing was not one of the expected values, then that means the argument was invalid
					state = invalidFlag
					continue

				}
			}
			state = reset
			continue
		} else if state == invalidFlag {
			// Invalid Flag encountered, Stop Parsing and return the current state of the parsed Buffer
			return parsedArgs, invalidArgumentError{string(parsedBuffer)}
		} else if state == unspecifiedFlag {
			// Unspecified Flag encountered, Stop Parsing and return the current state of the parsed Buffer
			return parsedArgs, unspecifiedArgumentError{currentlyParsing}
		}
		currentIndex++
	}

	// If successfully parsed other args, then parse the input again with the correct precision
	x, _, err = big.ParseFloat(args[0], 0, parsedArgs.precision, big.ToZero)
	if err != nil {
		return parsedArgs, invalidArgumentValueError{byte(0), args[0]}
	}

	parsedArgs.input = *x
	return parsedArgs, nil
}

// Function to parse the p argument
func parseP(input string) (uint, error) {
	val, err := strconv.ParseUint(input, 0, 32)
	return uint(val), err
}

// Function to parse the f argument
func parseF(input string) (string, error) {
	switch input {
	case "float":
		fallthrough
	case "float32":
		return "float32", nil
	case "bfloat16":
		return "bfloat16", nil
	default:
		return "", invalidArgumentValueError{}
	}
}

func parseT(input string) (bool, error) {
	val, err := strconv.ParseBool(input)
	return val, err
}

// Search till the end of haystack or till the next space is encountered starting at currentIndex, and return the
// string between
func getErrorArgument(haystack string, currentIndex int) string {
	retBytes := make([]byte, 0)
	for i := currentIndex; i < len(haystack); i++ {
		currentChar := haystack[i]
		if currentChar == ' ' {
			break
		}
		retBytes = append(retBytes, currentChar)
	}
	return string(retBytes)
}
