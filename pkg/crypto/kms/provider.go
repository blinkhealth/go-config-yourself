// Copyright 2018 Blink Health LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0

// Package kms adds kms support for go-config-yourself
//
// It uses the AWS KMS (https://aws.amazon.com/kms/) service to encrypt every secret value.
package kms

import (
	"fmt"
	"strings"

	pvd "github.com/blinkhealth/go-config-yourself/pkg/provider"

	"github.com/blinkhealth/go-config-yourself/internal/input"
	log "github.com/sirupsen/logrus"
)

func init() {
	pvd.RegisterProvider("kms", New, []pvd.Argument{
		{
			Name:        "key",
			Description: "The kms key ARN to use",
		},
	})
}

// Provider implements provider.Crypto for KMS
type Provider struct {
	key     string
	service *kmsService
}

// New creates a new kms.Provider and returns it
func New(config map[string]interface{}) (pvd.Crypto, error) {
	key, _ := config["key"].(string)

	var region string
	if key != "" {
		if err := validKey(key); err != nil {
			return nil, err
		}
		// arn:aws:kms:REGION:ACCOUNT:ID
		pieces := strings.Split(key, ":")
		log.Debugf("Inferring region from key %s", key)
		region = pieces[3]
	} else {
		// Initializing a file with no key provided, set a temporary one to be
		// replaced after the user selects a key from a list
		key = "initialization-temporary-key"
	}

	kmsSvc := newKMSService(region)
	log.Debugf("Initializing secure config with key %s in region %s", key, region)

	return &Provider{
		key:     key,
		service: kmsSvc,
	}, nil
}

// Enabled tells whether the provider is ready to operate on secrets
func (provider *Provider) Enabled() bool {
	return provider.key != ""
}

// Encrypt bytes
func (provider *Provider) Encrypt(plainText []byte) ([]byte, error) {
	return provider.service.Encrypt(provider.key, plainText)
}

// Decrypt bytes
func (provider *Provider) Decrypt(encryptedBytes []byte) (string, error) {
	return provider.service.Decrypt(encryptedBytes)
}

// Replace the key with a new one
//
// Will query every available AWS region and then prompt the user to select a key from it, unless `key` is present in `args`
// Unless `AWS_AP_EAST_1_ENABLED` is set on the environment, the `ap-east-1` region will be ignored when listing keys. This region is not enabled by default (https://docs.aws.amazon.com/general/latest/gr/rande.html).
func (provider *Provider) Replace(args map[string]interface{}) (err error) {
	var key string

	if value, exists := args["key"]; exists {
		if keyString, isString := value.(string); isString {
			if err = validKey(keyString); err != nil {
				return
			}
			key = keyString
		}
	}

	if key == "" {
		keys, err := provider.service.ListKeys()
		if err != nil {
			return fmt.Errorf("Failed to list keys: %s", err)
		}

		responses, err := input.SelectionFromList(keys, "kms", false)
		if err != nil {
			return err
		}
		key = responses[0]
	}

	provider.key = key
	pieces := strings.Split(key, ":")
	region := pieces[3]
	kmsSvc := newKMSService(region)
	provider.service = kmsSvc

	return
}

// Serialize into a map of config for later hydration
func (provider *Provider) Serialize() (serialized map[string]interface{}) {
	serialized = make(map[string]interface{})
	serialized["key"] = provider.key
	serialized["provider"] = "kms"
	return
}

func validKey(key string) (err error) {
	if !strings.Contains(key, "arn:aws:kms:") {
		err = fmt.Errorf("Unable to infer region from non fully-qualified KMS key ARN <%s>", key)
	}
	return
}
