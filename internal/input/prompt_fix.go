package input

import (
	"os"

	"github.com/chzyer/readline"
)

// https://github.com/manifoldco/promptui/issues/49
// This is a hack that solves two issues: 1. bells sounding when using the arrow keys during a prompt, and, 2. writing interactive output to stdout

type stderr struct{}

func (s *stderr) Write(b []byte) (int, error) {
	if len(b) == 1 && b[0] == 7 {
		return 0, nil
	}
	return os.Stderr.Write(b)
}

func (s *stderr) Close() error {
	return os.Stderr.Close()
}

func init() {
	readline.Stdout = &stderr{}
}
