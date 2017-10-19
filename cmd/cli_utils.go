package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// RepeatUntilNo repeats an action until the user stops
func RepeatUntilNo(message string, action func() bool) {
	for {
		if !action() {
			break
		}
		if !AskForConfirmation(message) {
			break
		}
	}
}

// AskForConfirmation asks the user for confirmation. The user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user.
func AskForConfirmation(s string) bool {
	scanner := bufio.NewScanner(os.Stdin)
	msg := fmt.Sprintf("%s [y/n]?: ", s)
	for fmt.Print(msg); scanner.Scan(); fmt.Print(msg) {
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
	return false
}
