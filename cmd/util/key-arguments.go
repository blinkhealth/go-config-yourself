package util

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v2"
)

// GetKeyArguments reads arguments from the command line and parses them into flags
func GetKeyArguments(ctx *cli.Context) (args map[string]interface{}) {
	commandArgs := ctx.Args().Tail()
	args = make(map[string]interface{})
	for _, flag := range ctx.Command.Flags {
		log.Debugf("looking up flag %s, args remaining: %d", flag.Names(), len(commandArgs))
		name := flag.Names()[0]
		if ctx.IsSet(name) {
			log.Debugf("found flag %s", name)
			switch f := flag.(type) {
			case *cli.StringSliceFlag:
				args[name] = ctx.StringSlice(name)
			case *cli.StringFlag, *cli.GenericFlag:
				args[name] = ctx.String(name)
			case *cli.BoolFlag:
				args[name] = ctx.Bool(name)
			default:
				panic(fmt.Sprintf("I don't know about type %T!\n", f))
			}
		}
	}

	if len(commandArgs) > 0 {
		if ctx.String("provider") == "gpg" && !ctx.IsSet("public-keys") {
			args["public-key"] = commandArgs
		}

		if ctx.String("provider") == "kms" && !ctx.IsSet("key") {
			args["key"] = commandArgs[0]
		}
	}

	return
}
