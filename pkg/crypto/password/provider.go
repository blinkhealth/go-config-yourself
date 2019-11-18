// Copyright 2018 Blink Health LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0

// Package password adds password support for go-config-yourself
//
// It encrypts values with a key derived from a user-defined password. The values are encrypted using AES in GCM mode.
package password

import (
	"errors"
	"fmt"
	"os"

	"github.com/muesli/crunchy"
	log "github.com/sirupsen/logrus"

	"github.com/blinkhealth/go-config-yourself/internal/input"
	pvd "github.com/blinkhealth/go-config-yourself/pkg/provider"
)

// The minimum password length
var validationMinLength = 12

// A folder to look for dictionaries in
// nolint:gosec
var validationDictionaryFolder = "/usr/share/dict"

func init() {
	pvd.RegisterProvider("password", New, []pvd.Argument{
		{
			Name:        "password",
			Description: "A password to use for encryption and decryption",
			EnvVarName:  "CONFIG_PASSWORD",
		},
		{
			Name:        "skip-password-validation",
			IsSwitch:    true,
			Description: "Skips password validation, potentially making encrypted secrets easier to crack",
		},
	})
}

// Provider implements provider.Crypto for passwords
type Provider struct {
	service *passwordService
}

// New creates a new password.Provider and returns it
func New(config map[string]interface{}) (pvd.Crypto, error) {
	var service *passwordService

	// we might be creating this provider from an existing file with a key
	if stringKey, isString := config["key"].(string); isString {
		var err error
		log.Debug("Creating password service from key")
		service, err = passwordServiceFromKey(stringKey)
		if err != nil {
			return nil, err
		}
	}

	// if flags are available, user is providing password from the tty during rekey
	if flags, hasFlags := config["flags"]; hasFlags {
		if password, hasPassword := flags.(map[string]interface{})["password"]; hasPassword {
			var err error
			passwordString, isString := password.(string)
			if !isString {
				return nil, fmt.Errorf("Unable to parse password as string")
			}

			log.Debug("Creating password service from password")
			service, err = newPasswordService(passwordString)
			if err != nil {
				return nil, err
			}
		}
	}

	return &Provider{service: service}, nil
}

// Replace the current data key with a new one, encrypting it with a different password
//
// Will prompt for a `password` unless present in `args` or is set as `CONFIG_PASSWORD` in the environment
func (provider *Provider) Replace(args map[string]interface{}) (err error) {
	var password string
	if passwordKey, exists := args["password"]; exists {
		password, _ = passwordKey.(string)
	}

	if password == "" {
		password, err = getPassword("Enter the new password")
		if err != nil {
			return
		}
	}

	if args["skip-password-validation"] == true {
		log.Warn("Password complexity validation skipped!")
	} else {
		log.Debugf("Validating password complexity: min-length %d, dictionary: %s", validationMinLength, validationDictionaryFolder)
		err = validatePassword(password)
		if err != nil {
			return
		}
	}

	svc, err := newPasswordService(password)
	provider.service = svc
	return err
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
func (provider *Provider) Decrypt(data []byte) (plainText string, err error) {
	if err = provider.readyForCrypto(); err == nil {
		plainText, err = provider.service.Decrypt(data)
	}
	return
}

// Serialize into a map of config for later hydration
func (provider *Provider) Serialize() (serialized map[string]interface{}) {
	serialized = make(map[string]interface{})
	serialized["provider"] = "password"
	if provider.service != nil {
		// Serialize the service key
		serialized["key"] = provider.service.Serialize()
	}
	return
}

func (provider *Provider) readyForCrypto() (err error) {
	if provider.service.IsAvailable() {
		// the passwordService has a decrypted key, continue
		return
	}

	// get a password to decrypt the passwordService.key
	var password string
	password, err = getPassword("Please enter this file's password")
	if err != nil {
		return err
	}

	return provider.service.DecryptKey(password)
}

func getPassword(promptText string) (password string, err error) {
	password, passwordInEnv := os.LookupEnv("CONFIG_PASSWORD")
	if !passwordInEnv {
		var secretBytes []byte
		secretBytes, err = input.ReadSecret(promptText, true)
		if err != nil {
			if err.Error() == "Input was empty" {
				err = fmt.Errorf("No password supplied")
			}
			return "", err
		}

		password = string(secretBytes)
	}

	return password, nil
}

func validatePassword(password string) error {
	validator := crunchy.NewValidatorWithOpts(crunchy.Options{
		MinLength:      validationMinLength,
		DictionaryPath: validationDictionaryFolder,
	})

	if err := validator.Check(password); err != nil {
		switch err {
		case crunchy.ErrEmpty, crunchy.ErrTooShort:
			return errors.New("Chosen password is too short, please use at least 12 characters")
		case crunchy.ErrDictionary, crunchy.ErrTooSystematic, crunchy.ErrTooFewChars, crunchy.ErrMangledDictionary:
			return errors.New("Password seems easy to guess or has very low entropy")
		default:
			return err
		}
	}

	return nil
}
