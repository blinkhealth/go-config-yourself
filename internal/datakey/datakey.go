// Package datakey provides a AES GCM service
package datakey

// Copyright 2018 Blink Health LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// KeySize defines the key length when creating new keys
const KeySize = 32

// NonceSize defines the nonce length when creating new keys
const NonceSize = 32

// Service operates on bytes with a data key
type Service struct {
	key []byte
}

// Fills a slice with random bytes
func RandomBytes(value *[]byte) error {
	_, err := rand.Read(*value)
	if err != nil {
		return fmt.Errorf("Could not read random bytes for data key: %s", err)
	}
	return nil
}

// New returns a new data key
func New() ([]byte, error) {
	key := make([]byte, KeySize)

	return key, RandomBytes(&key)
}

// NewService returns an AES service for a given key
func NewService(key []byte) *Service {
	return &Service{key}
}

// Decrypt returns a plaintext string
func (svc *Service) Decrypt(encryptedBytes []byte) (plainText string, err error) {
	var plainBytes []byte
	plainBytes, err = svc.DecryptAsBytes(encryptedBytes)
	return string(plainBytes), err
}

// DecryptAsBytes returns decrypted bytes
func (svc *Service) DecryptAsBytes(encryptedBytes []byte) (plainBytes []byte, err error) {
	var aes cipher.AEAD
	aes, err = newAes(svc.key)
	if err != nil {
		return
	}

	return aes.Open(nil, encryptedBytes[:NonceSize], encryptedBytes[NonceSize:], nil)
}

// Encrypt plaintext bytes into ciphertext
func (svc *Service) Encrypt(plainText []byte) (cipherText []byte, err error) {
	var aes cipher.AEAD
	aes, err = newAes(svc.key)
	if err != nil {
		return
	}
	nonce := make([]byte, NonceSize)

	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}
	cipherText = nonce
	cipherText = append(cipherText, aes.Seal(nil, nonce, plainText, nil)...)

	return
}

// newAes returns a cipher object and its nonce size ready for operations
func newAes(key []byte) (svc cipher.AEAD, err error) {
	var block cipher.Block
	block, err = aes.NewCipher(key)
	if err != nil {
		return
	}

	return cipher.NewGCMWithNonceSize(block, NonceSize)
}
