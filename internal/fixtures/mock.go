package fixtures

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/proglottis/gpgme"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type mockKey string

const (
	MockKMSKey mockKey = "arn:aws:kms:us-east-1:000000000000:key/00000000-AAAA-1111-BBBB-CC22DD33EE44"
	MockGPGKey mockKey = "test-software@blinkhealth.com"
)

// MockGPG setups the environment to mock GPG
func MockGPG() {
	// Hack gpgme into doing its thing
	gpgHomeDir := filepath.Join(Root, "gnupghome")
	os.Setenv("GNUPGHOME", gpgHomeDir)
	os.Unsetenv("GPG_AGENT_INFO")
	info, _ := gpgme.GetEngineInfo()
	_ = gpgme.SetEngineInfo(info.Protocol(), info.FileName(), os.Getenv("GNUPGHOME"))
	ctx, err := gpgme.New()
	if err != nil {
		log.Errorf("Could not create gpgme context: %s", err)
		return
	}

	f, err := os.Open(fmt.Sprintf("%s/pubring.gpg", gpgHomeDir))
	if err != nil {
		log.Errorf("Could not open pubring: %s", err)
		return
	}
	defer f.Close()
	dh, err := gpgme.NewDataFile(f)
	if err != nil {
		log.Errorf("Could not open datafile: %s", err)
		return
	}
	defer dh.Close()

	res, err := ctx.Import(dh)
	if err != nil {
		log.Errorf("Could not import: %s", err)
		return
	}

	if res.Imported != 1 {
		log.Errorf("Import failed: %v", res)
	}
}

// MockAWS supplies dummy credentials for the mock KMS client
func MockAWS() {
	os.Setenv("AWS_ACCESS_KEY_ID", "AGOODACCESSKEYID")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "AVERYSECRETACCESSKEYTHATSNOTBASE64ENC=")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "/dev/null")
	os.Unsetenv("AWS_PROFILE")
	os.Unsetenv("AWS_SESSION_TOKEN")
}

// MockStdin fills a buffer with data and returns a fuction to reset the
func MockStdin(data string) (resumeStdin func(), err error) {
	log.Debugf("Writing fake stdin file with %s", data)
	content := []byte(data)
	var tmpfile *os.File
	tmpfile, err = ioutil.TempFile("", "test-stdin")
	if err != nil {
		return
	}

	if _, err = tmpfile.Write(content); err != nil {
		return
	}

	if _, err = tmpfile.Seek(0, 0); err != nil {
		return
	}

	oldStdin := os.Stdin
	resumeStdin = func() {
		os.Remove(tmpfile.Name())
		os.Stdin = oldStdin
	}

	os.Stdin = tmpfile
	return
}

// MockStdoutAndArgs mocks os.Stdout and os.Args, returning a function that resets them
func MockStdoutAndArgs() func() string {
	originalArgs := os.Args
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	return func() string {
		w.Close()
		os.Args = originalArgs
		os.Stdout = rescueStdout
		out, _ := ioutil.ReadAll(r)
		return string(out)
	}
}

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
	// return nil
	return fmt.Errorf("Unknown provider %s; available providers are %s", value, strings.Join(e.Available, ", "))
}

func (e providerFlag) String() string {
	return e.selected
}

// MockCliCtx returns a mock cli.Context for given arguments
func MockCliCtx(args []string) *cli.Context {
	providerList := []string{"kms", "gpg", "password"}
	app := &cli.App{
		Name: "test-app",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Value:   false,
				Aliases: []string{"v"},
			},
			&cli.BoolFlag{
				Name:    "version",
				Value:   false,
				Aliases: []string{"V"},
			},
			&cli.GenericFlag{
				Name:        "provider",
				Aliases:     []string{"p"},
				Usage:       "Provider",
				Value:       &providerFlag{Available: providerList},
				DefaultText: "aws",
			},
		},
		Version: "beta",
		Commands: []*cli.Command{
			{
				Name: "command",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:   "keypath",
						Value:  "",
						Usage:  "Used internally by the app",
						Hidden: true,
					},
					&cli.StringFlag{
						Name:  "flag",
						Value: "",
						Usage: "a flag",
					},
				},
			},
		},
	}

	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print the version",
	}

	set := flag.NewFlagSet("mock", flag.ContinueOnError)
	set.SetOutput(ioutil.Discard)
	if args != nil {
		if len(args) > 0 {
			os.Setenv("CUR", args[len(args)-1])
		}
		os.Args = args

		for _, flag := range app.Flags {
			if !flagInSet(flag, set) {
				_ = flag.Apply(set)
			}
		}
		_ = set.Parse(args)
	}

	ctx := cli.NewContext(app, set, nil)
	return ctx
}

func flagInSet(flag cli.Flag, set *flag.FlagSet) bool {
	for _, name := range flag.Names() {
		if existing := set.Lookup(name); existing != nil {
			return true
		}
	}
	return false
}
