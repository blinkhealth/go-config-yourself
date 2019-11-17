package cmd

import (
	"os"

	"github.com/blinkhealth/go-config-yourself/cmd/autocomplete"
	"github.com/blinkhealth/go-config-yourself/cmd/util"
	"github.com/blinkhealth/go-config-yourself/pkg/file"

	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v2"
)

func init() {
	const description = "Re-encrypts all the secret values with specified arguments in `CONFIG_FILE`.\n\n" +
		"By default, it will reuse the same provider for this operation, unless `--provider` is passed. If needed, `gcy rekey` will query your provider for a list of keys to choose from when using the `aws` or `gpg` providers, and a password will be prompted for when using the `password` provider."

	App.Commands = append(App.Commands, &cli.Command{
		Name:        "rekey",
		Usage:       "Re-encrypts all the secret values with a new crypto config",
		Description: description,
		ArgsUsage:   "CONFIG_FILE",
		Action:      rekey,
		Flags:       KeyFlags,
		ShellComplete: func(ctx *cli.Context) {
			if ctx.NArg() == 0 {
				if !autocomplete.ListProviderFlags(ctx) {
					return
				}
			}

			if ctx.NArg() < 2 {
				// revert to file searching
				os.Exit(1)
			}
		},
	})
}

// Rekey a config file
func rekey(ctx *cli.Context) (err error) {
	fileName := ctx.Args().Get(0)
	originalConfig, err := file.Load(fileName)
	if err != nil {
		return Exit(err, ExitCodeInputError)
	}

	if originalConfig.Provider == "" {
		// Help non-migrated old configs
		log.Warnf("Unspecified crypto.provider for %s, defaulting to kms", fileName)
		originalConfig.Provider = "kms"
	}

	if !ctx.IsSet("provider") {
		log.Warnf("Re-encrypting with same crypto.provider: %s", originalConfig.Provider)
		_ = ctx.Set("provider", originalConfig.Provider)
	}

	args := util.GetKeyArguments(ctx)
	newProvider := ctx.String("provider")

	newConfig, err := originalConfig.Rekey(newProvider, args)
	if err != nil {
		return Exit(ExitCodeToolError, 3)
	}

	if err := util.SerializeAndWrite(fileName, newConfig); err != nil {
		return Exit(err, ExitCodeToolError)
	}
	log.Info("Re-encryption successful")

	return nil
}
