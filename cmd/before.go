package cmd

import (
	"github.com/blinkhealth/go-config-yourself/pkg/file"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

// The configfile instance for this command session
var configFile *file.ConfigFile

// sets the config file and keypath for commands
func beforeCommand(ctx *cli.Context) (err error) {
	log.Debugf("running beforeCommand with %d args", ctx.NArg())
	if ctx.NArg() < 2 {
		return Exit("Missing arguments", ExitCodeInputError)
	}

	configFile, err = file.Load(ctx.Args().Get(0))
	if err != nil {
		return Exit(err, ExitCodeInputError)
	}

	keyPath := ctx.Args().Get(1)

	if err := ctx.Set("keypath", keyPath); err != nil {
		return Exit(err, ExitCodeToolError)
	}

	return nil
}

func showUsage(ctx *cli.Context, message string) error {
	_ = cli.ShowCommandHelp(ctx, ctx.Command.Name)
	return Exit(message, ExitCodeInputError)
}
