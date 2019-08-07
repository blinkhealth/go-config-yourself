package cmd

import "fmt"

// ExitCode enumerates error codes available to commands
type ExitCode int

const (
	// ExitCodeToolError means we did something silly
	ExitCodeToolError ExitCode = 2
	// ExitCodeInputError means the user did something silly
	ExitCodeInputError ExitCode = 99
)

// CommandError encodes a message and exit codes
type CommandError struct {
	message interface{}
	code    ExitCode
}

// Exit creates a new CommandError
func Exit(message interface{}, code ExitCode) error {
	return CommandError{
		message: message,
		code:    code,
	}
}

func (err CommandError) Error() string {
	return fmt.Sprintf("%v", err.message)
}

// Code returns the error code
func (err CommandError) Code() int {
	return int(err.code)
}
