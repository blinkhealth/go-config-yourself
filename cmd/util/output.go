package util

import (
	"fmt"
	"io/ioutil"

	file "github.com/blinkhealth/go-config-yourself/pkg/file"
	cli "gopkg.in/urfave/cli.v2"
)

// SerializeAndWrite a config file to disk
func SerializeAndWrite(path string, cfg *file.ConfigFile) (err error) {
	outBytes, err := cfg.Serialize()
	if err != nil {
		return cli.Exit(err, 2)
	}

	err = ioutil.WriteFile(path, outBytes, 0644)
	if err != nil {
		err = fmt.Errorf("Unable to write configuration to %s: %v", path, err)
		return cli.Exit(err, 2)
	}
	return
}
