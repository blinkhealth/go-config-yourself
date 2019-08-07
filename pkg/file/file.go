// Copyright 2018 Blink Health LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0

// Package file provides an API to work encrypted config files
package file

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/blinkhealth/go-config-yourself/internal/yaml"
	"github.com/blinkhealth/go-config-yourself/pkg/provider"

	log "github.com/sirupsen/logrus"
)

// ConfigFile wraps parsed yaml data and provides an interface to interact with it
type ConfigFile struct {
	// A yaml tree
	data *yaml.Tree
	// A crypto provider to operate on `data`
	crypto provider.Crypto
	// The name of this config file's provider, one of `kms`, `gpg`, or `password`
	Provider string
}

// HasCrypto tells whether this file has a crypto provider or not
func (cfg *ConfigFile) HasCrypto() bool {
	if cfg.crypto != nil {
		return cfg.crypto.Enabled()
	}
	return false
}

// GetAll decrypts all secrets and returns the file as a map
func (cfg *ConfigFile) GetAll() (tree map[string]interface{}, err error) {
	tree = map[string]interface{}{}
	allValues := map[string]*yaml.Tree{}
	if err = cfg.data.Decode(&allValues); err != nil {
		return
	}

	for k, value := range allValues {
		decrypted, err := decryptNode(value, cfg.crypto)
		if err != nil {
			return tree, err
		}
		tree[k] = decrypted
	}

	return tree, nil
}

// Get returns the value at this dot-delimited `keyPath`
func (cfg *ConfigFile) Get(keyPath string) (value interface{}, err error) {
	node := &yaml.Tree{}
	err = cfg.data.Get(keyPath, &node)
	if err != nil {
		if err.Error() == "Could not unserialize ciphertext as base64" {
			err = fmt.Errorf("Failed decrypt, %s.ciphertext is not valid base64", keyPath)
		}
		return nil, err
	}

	value, err = decryptNode(node, cfg.crypto)
	return
}

// Rekey creates a copy of this file, initializing its crypto provider with given arguments, and reencrypts all secrets. The original ConfigFile will not be modified.
//
// The user may be prompted for details if connected to a TTY and these are not provided by `providerArgs`
func (cfg *ConfigFile) Rekey(providerName string, providerArgs map[string]interface{}) (newFile *ConfigFile, err error) {
	if !cfg.HasCrypto() {
		return newFile, errors.New("Cannot re-key a config without existing crypto provider")
	}

	log.Debugf("Creating copy for %s, %v", providerName, providerArgs)
	newFile, err = Create(providerName, providerArgs)
	if err != nil {
		return
	}

	nodes := cfg.data.Content
	for i := 0; i < len(nodes); i++ {
		if nodes[i].Value == "crypto" {
			// don't copy over the crypto node
			i++
			continue
		}
		newNode := &yaml.Tree{}
		if err = newNode.UnmarshalYAML(nodes[i]); err != nil {
			return
		}
		newFile.data.Content = append(newFile.data.Content, newNode.Node)
	}

	for _, keyPath := range cfg.ListSecrets() {
		log.Debugf("re-encrypting %s", keyPath)

		var value interface{}
		if value, err = cfg.Get(keyPath); err != nil {
			log.Debugf("Failed to decrypt secret at <%s>", keyPath)
			return nil, err
		}

		if err = newFile.Set(keyPath, []byte(fmt.Sprintf("%s", value))); err != nil {
			log.Debugf("Failed to set secret at <%s>", keyPath)
			return nil, err
		}
		log.Debugf("Re-encrypted %s", keyPath)
	}

	return
}

// Set into `keyPath` the encrypted value for `plainText`
func (cfg *ConfigFile) Set(keyPath string, plainText []byte) (err error) {
	log.Debugf("Setting secret value for %s", keyPath)

	if !cfg.HasCrypto() {
		return errors.New("Cannot encrypt, provider is not enabled for encryption. See logs")
	}
	var data interface{}
	data, err = encryptCipherText(plainText, cfg.crypto)
	if err != nil {
		return
	}

	return cfg.data.Set(keyPath, data)
}

// VeryInsecurelySetPlaintext very insecurely sets `plainText`, without encrypting, at `keyPath`
func (cfg *ConfigFile) VeryInsecurelySetPlaintext(keyPath string, plainText []byte) error {

	var data interface{}
	err := json.Unmarshal(plainText, &data)
	if err != nil {
		data = string(plainText)
	} else {
		log.Debug("Interpreting value as JSON data")
	}

	log.Debugf("Insecurely Setting %s to %v", keyPath, data)

	return cfg.data.Set(keyPath, data)
}

// ListSecrets returns a slice of all the encrypted keyPaths in this config file
func (cfg *ConfigFile) ListSecrets() []string {
	return secretsForNode(cfg.data, "")
}

// Serialize the config into YAML
func (cfg *ConfigFile) Serialize() ([]byte, error) {
	return cfg.data.Serialize()
}
