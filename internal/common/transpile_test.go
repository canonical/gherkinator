package common

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func strPtr(s string) *string {
	return &s
}

func TestGenerateGherkin_BasicScenario(t *testing.T) {
	plan := TestPlan{
		Feature:   "Login Feature",
		Type:      "functional",
		Status:    "planned",
		Risk:      "stable",
		Scenarios: []string{"User logs in\nGiven a user exists\nWhen the user logs in\nThen the user sees the dashboard"},
	}
	result := GenerateGherkin(plan)
	assert.Contains(t, result, "Feature: Login Feature")
	assert.Contains(t, result, "@functional")
	assert.Contains(t, result, "@stable")
	assert.Contains(t, result, "Scenario: User logs in")
	assert.Contains(t, result, "Given a user exists")
	assert.Contains(t, result, "When the user logs in")
	assert.Contains(t, result, "Then the user sees the dashboard")
}

func TestGenerateGherkin_WithDescription(t *testing.T) {
	plan := TestPlan{
		Feature:     "Login Feature",
		Type:        "functional",
		Status:      "planned",
		Risk:        "stable",
		Description: strPtr("This is a description"),
		Scenarios:   []string{"A scenario\nGiven something"},
	}
	result := GenerateGherkin(plan)
	assert.Contains(t, result, "This is a description")
}

func TestGenerateGherkin_WithBackground(t *testing.T) {
	plan := TestPlan{
		Feature:    "Login Feature",
		Type:       "functional",
		Status:     "planned",
		Risk:       "stable",
		Background: strPtr("Given the system is running"),
		Scenarios:  []string{"A scenario\nGiven something"},
	}
	result := GenerateGherkin(plan)
	assert.Contains(t, result, "Background:")
	assert.Contains(t, result, "Given the system is running")
}

func TestGenerateGherkin_WithExamples(t *testing.T) {
	plan := TestPlan{
		Feature:   "Parameterized Feature",
		Type:      "functional",
		Status:    "planned",
		Risk:      "stable",
		Scenarios: []string{"User logs in with <username>\nGiven a user <username>\nWhen the user enters <password>"},
		Examples:  [][]string{{"alice", "pass123"}, {"bob", "pass456"}},
	}
	result := GenerateGherkin(plan)
	assert.Contains(t, result, "Scenario Outline: User logs in with <username>")
	assert.Contains(t, result, "Examples:")
	assert.Contains(t, result, "| username | password |")
	assert.Contains(t, result, "| alice | pass123 |")
	assert.Contains(t, result, "| bob | pass456 |")
}

func TestGenerateGherkin_ScenarioWithoutSteps(t *testing.T) {
	plan := TestPlan{
		Feature:   "Minimal Feature",
		Type:      "functional",
		Status:    "planned",
		Risk:      "stable",
		Scenarios: []string{"Just a title"},
	}
	result := GenerateGherkin(plan)
	assert.Contains(t, result, "Scenario: Just a title")
}

func TestGenerateGherkin_NilDescription(t *testing.T) {
	plan := TestPlan{
		Feature:   "No Desc",
		Type:      "functional",
		Status:    "planned",
		Risk:      "stable",
		Scenarios: []string{"Scenario title\nGiven x"},
	}
	result := GenerateGherkin(plan)
	assert.NotContains(t, result, "Description")
}

func TestGenerateGherkin_NilBackground(t *testing.T) {
	plan := TestPlan{
		Feature:   "No Background",
		Type:      "functional",
		Status:    "planned",
		Risk:      "stable",
		Scenarios: []string{"Scenario title\nGiven x"},
	}
	result := GenerateGherkin(plan)
	assert.NotContains(t, result, "Background:")
}

func TestGenerateMarkdown_BasicPlan(t *testing.T) {
	plan := TestPlan{
		Feature:   "Login Feature",
		Type:      "functional",
		Status:    "planned",
		Risk:      "stable",
		Scenarios: []string{"User logs in\nGiven a user exists\nWhen the user logs in\nThen the user sees the dashboard"},
	}
	result := GenerateMarkdown(plan)
	assert.Contains(t, result, "# Login Feature")
	assert.Contains(t, result, "- **Type:** functional")
	assert.Contains(t, result, "- **Status:** planned")
	assert.Contains(t, result, "- **Risk:** stable")
	assert.Contains(t, result, "## Scenarios")
	assert.Contains(t, result, "### User logs in")
	assert.Contains(t, result, "**Given** a user exists")
	assert.Contains(t, result, "**When** the user logs in")
	assert.Contains(t, result, "**Then** the user sees the dashboard")
}

func TestGenerateMarkdown_WithDescription(t *testing.T) {
	plan := TestPlan{
		Feature:     "A Feature",
		Type:        "solution",
		Status:      "implemented",
		Risk:        "stable",
		Description: strPtr("Some description text"),
		Scenarios:   []string{"A scenario"},
	}
	result := GenerateMarkdown(plan)
	assert.Contains(t, result, "## Description")
	assert.Contains(t, result, "Some description text")
}

func TestGenerateMarkdown_WithBackground(t *testing.T) {
	plan := TestPlan{
		Feature:    "A Feature",
		Type:       "solution",
		Status:     "implemented",
		Risk:       "stable",
		Background: strPtr("Given precondition A\nGiven precondition B"),
		Scenarios:  []string{"A scenario"},
	}
	result := GenerateMarkdown(plan)
	assert.Contains(t, result, "## Background")
	assert.Contains(t, result, "- Given precondition A")
	assert.Contains(t, result, "- Given precondition B")
}

