package cmd

import (
	"fmt"
	"os"

	"github.com/blinkhealth/go-config-yourself/cmd/autocomplete"
	"github.com/blinkhealth/go-config-yourself/cmd/util"
	"github.com/blinkhealth/go-config-yourself/pkg/file"

	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v2"
)

func init() {
	App.Commands = append(App.Commands, &cli.Command{
		Name:        "init",
		Usage:       "Create a config file",
		ArgsUsage:   "CONFIG_FILE [KEYS]",
		Description: "Prompts the user to select the keys for this config, unless specified creating the config file at CONFIG_FILE",
		Flags:       KeyFlags,
		Action:      initAction,
		ShellComplete: func(ctx *cli.Context) {
			autocomplete.ListProviderFlags(ctx)
			os.Exit(1)
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
