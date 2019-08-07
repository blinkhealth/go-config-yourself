// Package input provides helpers to work with user inputs
package input

// Copyright 2018 Blink Health LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

// ReadFile returns a file as bytes
func ReadFile(file string) (plainText []byte, err error) {
	fileBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return fileBytes, err
	}

	return checkInputSize(fileBytes, err)
}

// ReadSecret returns the contents of Stdin as bytes, masking input optionally if
// reading from a TTY
func ReadSecret(prompt string, maskInput bool) (plainText []byte, err error) {
	if terminal.IsTerminal(int(os.Stdin.Fd())) {
		// we have an interactive session
		log.Debug("Prompting for input")
		var result string
		prompt := promptui.Prompt{
			Label: prompt,
			Templates: &promptui.PromptTemplates{
				Success: "Read secret ",
			},
		}

		if maskInput {
			prompt.Mask = '*'
		}

		result, err = prompt.Run()
		plainText = []byte(result)
	} else {
		// Not a tty, read from stdin until EOF
		log.Debug("Reading from stdin")
		plainText = readStdin()
	}

	if err != nil && err != io.EOF {
		return
	}

	return checkInputSize(plainText, nil)
}

// SelectionFromList returns a number of values from `list`
func SelectionFromList(list []string, prompt string, takeMultiple bool) (output []string, err error) {

	if len(list) == 0 {
		return nil, errors.New("Unable to make a selection, no items found")
	}

	searcher := func(input string, index int) bool {
		item := list[index]
		name := strings.ToLower(item)
		input = strings.ToLower(input)

		return strings.Contains(name, input)
	}

	ui := &promptui.Select{
		Label:    prompt,
		Items:    list,
		Size:     10,
		Searcher: searcher,
		Templates: &promptui.SelectTemplates{
			Help: "Move: ← ↓ ↑ →, search: /",
		},
	}

	selected, err := runUI(ui)

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	output = append(output, selected)

	if takeMultiple {
		items := ui.Items.([]string)
		ui.Items = append([]string{"Done"}, items...)
		ui.Label = ui.Label.(string) + ", or <Done> to continue"
		for {
			selected, err := runUI(ui)
			if err == promptui.ErrInterrupt {
				return []string{}, fmt.Errorf("Cancelled selection")
			}

			if err == promptui.ErrEOF {
				break
			}
			output = append(output, selected)
		}
	}

	return

}

func runUI(ui *promptui.Select) (out string, err error) {
	i, selected, err := ui.Run()

	if err != nil {
		return
	}

	if i == 0 && selected == "Done" {
		return "", promptui.ErrEOF
	}

	items := ui.Items.([]string)
	if i == len(out)-1 {
		ui.Items = items[:i]
	} else {
		newItems := items[:i]
		newItems = append(newItems, items[i+1:]...)
		ui.Items = newItems
	}

	return selected, nil
}
