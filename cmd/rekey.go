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
	App.Commands = append(App.Commands, &cli.Command{
		Name:  "rekey",
		Usage: "Re-encrypts all the secret values with a new crypto config",
		Description: `Re-keying a file will update its crypto config to new values. A file
   can be re-keyed with the same provider and new keys, or to a completely
   different provider. After the config is changed, go-config-yourself will re-encrypt
   all encrypted values with the new configuration. If no keys are specified,
   go-config-yourself will prompt the user to select them from a list.`,
		ArgsUsage: "CONFIG_FILE [key...]",
		Action:    rekey,
		Flags:     KeyFlags,
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
