// Copyright 2018 Blink Health LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0

// Package provider represents an abstract provider. Providers must implement the Crypto interface, and call RegisterProvider during their init method
//
// Provider packages must also be imported by pkg/file
package provider

// Argument represents values required for providers to initialize or rekey
type Argument struct {
	Name        string
	Description string
	Default     string
	EnvVarName  string
	Repeatable  bool
	IsSwitch    bool
}

// Crypto is what a provider implements to operate on secrets
type Crypto interface {
	// Enabled represents tells this provider is ready to operate on secrets or not
	Enabled() bool
	// Replace takes a map of arguments and reinitializes the provider to encrypt with new crypto values
	Replace(map[string]interface{}) error
	// Serialize must return a map with `provider: crypto` and any values that the provider needs to initialize itself via `New`
	Serialize() map[string]interface{}
	// Encrypt takes a byte slice and returns it encrypted
	Encrypt([]byte) ([]byte, error)
	// Decrypt takes a byte slice and returns plaintext for it
	Decrypt([]byte) (string, error)
}

// Constructor is the signature of the function to initialize providers
type Constructor = func(map[string]interface{}) (Crypto, error)

// Registration describes an available provider
type Registration struct {
	// New is the constructor function for this provider
	New Constructor
	// Flags are a list of arguments consumed by this provider
	Flags []Argument
}

// ProviderList enumerates the available providers
var ProviderList = []string{}

// Providers holds providers Registrations
var Providers = map[string]Registration{}

// RegisterProvider is what a provider calls in their `init` func to expose them to the config system
func RegisterProvider(name string, constructor Constructor, flags []Argument) {
	ProviderList = append(ProviderList, name)
	Providers[name] = Registration{
		New:   constructor,
		Flags: flags,
	}
}

// AvailableFlags enumerates all available provider flags
func AvailableFlags() (flags []Argument) {
	for _, pvd := range Providers {
		flags = append(flags, pvd.Flags...)
	}
	return
}
