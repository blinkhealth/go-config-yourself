package file_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/blinkhealth/go-config-yourself/pkg/file"

	fx "github.com/blinkhealth/go-config-yourself/internal/fixtures"
	log "github.com/sirupsen/logrus"
)

const testSecret = "asdf"

// Example usage of the package to load a file and operate on it
func Example() {
	// Load your config from a file
	cfg, err := file.Load("./config/my-file.yml")
	if err != nil {
		return
	}

	// Set some values
	if err := cfg.Set("path.to.secret", []byte("🤫")); err != nil {
		return
	}

	// Read them back
	var plaintextValue interface{}
	if plaintextValue, err = cfg.Get("path.to.secret"); err == nil {
		return
	}
	fmt.Printf("The password is %s\n", plaintextValue)
	// Outputs: The password is 🤫

	// Or get all of them at once, decrypted as a map
	mapOfValues, err := cfg.GetAll()
	if err == nil {
		fmt.Printf("The file as a map looks like: %v\n", mapOfValues)
		// Outputs: The file as a map looks like: map[string]...
	}

	// Serialize it as YAML
	if bytes, err := cfg.Serialize(); err == nil {
		fmt.Println(bytes)
		// Outputs:
		// crypto:
		//   provider: password
		//   key: someBase64Key
		// path:
		//   to:
		//     secret:
		//       ciphertext: someBase64Ciphertext
		//       encrypted: true
		//       hash: theSha256HashOfPlaintext
	}
}

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)
	// Mock AWS
	fx.MockAWS()
	fx.MockGPG()
	code := m.Run()
	os.Exit(code)
}

func TestGetValues(t *testing.T) {
	os.Setenv("CONFIG_PASSWORD", "password")
	for _, provider := range []string{"kms", "gpg", "password"} {
		provider := provider
		t.Run(provider, func(t *testing.T) {
			c := fx.LoadFile(fmt.Sprintf("encrypted.%s", provider), t)
			tests := []struct {
				name          string
				expectedValue interface{}
				expectedError error
			}{
				{"string", "value", nil},
				{"number", 1, nil},
				{"boolean", true, nil},
				{"object.key", "value", nil},
				{"list.0", "a", nil},
				{"nestedList.1.prop", false, nil},
				{"secret", testSecret, nil},
				{"fakeValue", nil, fmt.Errorf("Could not find a value at fakeValue")},
			}

			err := c.Set("secret", []byte(testSecret))

			if err != nil {
				t.Fatalf("Unable to encrypt, %s", err)
			}

			for _, tst := range tests {
				tst := tst // capture variable
				t.Run(tst.name, func(t *testing.T) {
					value, err := c.Get(tst.name)

					if err != nil {
						if tst.expectedError == nil {
							t.Fatalf("Could not read %s, unexpected error: %s", tst.name, err)
						}

						if err.Error() != tst.expectedError.Error() {
							t.Fatalf("Expected error %s, got %s", err, tst.expectedError)
						}
					}

					if value != tst.expectedValue {
						t.Errorf("Could not read %s, expected <%s> got <%T>", tst.name, tst.expectedValue, value)
					}
				})
			}
		})
	}
	os.Unsetenv("CONFIG_PASSWORD")
}