func TestGenerateMarkdown_WithIssuesAndDocs(t *testing.T) {
	plan := TestPlan{
		Feature:   "A Feature",
		Type:      "security",
		Status:    "deprecated",
		Risk:      "stable",
		Issues:    strPtr("https://github.com/issues/1"),
		Docs:      strPtr("https://docs.example.com"),
		Scenarios: []string{"A scenario"},
	}
	result := GenerateMarkdown(plan)
	assert.Contains(t, result, "- **Issues:** https://github.com/issues/1")
	assert.Contains(t, result, "- **Docs:** https://docs.example.com")
}

func TestGenerateMarkdown_NilOptionalFields(t *testing.T) {
	plan := TestPlan{
		Feature:   "Simple",
		Type:      "functional",
		Status:    "planned",
		Risk:      "stable",
		Scenarios: []string{"Test scenario"},
	}
	result := GenerateMarkdown(plan)
	assert.NotContains(t, result, "## Description")
	assert.NotContains(t, result, "## Background")
	assert.NotContains(t, result, "**Issues:**")
	assert.NotContains(t, result, "**Docs:**")
}

func TestGenerateMarkdown_AndStep(t *testing.T) {
	plan := TestPlan{
		Feature:   "And Steps",
		Type:      "functional",
		Status:    "planned",
		Risk:      "stable",
		Scenarios: []string{"And scenario\nGiven X\nAnd Y\nWhen Z\nThen W"},
	}
	result := GenerateMarkdown(plan)
	assert.Contains(t, result, "**And** Y")
}

func TestGenerateMarkdown_ScenarioWithoutSteps(t *testing.T) {
	plan := TestPlan{
		Feature:   "Minimal",
		Type:      "functional",
		Status:    "planned",
		Risk:      "stable",
		Scenarios: []string{"Just a title"},
	}
	result := GenerateMarkdown(plan)
	assert.Contains(t, result, "### Just a title")
	// Should not have step lines
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "- **Given**") ||
			strings.HasPrefix(strings.TrimSpace(line), "- **When**") ||
			strings.HasPrefix(strings.TrimSpace(line), "- **Then**") {
			t.Fatal("Should not have step lines for scenario without steps")
		}
	}
}

func TestGenerateGherkin_MultipleScenarios(t *testing.T) {
	plan := TestPlan{
		Feature: "Multi Scenario",
		Type:    "functional",
		Status:  "planned",
		Risk:    "stable",
		Scenarios: []string{
			"First scenario\nGiven A\nWhen B\nThen C",
			"Second scenario\nGiven D\nWhen E\nThen F",
		},
	}
	result := GenerateGherkin(plan)
	assert.Contains(t, result, "Scenario: First scenario")
	assert.Contains(t, result, "Scenario: Second scenario")
}

func TestGenerateMarkdown_MultipleScenarios(t *testing.T) {
	plan := TestPlan{
		Feature: "Multi Scenario",
		Type:    "functional",
		Status:  "planned",
		Risk:    "stable",
		Scenarios: []string{
			"First scenario\nGiven A\nWhen B\nThen C",
			"Second scenario\nGiven D\nWhen E\nThen F",
		},
	}
	result := GenerateMarkdown(plan)
	assert.Contains(t, result, "### First scenario")
	assert.Contains(t, result, "### Second scenario")
}

func TestGenerateMarkdown_WithExamples(t *testing.T) {
	plan := TestPlan{
		Feature:   "Parameterized Feature",
		Type:      "functional",
		Status:    "planned",
		Risk:      "stable",
		Scenarios: []string{"User logs in with <username>\nGiven a user <username>\nWhen the user enters <password>"},
		Examples:  [][]string{{"alice", "pass123"}, {"bob", "pass456"}},
	}
	result := GenerateMarkdown(plan)
	assert.Contains(t, result, "## Scenario Outlines")
	assert.NotContains(t, result, "## Scenarios")
	assert.Contains(t, result, "## Examples")
	assert.Contains(t, result, "| username | password |")
	assert.Contains(t, result, "| --- | --- |")
	assert.Contains(t, result, "| alice | pass123 |")
	assert.Contains(t, result, "| bob | pass456 |")
}

func TestGenerateMarkdown_WithExamplesNoParams(t *testing.T) {
	plan := TestPlan{
		Feature:   "Feature",
		Type:      "functional",
		Status:    "planned",
		Risk:      "stable",
		Scenarios: []string{"A scenario\nGiven something"},
		Examples:  [][]string{{"val1", "val2"}},
	}
	result := GenerateMarkdown(plan)
	assert.Contains(t, result, "## Scenario Outlines")
	assert.NotContains(t, result, "## Scenarios")
	assert.Contains(t, result, "## Examples")
	assert.Contains(t, result, "| val1 | val2 |")
	// No header row since no <param> placeholders
	assert.NotContains(t, result, "| --- |")
}

func TestGenerateMarkdown_NoExamples(t *testing.T) {
	plan := TestPlan{
		Feature:   "No Examples",
		Type:      "functional",
		Status:    "planned",
		Risk:      "stable",
		Scenarios: []string{"A scenario\nGiven something"},
	}
	result := GenerateMarkdown(plan)
	assert.NotContains(t, result, "## Examples")
}
