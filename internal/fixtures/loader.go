package fixtures

import (
	"fmt"
	"path"
	"runtime"
	"testing"

	"github.com/blinkhealth/go-config-yourself/pkg/file"
)

var _, _filename, _, _ = runtime.Caller(0)

// Root is the absolute path to the fixtures directory
var Root = path.Join(path.Dir(path.Dir(path.Dir(_filename))), "/test/fixtures")

// Path returns the path to a fixture with a given name
func Path(name string) (path string) {
	return fmt.Sprintf("%s/%s.yaml", Root, name)
}

// LoadFile returns a fixture as a file.ConfigFile or fails the test
func LoadFile(name string, t *testing.T) *file.ConfigFile {
	name = Path(name)
	c, err := file.Load(name)
	if err != nil {
		t.Fatalf("Could not load fixture %s: %s", name, err)
	}

	return c
}
