// Package go-config-yourself offers a command line interface to manipulate yaml files with encrypted values
package main

// Copyright 2018 Blink Health LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0

import (
	"github.com/blinkhealth/go-config-yourself/cmd"
)

// version gets overridden at compile time
var version = "alpha"

func main() {
	cmd.Main(version)
}
