package yaml_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	fx "github.com/blinkhealth/go-config-yourself/internal/fixtures"
	. "github.com/blinkhealth/go-config-yourself/internal/yaml"
	log "github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)
	os.Exit(m.Run())
}

func TestFromBytes(t *testing.T) {
	yaml, err := FromBytes([]byte{})

	if err != nil {
		t.Fatalf("could not initialize: %s", err)
	}

	serialized, err := yaml.Serialize()
	if err != nil {
		t.Fatalf("Could not serialize empty yaml: %s", err)
	}
	// clear known whitespace tricks
	serialized = bytes.Trim(serialized, "\n")

	if bytes.Equal(serialized, []byte("{}")) {
		t.Fatalf("Serialized yaml is not <{}> but: %s", serialized)
	}
}

func TestFromValue(t *testing.T) {
	value := map[string]interface{}{"string": "asdf"}
	yaml, err := FromValue(value)

	if err != nil {
		t.Fatalf("could not initialize: %s", err)
	}

	serialized, err := yaml.Serialize()
	if err != nil {
		t.Fatalf("Could not serialize empty yaml: %s", err)
	}
	// clear known whitespace tricks
	serialized = bytes.Trim(serialized, "\n")

	if !bytes.Equal(serialized, []byte("string: asdf")) {
		t.Fatalf("Serialized yaml does not match. got: %s want: string: asdf", serialized)
	}
}

func TestFileLoadErrors(t *testing.T) {
	tests := []struct {
		file string
		err  string
	}{
		{fx.Path("non-existent-file"), "Could not read file"},
		{fx.Path("bad/syntax"), "Unable to parse as yaml"},
	}

	for _, tst := range tests {
		tst := tst // capture variable
		t.Run(tst.file, func(t *testing.T) {
			cfg, err := FromPathname(tst.file)

			if err == nil {
				t.Fatalf("Invalid config did not throw error, %d", len(cfg.Content))
			}

			if !strings.Contains(err.Error(), tst.err) {
				t.Fatalf("Invalid configuration exploded with wrong error: %v; %s", err, t.Name())
			}
		})
	}
}

func TestFileLoad(t *testing.T) {
	tests := []struct {
		file     string
		provider string
	}{
		{fx.Path("plaintext"), ""},
		{fx.Path("encrypted.no-provider"), "kms"},
		{fx.Path("encrypted.kms"), "kms"},
		{fx.Path("encrypted.gpg"), "gpg"},
		{fx.Path("encrypted.password"), "password"},
	}

	for _, tst := range tests {
		tst := tst // capture variable
		t.Run(tst.file, func(t *testing.T) {
			yaml, err := FromPathname(tst.file)

			if err != nil {
				t.Fatalf("Valid configuration exploded with %v", err)
			}

			serialized, err := yaml.Serialize()
			if err != nil {
				t.Fatalf("Could not serialize loaded yaml: %s", err)
			}
			// clear known whitespace tricks
			serialized = bytes.Trim(serialized, "\n")
			originalBytes, _ := ioutil.ReadFile(tst.file)
			originalBytes = bytes.Trim(originalBytes, "\n")

			if !bytes.Equal(serialized, originalBytes) {
				t.Fatalf("Serialized yaml does not match. have:\n%s\n---\nwant:\n<%s>", serialized, originalBytes)
			}
		})
	}
}

func TestGetValidValues(t *testing.T) {
	tests := []struct {
		prop  string
		value interface{}
	}{
		{"boolean", true},
		{"empty-list", []interface{}{}},
		{"list", []interface{}{"a", "b", "c"}},
		{"list.0", "a"},
		{"list.4", nil},
		{"nestedList", []interface{}{
			map[string]interface{}{"prop": true},
			map[string]interface{}{"prop": false},
		}},
		{"nestedList.1", map[string]interface{}{"prop": false}},
		{"nestedList.1.prop", false},
		{"number", 1},
		{"object", map[string]interface{}{"key": "value"}},
		{"object.key", "value"},
		{"string", "value"},
	}

	yaml, err := FromPathname(fx.Path("plaintext"))
	if err != nil {
		t.Fatalf("Valid configuration exploded with %v", err)
	}

	for _, tst := range tests {
		tst := tst
		t.Run(tst.prop, func(t *testing.T) {
			var value interface{}
			err := yaml.Get(tst.prop, &value)
			if tst.value == nil && err == nil {
				t.Fatal("Got no error with missing prop")
			}

			if !reflect.DeepEqual(value, tst.value) {
				t.Errorf("Get value (%s) mismatch. Got: %T, want: %T", tst.prop, value, tst.value)
			}
		})
	}
}

func TestSetValidValues(t *testing.T) {
	tests := []struct {
		prop  string
		value interface{}
		query string
	}{
		{"new.boolean", true, ""},
		{"new.empty-list", []interface{}{}, ""},
		{"new.list", []interface{}{"a", "b", "c"}, ""},
		{"new.list.3", "a", ""},
		// new list elements don't magically grow arrays
		{"new.list.5", "a", "new.list.4"},
		{"new.nestedList", []interface{}{
			map[string]interface{}{"prop": true},
			map[string]interface{}{"prop": false},
		}, ""},
		{"new.nestedList.2", map[string]interface{}{"prop": false}, ""},
		{"new.nestedList.2.prop", false, ""},
		{"new.number", 1, ""},
		{"new.object", map[string]interface{}{"key": "value"}, ""},
		{"new.object.newKey", "value", ""},
		{"new.string", "value", ""},
		{"new.doublenestedList.0.0", "value", ""},
		// overwrite
		{"string", true, ""},
		{"nestedList", "[1,2,3]", ""},
		{"list.1", 10, ""},
	}

	yaml, err := FromPathname(fx.Path("plaintext"))
	if err != nil {
		t.Fatalf("Valid configuration exploded with %v", err)
	}

	for _, tst := range tests {
		err := yaml.Set(tst.prop, tst.value)
		if err != nil {
			t.Errorf("Could not set %s: %s", tst.prop, err)
		}

		bytes, _ := yaml.Serialize()

		var value interface{}
		var prop = tst.prop
		if tst.query != "" {
			prop = tst.query
		}
		err = yaml.Get(prop, &value)
		if err != nil {
			log.Debug(string(bytes))
			t.Fatalf("Could not get %s after setting: %s", prop, err)
		}

		if !reflect.DeepEqual(value, tst.value) {
			log.Debug(string(bytes))
			t.Errorf("Get value (%s) mismatch. Got: %T, want: %T", prop, value, tst.value)
		}
	}
}

func TestGetBadEncryptedNode(t *testing.T) {

	yaml, err := FromPathname(fx.Path("bad/secret.ciphertext"))
	if err != nil {
		t.Fatalf("Valid configuration exploded with %v", err)
	}

	var value interface{}
	err = yaml.Get("secret", &value)

	if err == nil {
		t.Fatalf("decryption with bad ciphertext succeeded: %s", value)
	}

	if err.Error() != "Could not unserialize ciphertext as base64" {
		t.Fatalf("received wrong error: %s", err)
	}

	if value != nil {
		t.Fatal("non-nil value returned from decryption")
	}
}
