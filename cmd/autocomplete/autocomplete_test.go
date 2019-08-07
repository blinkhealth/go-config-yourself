package autocomplete_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	comp "github.com/blinkhealth/go-config-yourself/cmd/autocomplete"
	fx "github.com/blinkhealth/go-config-yourself/internal/fixtures"

	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)
	os.Exit(m.Run())
}

func TestCommandAutoComplete(t *testing.T) {
	out := fx.MockStdoutAndArgs()
	comp.CommandAutocomplete(fx.MockCliCtx(nil))

	if out := out(); out != strings.Join(append(allFlags, "command\n"), "\n") {
		t.Fatalf("Invalid output: %s", out)
	}
}

var allFlags = []string{
	"--verbose",
	"--version",
}

func TestListProviderFlags(t *testing.T) {
	tests := []struct {
		args []string
		out  string
	}{
		{[]string{}, strings.Join(allFlags, "\n")},
		{[]string{"-"}, strings.Join(allFlags, "\n")},
		{[]string{"--ver"}, strings.Join(allFlags, "\n")},
		{[]string{"--var"}, ""},
		{[]string{"--provider"}, "gpg\nkms\npassword"},
		{[]string{"--provider", "g"}, "gpg"},
		{[]string{"init", "--provider", "gpg"}, strings.Join(allFlags, "\n")},
	}

	for _, tst := range tests {
		out := fx.MockStdoutAndArgs()

		ctx := fx.MockCliCtx(tst.args)
		comp.ListProviderFlags(ctx)

		if out := out(); out != tst.out {
			log.Errorf("Invalid output: %s", out)
		}
	}

	out := fx.MockStdoutAndArgs()
	ctx := fx.MockCliCtx([]string{"command"})
	ctx.Command = ctx.App.Commands[0]
	comp.ListAllFlags(ctx)
	if out := out(); out != "--keypath\n--flag" {
		log.Errorf("Bad command flags: %s", out)
	}
}

func TestListKeys(t *testing.T) {
	tests := []struct {
		query         string
		results       string
		expectedError error
	}{
		{"", "boolean\nempty-list.\nlist.\nnestedList.\nnumber\nobject.\nstring", nil},
		{"non-existent", "", nil},
		{"n", "nestedList.\nnumber", nil},
		{"nestedList.", "nestedList.0.\nnestedList.1.", nil},
		{"badPath.", "", errors.New("Could not find a value at badPath")},
		{"empty-list.", "", nil},
		{"list.", "list.0\nlist.1\nlist.2", nil},
		{"object.k", "object.key", nil},
	}

	c := fx.Path("plaintext")
	for _, tst := range tests {
		tst := tst
		t.Run(tst.query, func(t *testing.T) {
			out := fx.MockStdoutAndArgs()
			ctx := fx.MockCliCtx(append([]string{c}, tst.query))
			comp.ListKeys(ctx)

			if out := out(); strings.Trim(out, "\n") != tst.results {
				t.Fatalf("Results did not match: %s != %s",
					strings.ReplaceAll(tst.results, "\n", ","),
					strings.ReplaceAll(out, "\n", ","))
			}
		})
	}

	out := fx.MockStdoutAndArgs()
	defer out()
	ctx := fx.MockCliCtx([]string{"something-random"})
	if code := comp.ListKeys(ctx); code != 99 {
		log.Error("List keys did not fail with bad pathname")
	}
}

// func enableDebug() {
// 	if file, err := os.OpenFile("/dev/ttys000", os.O_WRONLY|os.O_APPEND, 0666); err == nil {
// 		log.Debug("sending output to tty")
// 		log.SetLevel(log.DebugLevel)
// 		log.SetOutput(file)
// 	}
// }
