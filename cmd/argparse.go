package main

import (
	"fmt"
	"math/big"
	"strings"
)

type Args struct {
	input     big.Float
	precision uint
	asTable   bool
}

type noInputError struct{}

type invalidArgumentError struct {
	arg string
}

func (n noInputError) Error() string {
	return "No Input. Please provide an input number."
}

func (e invalidArgumentError) Error() string {
	return fmt.Sprintf("Invalid argument: %s", e.arg)
}

func parseArgs(args []string) (Args, error) {
	// Construct default set of arguments
	parsedArgs := Args{}

	if len(args) == 0 {
		return Args{}, noInputError{}
	}

	// First, we parse the positional input
	// The first positional input must be the number to convert
	x, _, err := big.ParseFloat(args[0], 0, 53, big.ToZero)
	if err != nil {
		return parsedArgs, err
	}

	// If no other arguments are provided, then we are done
	if len(args) == 1 {
		return parsedArgs, nil
	}

	// Right now, we have a separate state for every different valid parameter :(
	const (
		reset                           = 0
		encounteredFirstDash            = 1
		encounteredSecondDash           = 2
		encounteredEqualsAfterFirstDash = 3
		encounteredSpaceAfterFirstDash  = 4
		insideParsingPrecision          = 5
		done                            = 255
	)

	// We use a state-machine to parse the rest of the arguments
	joinedArgs := strings.Join(args[1:], " ")

	currentState := reset
	currentIndex := 0
	for currentState != done {
		// If we're at the end, and a proper parse wasn't finished
		if currentIndex >= len(joinedArgs) && currentState != reset {
			return parsedArgs, invalidArgumentError{"-"}
		}
		currentChar := joinedArgs[currentIndex]
		if currentState == reset {
			fmt.Println("Initial")
			// This is the state either when the parser is beginning, or some argument was finished parsing
			// At this point we are looking for dashes. Any other character is an invalid argument
			if currentChar == '-' {
				currentState = encounteredFirstDash
			} else {
				// Invalid Argument. Parse till next space or end
				indexBegin := currentIndex
				errString := getErrorArgument(joinedArgs, indexBegin)
				return parsedArgs, invalidArgumentError{errString}
			}
		} else if currentState == encounteredFirstDash {
			fmt.Println("First Dash")
			// If the first dash is encountered, then the next element could either be something we expect
			if currentChar == '-' { // Check for dash
				currentState = encounteredSecondDash
				continue
			}
		} else if currentState == encounteredSecondDash {
			fmt.Println("Second Dash")
			// DoubleDash arguments are not supported at the moment. Parse till the next space or end
			// And report as unsupported argument
			indexBegin := currentIndex - 1
			errString := getErrorArgument(joinedArgs, indexBegin)
			return parsedArgs, invalidArgumentError{errString}
		} else if currentState == encounteredSpaceAfterFirstDash {

		} else if currentState == insideParsingPrecision {

		}
		currentIndex++
	}

	parsedArgs.input = *x
	return parsedArgs, nil
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
