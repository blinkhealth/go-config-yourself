// Package yaml provides easy access to yml.Nodes
package yaml

// Copyright 2018 Blink Health LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	yml "gopkg.in/yaml.v3"
)

// var log = logrus.WithField("yaml", "node")

// Tree is just another yaml Node
type Tree struct {
	*yml.Node
	Secret *[]byte
}

type encryptedNode struct {
	Encrypted  bool
	Ciphertext string
	Hash       string
}

type nodePair struct {
	Key   *yml.Node
	Value *yml.Node
}

// UnmarshalYAML implements Decoder
// When a node is decoded into a Tree, it will be tested for secret values
// Tree.Secret will be set accordingly, for use by providers
func (n *Tree) UnmarshalYAML(node *yml.Node) error {
	if node.Kind == yml.MappingNode {
		en := &encryptedNode{}
		if err := node.Decode(&en); err == nil && en.Encrypted {
			cipherBytes, err := base64.StdEncoding.DecodeString(en.Ciphertext)
			if err != nil {
				return fmt.Errorf("Could not unserialize ciphertext as base64")
			}
			n.Secret = &cipherBytes
		}
	}
	n.Node = node
	return nil
}

// MarshalYAML implements Encoder
func (n *Tree) MarshalYAML() (interface{}, error) {
	return n.Node, nil
}

func orderNode(n *yml.Node) {
	if n.Kind == yml.MappingNode {
		nodePairs := make([]*nodePair, 0)
		for i := 0; i < len(n.Content); i++ {
			nodePairs = append(nodePairs, &nodePair{
				Key:   n.Content[i],
				Value: n.Content[i+1],
			})
			i++
		}
		sort.SliceStable(nodePairs, func(i, j int) bool {
			return nodePairs[i].Key.Value < nodePairs[j].Key.Value
		})

		newContent := make([]*yml.Node, 0)
		for _, node := range nodePairs {
			orderNode(node.Value)
			newContent = append(newContent, node.Key, node.Value)
		}
		n.Content = newContent
	}
}

// Serialize returns a byte slice representation of this yaml file
func (n *Tree) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	enc := yml.NewEncoder(writer)
	// This line prevents us from simply using yml.Marshal
	enc.SetIndent(2)
	if n != nil {
		// Sort nodes
		orderNode(n.Node)
	}
	if err := enc.Encode(n); err != nil {
		return nil, fmt.Errorf("Encoding failed: %s", err)
	}
	err := enc.Close()
	writer.Flush()
	return buf.Bytes(), err
}

// Get a value at path from this node
func (n *Tree) Get(path string, dest interface{}) (err error) {
	keyPath := strings.SplitN(path, ".", 2)

	var result *yml.Node
	result, _, err = findInNode(n.Node, keyPath[0])
	if err != nil {
		if err.Error() == "Not found" {
			err = fmt.Errorf("Could not find a value at %s", path)
		}
		return
	}

	if len(keyPath) > 1 {
		// recurse
		v := &Tree{}
		if err = v.UnmarshalYAML(result); err != nil {
			return
		}
		return v.Get(keyPath[1], dest)
	}

	// Check for valid secrets
	v := &Tree{}
	if err = v.UnmarshalYAML(result); err != nil {
		return
	}

	err = v.Decode(dest)
	return
}

// Set a value for a path within this node
func (n *Tree) Set(path string, value interface{}) (err error) {
	keyPath := strings.Split(path, ".")
	currentKey := keyPath[0]

	result, nodeIndex, _ := findInNode(n.Node, currentKey)
	var valueToSet interface{}

	if len(keyPath) > 1 {
		if result == nil {
			// we still have more to create
			if _, err := strconv.Atoi(keyPath[1]); err == nil {
				valueToSet = []interface{}{}
			} else {
				valueToSet = make(map[string]interface{})
			}

			// children does not exist
			n.Node.Content = append(n.Node.Content, newNode(currentKey, valueToSet)...)
			result = n.Node.Content[len(n.Node.Content)-1]
		}

		// recurse
		v := &Tree{}
		if err = v.UnmarshalYAML(result); err != nil {
			return
		}
		return v.Set(strings.Join(keyPath[1:], "."), value)
	}

	// we're setting the actual value
	if result != nil {
		newNodes := newNode("0", value)
		n.Node.Content[nodeIndex] = newNodes[0]
	} else {
		newNodes := newNode(currentKey, value)
		n.Node.Content = append(n.Node.Content, newNodes...)
	}

	return
}

// IsMap returns true if this is a mapping node
func (n *Tree) IsMap() bool {
	return n.Node != nil && n.Node.Kind == yml.MappingNode
}

// IsSlice returns true if this is a sequence node
func (n *Tree) IsSlice() bool {
	return n.Node != nil && n.Node.Kind == yml.SequenceNode
}

// IsEncrypted returns true if this is a sequence node
func (n *Tree) IsEncrypted() bool {
	return n.Secret != nil
}

func findInNode(node *yml.Node, key string) (*yml.Node, int, error) {
	switch node.Kind {
	case yml.MappingNode:
		size := len(node.Content)
		for i := 0; i < size; i++ {
			if node.Content[i].Value == key {
				return node.Content[i+1], i + 1, nil
			}
			i++
		}
	case yml.SequenceNode:
		if index, err := strconv.Atoi(key); err == nil {
			if len(node.Content) > index {
				return node.Content[index], index, nil
			}
		}
	case yml.AliasNode:
		return findInNode(node.Alias, key)
	}

	return nil, 0, fmt.Errorf("Not found")
}

func newNode(key string, value interface{}) (nodes []*yml.Node) {
	if _, err := strconv.Atoi(key); err != nil {
		nodes = append(nodes, &yml.Node{
			Kind:    yml.ScalarNode,
			Value:   key,
			Tag:     "!!str",
			Content: []*yml.Node{},
		})
	}

	v := reflect.ValueOf(value)
	kind := v.Kind()
	switch kind {
	case reflect.Map:
		theMap := &yml.Node{
			Kind:    yml.MappingNode,
			Tag:     "!!map",
			Content: []*yml.Node{},
		}

		for _, k := range v.MapKeys() {
			item := v.MapIndex(k)
			nodes := newNode(k.Interface().(string), item.Interface())
			theMap.Content = append(theMap.Content, nodes...)
		}
		nodes = append(nodes, theMap)
	case reflect.Slice, reflect.Array:
		theSlice := &yml.Node{
			Kind:    yml.SequenceNode,
			Content: []*yml.Node{},
		}

		for i := 0; i < v.Len(); i++ {
			nodes := newNode("0", v.Index(i).Interface())
			theSlice.Content = append(theSlice.Content, nodes...)
		}
		nodes = append(nodes, theSlice)
	default:
		nodes = append(nodes, &yml.Node{
			Kind:    yml.ScalarNode,
			Value:   fmt.Sprint(value),
			Content: []*yml.Node{},
		})
	}

	return
}
