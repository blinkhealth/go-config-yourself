package cmd

import (
	"fmt"
	"os"

	"github.com/blinkhealth/go-config-yourself/cmd/autocomplete"
	"github.com/blinkhealth/go-config-yourself/cmd/util"
	"github.com/blinkhealth/go-config-yourself/pkg/file"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

func init() {
	description := multiLineDescription(
		"Creates a YAML config file at `CONFIG_FILE`.",

		"If needed, `gcy init` will query your provider for a list of keys to choose from when using the `aws` or `gpg` providers, and a password will be prompted for when using the `password` provider. `gcy init` will select the `aws` provider by default, and you can override it with the `--provider` flag.",

		"See `gcy help config-file` for more information about `CONFIG_FILE`.",
	)

	App.Commands = append(App.Commands, &cli.Command{
		Name:        "init",
		Usage:       "Create a config file",
		ArgsUsage:   "CONFIG_FILE",
		Description: description,
		Flags:       KeyFlags,
		Action:      initAction,
		BashComplete: func(ctx *cli.Context) {
			if ctx.NArg() == 0 {
				if !autocomplete.ListProviderFlags(ctx) {
					return
				}
			}

			if ctx.NArg() < 2 {
				// offer file searching
				os.Exit(2)
			}
		},
	})
}

// Init creates a config file
func initAction(ctx *cli.Context) error {
	if !ctx.Args().Present() {
		return showUsage(ctx, "Destination to save config file missing")
	}

	target := ctx.Args().Get(0)
	if _, err := os.Stat(target); err == nil {
		msg := fmt.Sprintf("A file at %s already exists, won't overwrite", target)
		return Exit(msg, ExitCodeInputError)
	}

	log.Infof("Creating config at %s", target)
	if !ctx.IsSet("provider") {
		_ = ctx.Set("provider", "kms")
	}

	configData, err := file.Create(ctx.String("provider"), util.GetKeyArguments(ctx))
	if err != nil {
		return Exit(err, ExitCodeToolError)
	}

	if err := util.SerializeAndWrite(target, configData); err != nil {
		return Exit(err, ExitCodeToolError)
	}

	log.Infof("Created config at %s", target)
	return nil
}
