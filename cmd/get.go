package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/blinkhealth/go-config-yourself/cmd/autocomplete"

	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v2"
)

func init() {
	App.Commands = append(App.Commands, &cli.Command{
		Name:      "get",
		Before:    beforeCommand,
		Aliases:   []string{"show"},
		Usage:     "Get a config value out of CONFIG_FILE",
		ArgsUsage: "CONFIG_FILE KEYPATH",
		Description: `Prints to stdout the decrypted value for KEYPATH. It will print out JSON if the value at KEYPATH is a dictionary or list.
A KEYPATH is a sequence of keys delimited by the dot character. Integers in a keypath specify the index of an item in a list.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:   "keypath",
				Value:  "",
				Usage:  "Used internally by the app",
				Hidden: true,
			},
		},
		Action: get,
		ShellComplete: func(ctx *cli.Context) {
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
