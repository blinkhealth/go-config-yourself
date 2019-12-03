package autocomplete_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	comp "github.com/blinkhealth/go-config-yourself/cmd/autocomplete"
	fx "github.com/blinkhealth/go-config-yourself/internal/fixtures"
	diff "github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"
)

var allFlags = []string{
	"--verbose",
	"--version",
	"--provider",
}

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)
	os.Exit(m.Run())
}

func TestCommandAutoComplete(t *testing.T) {
	mockedStdOut := fx.MockStdoutAndArgs()
	mockCtx := fx.MockCliCtx(nil)
	comp.CommandAutocomplete(mockCtx)
	output := mockedStdOut()
	options := strings.Split(output, "\n")
	expected := append(allFlags, "command", "")

	if !diff.Equal(options, expected) {
		t.Errorf("Invalid output, got: %v, expected %v", options, expected)
	}
}

func TestListProviderFlags(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"empty-args", []string{}, allFlags},
		{"single-dash", []string{"-"}, allFlags},
		{"complete-ver", []string{"--ver"}, []string{"--verbose", "--version"}},
		{"expect-empty", []string{"--var"}, []string{}},
		{"provider", []string{"--provider"}, []string{"gpg", "kms", "password"}},
		{"provider-query", []string{"--provider", "g"}, []string{"gpg"}},
		{"post-provider", []string{"--provider", "gpg", "-"}, []string{"--verbose", "--version"}},
	}

	for _, tst := range tests {
		tst := tst
		t.Run(tst.name, func(t *testing.T) {
			out := fx.MockStdoutAndArgs()
			ctx := fx.MockCliCtx(tst.input)
			comp.ListProviderFlags(ctx)

			outStr := out()
			trimmed := strings.TrimSuffix(outStr, "\n")
			got := []string{}
			if trimmed != "" {
				got = strings.Split(trimmed, "\n")
			}
			if !diff.Equal(got, tst.expected) {
				t.Errorf("Invalid output: %+v", got)
			}
		})
	}
}

func TestSubCommandFlags(t *testing.T) {
	out := fx.MockStdoutAndArgs()
	ctx := fx.MockCliCtx([]string{"command"})
	ctx.Command = ctx.App.Commands[0]
	comp.ListAllFlags(ctx)

	outStr := out()
	trimmed := strings.TrimSuffix(outStr, "\n")
	got := []string{}
	if trimmed != "" {
		got = strings.Split(trimmed, "\n")
	}

	if !diff.Equal(got, []string{"--flag"}) {
		t.Errorf("Invalid command flags: %v", got)
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
		t.Error("List keys did not fail with bad pathname")
	}
}

// func enableDebug() {
// 	if file, err := os.OpenFile("/dev/ttys000", os.O_WRONLY|os.O_APPEND, 0666); err == nil {
// 		log.Debug("sending output to tty")
// 		log.SetLevel(log.DebugLevel)
// 		log.SetOutput(file)
// 	}
// }
