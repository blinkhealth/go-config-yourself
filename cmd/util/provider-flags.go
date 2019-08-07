package util

import (
	"fmt"
	"strings"

	pvd "github.com/blinkhealth/go-config-yourself/pkg/provider"
	cli "gopkg.in/urfave/cli.v2"
)

// providerFlag implements urfave/cli.v2/Generic
type providerFlag struct {
	Available []string
	Default   string
	selected  string
}

func (e *providerFlag) Set(value string) error {
	for _, enum := range e.Available {
		if enum == value {
			e.selected = value
			return nil
		}
	}

	return fmt.Errorf("Unknown provider %s; available providers are %s", value, strings.Join(e.Available, ", "))
}

func (e providerFlag) String() string {
	return e.selected
}

// KeyFlags returns a list of cli flags for key-related operations
func KeyFlags() (flags []cli.Flag) {
	flags = append(flags, &cli.GenericFlag{
		Name:    "provider",
		Aliases: []string{"p"},
		Usage:   fmt.Sprintf("The provider to encrypt values with (one of: %s)", strings.Join(pvd.ProviderList, ", ")),
		Value:   &providerFlag{Available: pvd.ProviderList},
	})

	for _, flag := range pvd.AvailableFlags() {
		if flag.Repeatable {
			f := &cli.StringSliceFlag{
				Name:  flag.Name,
				Usage: flag.Description,
			}
			if flag.EnvVarName != "" {
				f.EnvVars = []string{flag.EnvVarName}
			}

			flags = append(flags, f)
		} else {
			f := &cli.StringFlag{
				Name:  flag.Name,
				Usage: flag.Description,
			}
			if flag.EnvVarName != "" {
				f.EnvVars = []string{flag.EnvVarName}
			}

			if flag.Default != "" {
				f.Value = flag.Default
			}
			flags = append(flags, f)
		}
	}

	return
}
