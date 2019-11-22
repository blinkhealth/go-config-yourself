package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/blinkhealth/go-config-yourself/cmd/autocomplete"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

func init() {
	const description = "Outputs the plain-text value for `KEYPATH` in `CONFIG_FILE`. If the queried value is a dictionary, it will be encoded as JSON, with all of the encrypted values within decrypted.\n\n" +
		"`KEYPATH` refers to a dot-delimited path to values, see `gcy help keypath` for examples. If a given `KEYPATH` is not found in `CONFIG_FILE`, `gcy get` will fail with exit code 2."

	App.Commands = append(App.Commands, &cli.Command{
		Name:        "get",
		Before:      beforeCommand,
		Aliases:     []string{"show"},
		Usage:       "Output a value from a file",
		ArgsUsage:   "CONFIG_FILE KEYPATH",
		Description: description,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:   "keypath",
				Value:  "",
				Usage:  "Used internally by the app",
				Hidden: true,
			},
		},
		Action: get,
		BashComplete: func(ctx *cli.Context) {
			if ctx.NArg() == 0 {
				os.Exit(1)
			}

			if ctx.NArg() >= 1 {
				autocomplete.ListKeys(ctx)
			}
		},
	})
}

// Get a value from a config file
func get(ctx *cli.Context) error {
	value, err := configFile.Get(ctx.String("keypath"))

	if err != nil {
		return Exit(err, ExitCodeInputError)
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Bool, reflect.Slice, reflect.Map:
		log.Debug("Encoding as json")
		jsonBytes, jsonErr := json.Marshal(v.Interface())
		if jsonErr != nil {
			err := fmt.Sprintf("Could not encode as json: %s", jsonErr)
			return Exit(err, ExitCodeToolError)
		}
		fmt.Println(string(jsonBytes))
	default:
		fmt.Println(value)
	}

	return nil
}
