// Package cmd implements the command line interface
package cmd

// Copyright 2018 Blink Health LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0

import (
	"os"

	"github.com/blinkhealth/go-config-yourself/cmd/autocomplete"
	"github.com/blinkhealth/go-config-yourself/cmd/util"
	log "github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v2"
)

var App = &cli.App{
	Name:  "gcy",
	Usage: "gcy COMMAND CONFIG_FILE [KEYPATH]",
	Authors: []*cli.Author{
		&cli.Author{
			Name:  "Blink Health",
			Email: "opensource@blinkhealth.com",
		},
	},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "verbose",
			Value:   false,
			Aliases: []string{"v"},
		},
	},
	Before: func(ctx *cli.Context) error {
		if ctx.Bool("verbose") {
			log.SetLevel(log.DebugLevel)
			log.Debug("Verbose output enabled")
		}
		return nil
	},
	Commands:              []*cli.Command{},
	EnableShellCompletion: true,
	ShellComplete:         autocomplete.CommandAutocomplete,
	CommandNotFound: func(ctx *cli.Context, name string) {
		// Show help, then error out
		_ = cli.ShowAppHelp(ctx)
		log.Errorf("Unknown command <%s>", name)
		os.Exit(1)
	},
}
var KeyFlags = util.KeyFlags()

// Main main function for go-config-yourself
func Main(version string) {
	log.SetFormatter(&log.TextFormatter{
		DisableLevelTruncation: true,
		DisableTimestamp:       true,
	})

	App.Version = version
	// Use -V and --version
	// https://medium.com/@jdxcode/12-factor-cli-apps-dd3c227a0e46
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print the version",
	}

	cli.InitCompletionFlag.Hidden = true

	cli.HelpPrinter = helpPrinter
	cli.AppHelpTemplate = helpTemplateApp
	cli.CommandHelpTemplate = helpTemplateCmd

	if err := App.Run(os.Args); err != nil {
		log.Debug("Exiting with error")
		exitCode := 1
		if cmdErr, isCmdCoder := err.(CommandError); isCmdCoder {
			exitCode = cmdErr.Code()
		}
		log.Error(err)
		os.Exit(exitCode)
	} else {
		log.Debug("Exited cleanly")
	}
}
