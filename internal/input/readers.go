// Package input provides helpers to work with user inputs
package input

import (
	"bufio"
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
)

// MaxSecretSize specifies how long input read from stdin can be
var MaxSecretSize = 4 * 1024

func readStdin() []byte {
	scanner := bufio.NewScanner(os.Stdin)
	readBytes := make([]byte, 0, MaxSecretSize)
	for scanner.Scan() {
		newBytes := scanner.Bytes()
		if len(newBytes)+len(readBytes) > MaxSecretSize {
			remainingBytes := MaxSecretSize - len(readBytes)
			log.Warnf("Supplied more than %d bytes to read, discarding input", MaxSecretSize)
			readBytes = append(readBytes, newBytes[:remainingBytes]...)
			break
		} else {
			readBytes = append(readBytes, newBytes...)
		}
	}
	return readBytes
}

func checkInputSize(bytes []byte, err error) ([]byte, error) {
	if err != nil {
		return bytes, err
	}

	bytesRead := len(bytes)
	if bytesRead == 0 {
		return bytes, errors.New("Input was empty")
	}

	if bytesRead > MaxSecretSize {
		log.Warnf("Supplied more than %d bytes to read, discarding input", MaxSecretSize)
		bytes = bytes[:MaxSecretSize]
	}

	return bytes, nil
}
