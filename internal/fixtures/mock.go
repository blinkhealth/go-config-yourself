package fixtures

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/proglottis/gpgme"
	log "github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v2"
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

// MockCliCtx returns a mock cli.Context for given arguments
func MockCliCtx(args []string) *cli.Context {
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

	set := flag.NewFlagSet("", 0)
	if args != nil {
		if len(args) > 0 {
			os.Setenv("CUR", args[len(args)-1])
		}
		os.Args = args
		_ = set.Parse(args)
	}
	ctx := cli.NewContext(app, set, nil)
	return ctx
}
