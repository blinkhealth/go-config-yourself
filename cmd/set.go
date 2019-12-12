package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blinkhealth/go-config-yourself/cmd/autocomplete"
	"github.com/blinkhealth/go-config-yourself/cmd/util"
	"github.com/blinkhealth/go-config-yourself/internal/input"
	"github.com/blinkhealth/go-config-yourself/pkg/file"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

func init() {
	description := multiLineDescription(
		"Stores a value at `KEYPATH`, encrypting it by default, and saves it to `CONFIG_FILE`.",

		"`KEYPATH` is a dot-delimited path to values, see `gcy help keypath` for examples.",

		"`gcy set` prompts for input, unless a value is provided via `stdin` or the `--input-file` flag. Values will be interpreted with golang’s default JSON parser before storage, so for example the string `“true”` will be stored as the boolean `true`. Due to existing AWS KMS service limitations, `gcy set` will read up to 4096 bytes before exiting with an error and closing its input.",

		"A properly configured `crypto` property must exist `CONFIG_FILE` for encryption to succeed, `gcy set` will exit with a non-zero status code otherwise. See `gcy help config-file` for more information about `CONFIG_FILE`.",

		"If a `defaults` or `default` file with the same extension as `CONFIG_FILE` exists in the same directory, `gcy set` will add a nil value for `KEYPATH` in said file.",
	)

	App.Commands = append(App.Commands, &cli.Command{
		Name:        "set",
		Before:      beforeCommand,
		Aliases:     []string{"edit"},
		Usage:       "Set a value in CONFIG_FILE at KEYPATH",
		ArgsUsage:   "CONFIG_FILE KEYPATH",
		Description: description,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:   "keypath",
				Value:  "",
				Usage:  "Used internally by the app",
				Hidden: true,
			},
			&cli.BoolFlag{
				Name:    "plain-text",
				Value:   false,
				Usage:   "Store the value as plain text with no encryption",
				Aliases: []string{"p"},
			},
			&cli.StringFlag{
				Name:    "input-file",
				Value:   "",
				Usage:   "Use the specified file path instead of prompting for input from `stdin`",
				Aliases: []string{"i"},
			},
		},
		BashComplete: func(ctx *cli.Context) {
			argCount := ctx.NArg()

			if argCount == 0 {
				autocomplete.ListAllFlags(ctx)
				if !ctx.IsSet("input-file") {
					if _, ok := autocomplete.LastFlagIs("input-file"); ok {
						os.Exit(1)
					}
				}
			}

			if argCount == 1 {
				autocomplete.ListKeys(ctx)
			}

			os.Exit(1)
		},
		Action: set,
	})
}

const noCryptoError = "Unable to store an encrypted value for '%s', use --plain-text to store as a non-encrypted value"

//Set saves an encrypted or plaintext value on the file
func set(ctx *cli.Context) error {
	keyPath := ctx.String("keypath")

	if keyPath == "crypto" || strings.HasPrefix(keyPath, "crypto.") {
		return Exit(fmt.Errorf("Unable to modify `crypto` property, use `rekey` instead."), ExitCodeInputError)
	}

	isPlainText := ctx.Bool("plain-text")
	if !configFile.HasCrypto() && !isPlainText {
		// Won't store a plaintext unless we very explicitly ask for it
		message := fmt.Sprintf(noCryptoError, keyPath)
		return Exit(errors.New(message), ExitCodeInputError)
	}

	var plainText []byte
	var err error
	if file := ctx.String("input-file"); file != "" {
		plainText, err = input.ReadFile(file)
	} else {
		prompt := fmt.Sprintf("Enter value for “%s”", keyPath)
		plainText, err = input.ReadSecret(prompt, !isPlainText)
	}

	if err != nil {
		return Exit(err, ExitCodeInputError)
	}

	if isPlainText {
		err = configFile.VeryInsecurelySetPlaintext(keyPath, plainText)
	} else {
		err = configFile.Set(keyPath, plainText)
	}

	if err != nil {
		return Exit(fmt.Sprintf("Could not set %s: %s", keyPath, err), ExitCodeToolError)
	}

	target := ctx.Args().Get(0)
	if err := util.SerializeAndWrite(target, configFile); err != nil {
		return Exit(err, ExitCodeToolError)
	}

	log.Infof("Value set at %s", keyPath)
	// update defaults file if write was successful
	updateDefaultsFile(target, keyPath)

	return nil
}

func updateDefaultsFile(target string, keyPath string) {
	if strings.HasPrefix(filepath.Base(target), "default") {
		return
	}
	configFolder := filepath.Dir(target)
	extension := filepath.Ext(target)
	if configFolder != "" {
		configFolder = fmt.Sprintf("%s/", configFolder)
	}

	for _, name := range []string{"default", "defaults"} {
		candidate := fmt.Sprintf("%s%s%s", configFolder, name, extension)
		if _, err := os.Stat(candidate); !os.IsNotExist(err) {
			log.Debugf("Found defaults file: %s", candidate)
			defaultsFile, err := file.Load(candidate)
			if err == nil {
				_, err := defaultsFile.Get(keyPath)
				if err != nil && strings.Contains(err.Error(), "Could not find a value") {
					if err := defaultsFile.VeryInsecurelySetPlaintext(keyPath, nil); err == nil {
						// Don't panic if it doesn't get updated
						if util.SerializeAndWrite(candidate, defaultsFile) != nil {
							log.Infof("Updated value in defaults file %s", candidate)
						}
						return
					}
				}
			}
		}
	}
}
