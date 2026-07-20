package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateGherkin_ValidSyntax(t *testing.T) {
	gherkinText := `Feature: Login
  Scenario: User logs in
    Given a user exists
    When the user logs in
    Then the user sees the dashboard
`
	err := ValidateGherkin(gherkinText)
	assert.NoError(t, err)
}

func TestValidateGherkin_InvalidSyntax(t *testing.T) {
	gherkinText := `This is not valid gherkin at all`
	err := ValidateGherkin(gherkinText)
	assert.Error(t, err)
}
