package yaml

import (
	"fmt"
	"io/ioutil"

	yml "gopkg.in/yaml.v3"
)

// FromValue returns Tree a data map
func FromValue(data map[string]interface{}) (fy *Tree, err error) {
	doc := &yml.Node{
		Kind:    yml.MappingNode,
		Content: []*yml.Node{},
		Style:   yml.FoldedStyle,
	}

	for k, v := range data {
		doc.Content = append(doc.Content, newNode(k, v)...)
	}

	fy = &Tree{}
	err = fy.UnmarshalYAML(doc)
	return
}

// FromBytes returns Tree a byte slice
func FromBytes(data []byte) (fy *Tree, err error) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = fmt.Errorf("Unable to parse as yaml: %s", panicErr)
		}
	}()
	err = yml.Unmarshal(data, &fy)
	return
}

// FromPathname returns Tree a file path
func FromPathname(path string) (fy *Tree, err error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Could not read file %s", path)
	}

	// If the file is empty, but still a file we can work with it
	if len(buf) == 0 {
		buf = []byte("{}")
	}

	return FromBytes(buf)
}