func runProviderSetValue(provider string) func(t *testing.T) {
	return func(t *testing.T) {
		c := fx.LoadFile(fmt.Sprintf("encrypted.%s", provider), t)

		err := c.VeryInsecurelySetPlaintext("non-json", []byte("a"))
		if err != nil {
			t.Errorf(err.Error())
		}

		fetched, fetchErr := c.Get("non-json")
		if fetchErr != nil {
			t.Errorf("Failed to fetch new list %v", fetchErr)
		}

		newProp, ok := fetched.(string)
		if !ok {
			t.Errorf("Not a string %T", fetched)
		}

		if newProp != "a" {
			t.Errorf("New property does not match input value %v", newProp)
		}

		listTests := []struct {
			keyPath string
			value   string
			length  int
		}{
			{"empty-list", "[]", 0},
			{"empty-list.0", "value", 1},
			{"json-list", "[4,2]", 2},
			{"newList.0.a", "true", 1},
			{"list.0", "z", 3},
			{"list.10", "z", 4},
		}

		for _, tst := range listTests {
			tst := tst
			t.Run(tst.keyPath, func(t *testing.T) {

				err := c.VeryInsecurelySetPlaintext(tst.keyPath, []byte(tst.value))
				if err != nil {
					t.Errorf(err.Error())
				}

				fetchKey := strings.Split(tst.keyPath, ".")[0]
				fetched, fetchErr := c.Get(fetchKey)

				if fetchErr != nil {
					t.Errorf("Failed to fetch new list %v", fetchErr)
				}

				newList, ok := fetched.([]interface{})
				if !ok {
					t.Errorf("Not a list %T", fetched)
				}

				if len(newList) != tst.length {
					t.Errorf("Tested list (%s) contains %d values. Expected %d values ", fetchKey, len(newList), tst.length)
				}

			})
		}

		// Set properties and values on a list item
		fetched, fetchErr = c.Get("newList.0.a")
		if fetchErr != nil {
			t.Errorf("Failed to fetch new object %v", fetchErr)
		}

		newBool, ok := fetched.(bool)
		if !ok {
			t.Errorf("New object property is not a string but a: %T", fetched)
		}

		if newBool != true {
			t.Errorf("New object property is not the value '1' but '%v'", newBool)
		}

		// Set encrypted values on the main tree
		err = c.Set("secret", []byte(testSecret))
		if err != nil {
			t.Fatal(err.Error())
		}

		encFetch, fetchErr := c.Get("secret")
		if fetchErr != nil {
			t.Errorf("Failed to fetch new encrypted property %v", fetchErr)
		}

		if encFetch != testSecret {
			t.Errorf("New encrypted property is not a string but a: %v", encFetch)
		}

		// Set encrypted values on a list index
		err = c.Set("list.3", []byte(testSecret))
		if err != nil {
			t.Fatal(err.Error())
		}

		encListFetch, fetchErr := c.Get("list.3")
		if fetchErr != nil {
			t.Errorf("Failed to fetch new encrypted list property %v", fetchErr)
		}

		if encListFetch != testSecret {
			t.Errorf("New encrypted list property is not a string but a: %v", encListFetch)
		}

		// Set encrypted values on a list object
		err = c.Set("newList.1.a", []byte(testSecret))
		if err != nil {
			t.Fatal(err.Error())
		}

		encListObjFetch, fetchErr := c.Get("newList.1.a")
		if fetchErr != nil {
			t.Errorf("Failed to fetch new encrypted property %v", fetchErr)
		}

		if encListObjFetch != testSecret {
			t.Errorf("New encrypted list property is not a string but a: %v", encListObjFetch)
		}
	}
}
func TestSetValues(t *testing.T) {
	providers := []string{
		"kms",
		"gpg",
		"password",
	}

	for _, provider := range providers {
		os.Setenv("CONFIG_PASSWORD", "password")
		t.Run(provider, runProviderSetValue(provider))
		os.Unsetenv("CONFIG_PASSWORD")
	}
}

func TestReplaceCrypto(t *testing.T) {
	nonEnc := fx.LoadFile("plaintext", t)
	c := fx.LoadFile("encrypted.kms", t)
	_ = c.Set("secret", []byte(testSecret))

	selection := "2\n"
	tests := []struct {
		name          string
		config        *file.ConfigFile
		args          map[string]interface{}
		stdin         *string
		shouldExplode bool
	}{
		{"empty-crypto", nonEnc, kmsKeyArgs("replace-empty-key"), nil, true},
		{"wrong-key-alias", c, kmsKeyArgs("replace-arg-key"), nil, true},
		{"argument-replace", c, kmsKeyArgs("arn:aws:kms:us-east-1:000000000000:alias/an-alias"), nil, false},
		{"stdin-select", c, nil, &selection, false},
	}

	for _, tst := range tests {
		tst := tst
		t.Run(tst.name, func(t *testing.T) {
			var restoreStdin func()
			var err error

			if tst.stdin != nil {
				restoreStdin, err = fx.MockStdin(*tst.stdin)
				if err != nil {
					t.Error(err)
				}
			}

			_, err = tst.config.Rekey("kms", tst.args)
			if tst.stdin != nil {
				restoreStdin()
			}

			if tst.shouldExplode {
				if err == nil {
					t.Fatalf("Expected error, none raised")
				}
				return
			}

			if err != nil {
				t.Errorf("Unable to replace crypto: %s", err)
			}

			_, err = tst.config.Get("crypto.key")
			if err != nil {
				t.Errorf("Missing crypto.key: %s", err)
			}

			b, _ := tst.config.Serialize()
			if tst.args["key"] != nil {
				if bytes.Contains(b, []byte(tst.args["key"].(string))) {
					t.Errorf("Crypto key not changed")
				}
			}
		})
	}

	_, err := nonEnc.Rekey("kms", kmsKeyArgs("a-dumb-key"))
	if err != nil && err.Error() != "Cannot re-key a config without existing crypto provider" {
		t.Errorf("Able to replace crypto on non encrypted file. %v", err)
	}

	otherKey := strings.Replace(string(fx.MockKMSKey), "us-east-1", "us-west-1", 1)
	newFile, err := c.Rekey("kms", kmsKeyArgs(otherKey))
	if err != nil {
		t.Fatalf("Unable to replace crypto: %s", err)
	}

	_, err = newFile.Get("crypto.key")
	if err != nil {
		t.Fatalf("Missing crypto.key: %s", err)
	}

	b, err := newFile.Serialize()

	if err != nil {
		t.Fatalf("Unable to serialize: %s", err)
	}

	if !bytes.Contains(b, []byte(fmt.Sprintf("key: %s", otherKey))) {
		t.Errorf("Crypto key not changed: %s", b)
	}

}

