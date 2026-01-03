package prompt

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

// Select displays an interactive selection menu for the user.
func Select(matches []string) (string, error) {
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
		return "", fmt.Errorf("prompt failed: %w", err)
	}

	return result, nil
}
