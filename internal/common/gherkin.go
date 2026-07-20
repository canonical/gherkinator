package common

import (
	"strings"

	gherkin "github.com/cucumber/gherkin/go/v28"
	messages "github.com/cucumber/messages/go/v24"
)

// ValidateGherkin parses a Gherkin string with the official Cucumber parser
// and returns an error if the syntax is invalid.
func ValidateGherkin(gherkinText string) error {
	reader := strings.NewReader(gherkinText)
	_, err := gherkin.ParseGherkinDocument(reader, (&messages.Incrementing{}).NewId)
	return err
}
