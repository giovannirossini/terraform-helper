package main

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

// promptUserSelection displays an interactive selection menu for the user
func promptUserSelection(matches []string) (string, error) {
	if len(matches) == 0 {
		return "", fmt.Errorf("no matches to select from")
	}

	prompt := promptui.Select{
		Label: "Multiple resources found, please select one",
		Items: matches,
		Size:  10,
	}

	_, result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}
