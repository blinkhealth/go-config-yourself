# go-config-yourself/pkg/file

[![](https://godoc.org/github.com/blinkhealth/go-config-yourself/pkg/file?status.svg)](https://godoc.org/github.com/blinkhealth/go-config-yourself/pkg/file)

A library to read `go-config-yourself` config files from your golang application.

## Usage

```go
package main

import "fmt"
import "github.com/blinhealth/go-config-yourself/pkg/file"

func main() {
	cfg, err := file.Load("./config/my-file.yml")
	if err != nil {
		panic(err)
	}

	plaintextValue, err := cfg.Get("path.to.secret")

	if err == nil {
		fmt.Println(fmt.Sprintf("The password is %s", plaintextValue))
		// Outputs: The password is hunter2
	}

	mapOfValues, err := cfg.GetAll()
	if err == nil {
		fmt.Println(fmt.Sprintf("The file as a map looks like: %v", mapOfValues))
		// Outputs: The file as a map looks like: map[string]...
	}
}
```
