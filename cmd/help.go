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
	"bytes"
	"io"
	"regexp"
	"strings"

	"github.com/mitchellh/go-wordwrap"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
	cli "github.com/urfave/cli/v2"
)

var markdownCodeBlock = regexp.MustCompile("`([^`]+)`")
var boldReplacement = []byte("\033[1m$1\033[0m")

var helpPrinter = func(w io.Writer, templ string, data interface{}) {
	var buf bytes.Buffer
	log.Debug(data)

	termCols, _, err := terminal.GetSize(0)
	if err != nil {
		termCols = 80
	}
	cli.HelpPrinterCustom(&buf, templ, data, map[string]interface{}{
		"padded": func(data string) string {
			wrapped := ""
			for _, line := range strings.Split(data, "\n") {
				wrapped = wrapped + "\n" + wordwrap.WrapString(line, uint(termCols-2))
			}
			return strings.Replace(wrapped, "\n", "\n  ", -1)
		},
	})
	boldened := markdownCodeBlock.ReplaceAll(buf.Bytes(), boldReplacement)
	_, err = w.Write(boldened)
	if err != nil {
		log.Error(err)
	}
}

var helpTemplateApp = "`{{.Name}}` - {{.Usage}}" + `

USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}{{if .Commands}} command [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}
{{if .VisibleFlags}}
GLOBAL OPTIONS:
    {{range .VisibleFlags}}{{.}}
    {{end}}{{end}}{{if .Commands}}
COMMANDS:
{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}
   {{end}}{{if .Version}}
VERSION:
   {{.Version}}
   {{end}}
`

var helpTemplateCmd = "`{{ .HelpName }}` - {{.Usage}}" + `

USAGE:
	{{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [command options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[arguments...]{{end}}{{end}}
{{if .VisibleFlags}}
OPTIONS:
	{{range .VisibleFlags}}{{.}}
	{{end}}{{end}}{{if .Description}}
DESCRIPTION:
	{{padded .Description}}{{end}}

`

var exampleConfig = `# "crypto" is a reserved top-level property, it stores metadata about this file
crypto:
  # provider specifies the encryption/decryption backend
  provider: kms
  key: arn:aws:kms:an-aws-region:an-account:alias/an-alias

# Comments will be preserved by gcy
aBoolean: false

# Lists look like
hats:
  - worn-by: Alice
    color: rainbow
  - worn-by: Bob
    color: transparent

# Use nested objects to group related values
wanna:
  store-secrets:
    in-your-repo: go-config-yourself

# Encrypted values are indicated by the reserved "encrypted" property
secretColor:
  encrypted: true
  ciphertext: cGluaw==
  hash: f0e4c2f76c58916ec258f246851bea091d14d4247a2fc3e18694461b1816e13b

`

func helpCommandAction(ctx *cli.Context) error {
	App.CommandNotFound(ctx, ctx.Command.Name)
	return nil
}

func init() {
	keypathHelp := &cli.Command{
		Name:     "keypath",
		Hidden:   true,
		HelpName: "gcy help keypath",
		HideHelp: true,
		Usage:    "Shows help about KEYPATH",
		Action:   helpCommandAction,
		Description: "`KEYPATH` is a dot-delimited path to values, that is a list of keys joined by the `.` character. Integers in a `KEYPATH` specify the index of an item in a list.\n\n" +
			"For `CONFIG_FILE`:\n\n" +
			exampleConfig +
			"These would be valid `KEYPATH` examples:\n\n" +
			"  - `wanna.store-secrets.in-your-repo` is the key path for `go-config-yourself`\n" +
			"  - `hats.0.worn-by` => `Alice`\n" +
			"  - `hats.1.color` => `transparent`\n" +
			"  - `aBoolean` => `true`" +
			"  - `secretColor` => `pink`",
	}

	configfileHelp := &cli.Command{
		Name:     "config-file",
		Hidden:   true,
		HelpName: "gcy help config-file",
		HideHelp: true,
		Usage:    "Shows help about CONFIG_FILE",
		Action:   helpCommandAction,
		Description: "Config files are [YAML](https://yaml.org/) files with nested dictionaries representing a configuration tree. Storing encrypted values requires the presence of a `crypto` property with configuration for that provider, but the rest is up to you. `gcy` keeps keys ordered alphabetically, doing its best-effort to keep comments in place. Here's a typical example of such a file, using the `kms` provider:\n\n" +
			exampleConfig +
			"The recommended location for config files for projects is in the `config/` directory of a repository. A common usage pattern is to start with a `config/defaults.yml` file and then add override files for each environment the application will run in, like so:\n\n" +
			`- your-awesome-project/
			  | - config/
			    | - defaults.yml
			    | - staging.yml
			    | - production.yml

` +
			"In the above scenario, you may store defaults or placeholders in `defaults.yml` with no encryption, while storing only the necessary secrets to override these placeholders in separate files. `staging.yml` and `production.yml` will only contain overrides to be applied on top of `defaults.yml`. `gcy set` automatically adds placeholder values to `defaults.yml` after storing secrets in environment-specific files.",
	}
	App.Commands = append(App.Commands, keypathHelp, configfileHelp)
}
