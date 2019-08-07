package file_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	fx "github.com/blinkhealth/go-config-yourself/internal/fixtures"
	"github.com/blinkhealth/go-config-yourself/pkg/file"
)

func TestConfigLoading(t *testing.T) {
	tests := []struct {
		file          string
		err           string
		shouldExplode bool
	}{
		{"non-existent-file", "Could not read file", true},
		{"main.go", "Could not parse YAML", true},
		{"bad/crypto.type", "Invalid config", true},
		{"encrypted.fake-provider", "Unknown provider <fake-provider>", true},
		{"encrypted.no-provider", "", false},
		{"encrypted.kms", "", false},
		{"encrypted.gpg", "", false},
		{"encrypted.password", "", false},
	}

	for _, tst := range tests {
		tst := tst // capture variable
		t.Run(tst.file, func(t *testing.T) {
			_, err := file.Load(fx.Path(tst.file))

			if tst.shouldExplode {
				if err == nil {
					t.Fatalf("Invalid config did not throw error")
				}

				if !strings.Contains(err.Error(), tst.err) {
					t.Errorf("Invalid configuration exploded with wrong error: %v", err)
				}
			} else if err != nil {
				t.Errorf("Valid configuration exploded with %v", err)
			}

		})
	}
}

func TestCreateConfig(t *testing.T) {
	selection := "0\n"
	impossibleSelection := "128"
	passwordSelection := "password"
	emptyStdin := ""
	nonInteger := "asdf"
	envPassword := map[string]string{
		"CONFIG_PASSWORD": passwordSelection,
	}
	gpgFlags := map[string]interface{}{
		"recipients": []string{string(fx.MockGPGKey)},
	}

	tests := []struct {
		provider      string
		args          map[string]interface{}
		env           *map[string]string
		stdin         *string
		shouldExplode bool
	}{
		// Wrong provider
		{"unknown", nil, nil, nil, true},
		// kms ideal case
		{"kms", kmsKeyArgs(string(fx.MockKMSKey)), nil, nil, false},
		// kms stdin
		{"kms", nil, nil, &selection, false},
		// kms out of bounds selection
		{"kms", nil, nil, &impossibleSelection, true},
		// kms out of bounds selection
		{"kms", nil, nil, &nonInteger, true},
		// gpg ideal case
		{"gpg", gpgFlags, nil, nil, false},
		// gpg out of bounds selection
		{"gpg", nil, nil, &impossibleSelection, true},
		// gpg out of bounds selection
		{"gpg", nil, nil, &nonInteger, true},
		// password env var
		{"password", nil, &envPassword, nil, false},
		// password arg
		{"password", map[string]interface{}{"password": "secret"}, nil, nil, false},
		// password through stdin selection
		{"password", nil, nil, &passwordSelection, false},
		// password with empty stdin
		{"password", nil, nil, &emptyStdin, true},
	}

	for _, tst := range tests {
		tst := tst
		stdin := ""
		if tst.stdin != nil {
			stdin = *tst.stdin
		}
		testName := fmt.Sprintf("%s/args=%d,stdin=%s,env=%v", tst.provider, len(tst.args), stdin, tst.env == nil)
		t.Run(testName, func(t *testing.T) {
			var restoreStdin func()
			var err error

			if tst.env != nil {
				for key, value := range *tst.env {
					os.Setenv(key, value)
				}
			}

			if tst.stdin != nil {
				restoreStdin, err = fx.MockStdin(*tst.stdin)
				if err != nil {
					t.Fatal(err)
				}
			}
			cfg, err := file.Create(tst.provider, tst.args)
			if tst.stdin != nil {
				restoreStdin()
			}
			if tst.env != nil {
				for key := range *tst.env {
					os.Unsetenv(key)
				}
			}

			if tst.shouldExplode {
				if err == nil {
					crypto, _ := cfg.Get("crypto")
					t.Fatalf("Expected to explode, did not, %v", crypto)
				}
				return
			} else if err != nil {
				t.Fatalf("Unable to create: %v", err)
			}

			cryptoNode, err := cfg.Get("crypto")

			if err != nil {
				t.Fatalf("Could not get crypto node: %v", err)
			}

			_, ok := cryptoNode.(map[string]interface{})

			if !ok {
				t.Errorf("Could not cast crypto node, wrong type: %T", cryptoNode)
				t.Fail()
			}

			if cfg.Provider != tst.provider {
				t.Fatalf("Created with wrong provider: want %s, got: %s", cfg.Provider, tst.provider)
			}

		})
	}
}