func TestGetAll(t *testing.T) {
	os.Setenv("CONFIG_PASSWORD", "password")
	for _, provider := range []string{"kms", "gpg", "password"} {
		provider := provider
		t.Run(provider, func(t *testing.T) {
			c := fx.LoadFile(fmt.Sprintf("encrypted.%s", provider), t)
			_ = c.Set("secret", []byte(testSecret))

			data, err := c.GetAll()
			if err != nil {
				t.Fatal(err)
			}

			plaintextValue := data["secret"]

			if plaintextValue != testSecret {
				t.Errorf("Could not decrypt plaintext, %v", plaintextValue)
			}
		})
	}
	os.Unsetenv("CONFIG_PASSWORD")
}

func TestListAllSecrets(t *testing.T) {
	os.Setenv("CONFIG_PASSWORD", "password")
	for _, provider := range []string{"kms", "gpg", "password"} {
		provider := provider
		t.Run(provider, func(t *testing.T) {
			c := fx.LoadFile(fmt.Sprintf("encrypted.%s", provider), t)
			_ = c.Set("secret", []byte(testSecret))

			secrets := c.ListSecrets()

			if len(secrets) != 1 {
				t.Fatalf("Did not find all the secrets, found %d", len(secrets))
			}

			if secrets[0] != "secret" {
				t.Error("Did not find secret")
			}
		})
	}
	os.Unsetenv("CONFIG_PASSWORD")
}

func kmsKeyArgs(key string) map[string]interface{} {
	return map[string]interface{}{"key": key}
}

func TestMissingCrypto(t *testing.T) {
	c := fx.LoadFile("bad/crypto.nil", t)

	value, err := c.Get("secret")

	if err == nil {
		t.Fatal("decryption sans crypto node succeeded")
	}

	_, err = c.GetAll()
	if err == nil {
		t.Fatalf("Get all failed with no error, %s", err)
	}

	if err.Error() != "Unable to decrypt, config file has no `crypto` property, or the crypto provider is not enabled" {
		t.Fatalf("received wrong error: %s", err)
	}

	if value != nil {
		t.Fatal("non-nil value returned from decryption")
	}

	err = c.Set("new-secret", []byte("secret"))

	if err == nil {
		t.Fatal("encyption sans crypto node succeeded")
	}
}

func TestBadRekey(t *testing.T) {
	badKeyConfig := fx.LoadFile("bad/crypto.key", t)
	newConfig, err := badKeyConfig.Rekey("kms", kmsKeyArgs(string(fx.MockKMSKey)))
	if err == nil {
		t.Fatal("rekey sans crypto node succeeded")
	}

	if newConfig != nil {
		t.Fatal("non-nil value returned from bad rekey")
	}

	goodKeyConfig := fx.LoadFile("encrypted.kms", t)
	badKey, _ := badKeyConfig.Get("crypto.key")
	newConfig, err = goodKeyConfig.Rekey("kms", kmsKeyArgs(badKey.(string)))
	if err == nil {
		t.Fatal("rekey with bad target key succeeded")
	}

	if newConfig != nil {
		t.Fatal("non-nil value returned from bad rekey")
	}
}

func TestGetBadSecret(t *testing.T) {
	tests := []struct {
		path string
		err  string
	}{
		{"bad/secret.ciphertext", "Failed decrypt, secret.ciphertext is not valid base64"},
		{"bad/crypto.key", "No AWS credentials found"},
	}

	for _, tst := range tests {
		tst := tst
		t.Run(tst.path, func(t *testing.T) {
			c := fx.LoadFile(tst.path, t)

			value, err := c.Get("secret")

			if err == nil {
				t.Fatalf("decryption with bad ciphertext succeeded: %s", value)
			}

			if err.Error() != tst.err {
				t.Fatalf("received wrong error: %s", err)
			}

			if value != nil {
				t.Fatal("non-nil value returned from decryption")
			}
		})
	}
}

func TestSetBadSecret(t *testing.T) {
	tests := []struct {
		path string
		err  string
	}{
		{"bad/crypto.key", "No AWS credentials found"},
	}

	for _, tst := range tests {
		tst := tst
		t.Run(tst.path, func(t *testing.T) {
			c := fx.LoadFile(tst.path, t)

			err := c.Set("secret", []byte("nope"))

			if err == nil {
				t.Fatal("encryption with bad ciphertext succeeded")
			}

			if err.Error() != tst.err {
				t.Fatalf("received wrong error: %s", err)
			}
		})
	}
}
