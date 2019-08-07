// Copyright 2018 Blink Health LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0

// Package gpg adds gpg support for go-config-yourself
package gpg

import (
	"fmt"

	"github.com/blinkhealth/go-config-yourself/internal/input"
	pvd "github.com/blinkhealth/go-config-yourself/pkg/provider"
	log "github.com/sirupsen/logrus"
)

func init() {
	pvd.RegisterProvider("gpg", New, []pvd.Argument{
		{
			Name:        "public-key",
			Description: "The gpg public key indentity to use (fingerprint or email), can be specified multiple times.",
			Repeatable:  true,
		},
	})
}

// Provider implements provider.Crypto for GPG
type Provider struct {
	service *gpgService
}

// New creates a new gpg.Provider and returns it
func New(config map[string]interface{}) (pvd.Crypto, error) {
	var service *gpgService
	// we might be creating this provider from an existing file with a key
	configKey, isString := config["key"].(string)
	configRecipientsIface, isList := config["recipients"].([]interface{})
	if isString && isList {
		log.Debug("Creating gpg service from key")

		var configRecipients []string
		for _, r := range configRecipientsIface {
			configRecipients = append(configRecipients, r.(string))
		}
		service = gpgServiceFromConfig(configKey, configRecipients)
	}

	if flags, hasFlags := config["flags"]; hasFlags {
		if recipients, hasRecipients := flags.(map[string]interface{})["public-key"]; hasRecipients {
			var err error
			keys, isList := recipients.([]string)
			if !isList {
				return nil, fmt.Errorf("Unable to parse public keys as list")
			}

			log.Debug("Creating gpg service")
			service, err = newGPGService(keys)
			if err != nil {
				return nil, err
			}
		}
	}

	return &Provider{service: service}, nil
}

// Enabled tells whether the provider is ready to operate on secrets
func (provider *Provider) Enabled() bool {
	return provider.service != nil
}

// Encrypt bytes
func (provider *Provider) Encrypt(plainText []byte) (cipherText []byte, err error) {
	if err = provider.readyForCrypto(); err == nil {
		cipherText, err = provider.service.Encrypt(plainText)
	}
	return
}

// Decrypt bytes
func (provider *Provider) Decrypt(cipherText []byte) (plainText string, err error) {
	if err = provider.readyForCrypto(); err == nil {
		plainText, err = provider.service.Decrypt(cipherText)
	}
	return
}

// Replace the key with a new one
//
// Will query the GPG agent for keys, and prompt the user to select one or more keys from it, unless `public-key` is present in `args`
func (provider *Provider) Replace(args map[string]interface{}) (err error) {
	var service *gpgService
	var keys []string

	recipients, hasRecipients := args["recipients"]
	if !hasRecipients {
		recipients, hasRecipients = args["public-key"]
	}

	if hasRecipients {
		var isList bool

		keys, isList = recipients.([]string)
		if !isList {
			return fmt.Errorf("Unable to parse public keys as list")
		}

	} else {
		log.Debugf("No GPG recipients specified, querying agent for keys")
		allKeys, err := provider.service.ListKeys()
		if err != nil {
			return err
		}

		keys, err = input.SelectionFromList(allKeys, "Select a gpg identity", true)
		if err != nil {
			return err
		}
	}

	log.Debugf("Creating gpg service for %s", keys)
	service, err = newGPGService(keys)
	if err != nil {
		return
	}
	provider.service = service

	return
}

// Serialize into a map of config for later hydration
func (provider *Provider) Serialize() (serialized map[string]interface{}) {
	serialized = make(map[string]interface{})
	serialized["provider"] = "gpg"

	if provider.service != nil {
		provider.service.Serialize(serialized)
	}
	return
}

func (provider *Provider) readyForCrypto() (err error) {
	if provider.service.IsAvailable() {
		// the gpgService has a decrypted key, continue
		return
	}

	return provider.service.DecryptKey()
}
