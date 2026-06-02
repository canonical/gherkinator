package main

import (
	"fmt"
	"regexp"
	"strings"
)

// GenerateGherkin transpiles a TestPlan struct into a valid Gherkin string.
func GenerateGherkin(plan TestPlan) string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "Feature: %s\n", plan.Feature)
	if plan.Description != nil {
		fmt.Fprintf(&builder, "  %s\n", *plan.Description)
	}
	if plan.Background != nil {
		builder.WriteString("\n  Background:\n")
		for _, line := range strings.Split(strings.TrimSpace(*plan.Background), "\n") {
			fmt.Fprintf(&builder, "    %s\n", strings.TrimSpace(line))
		}
	}

	fmt.Fprintf(&builder, "\n  @%s @%s\n", plan.Type, plan.Risk)

	hasExamples := len(plan.Examples) > 0
	for _, scenario := range plan.Scenarios {
		parts := strings.SplitN(scenario, "\n", 2)
		title := strings.TrimSpace(parts[0])
		steps := ""
		if len(parts) > 1 {
			steps = parts[1]
		}

		if hasExamples {
			fmt.Fprintf(&builder, "  Scenario Outline: %s\n", title)
		} else {
			fmt.Fprintf(&builder, "  Scenario: %s\n", title)
		}
		if steps != "" {
			for _, step := range strings.Split(strings.TrimSpace(steps), "\n") {
				fmt.Fprintf(&builder, "    %s\n", strings.TrimSpace(step))
			}
		}
	}

	if hasExamples {
		builder.WriteString("\n  Examples:\n")
		paramRegex := regexp.MustCompile(`<([^>]+)>`)
		var params []string
		seen := make(map[string]bool)
		for _, scenario := range plan.Scenarios {
			for _, match := range paramRegex.FindAllStringSubmatch(scenario, -1) {
				if !seen[match[1]] {
					seen[match[1]] = true
					params = append(params, match[1])
				}
			}
		}
		builder.WriteString("    | " + strings.Join(params, " | ") + " |\n")
		for _, row := range plan.Examples {
			builder.WriteString("    | " + strings.Join(row, " | ") + " |\n")
		}
	}
	return builder.String()
}

// GenerateMarkdown transpiles a TestPlan struct into a Markdown string.
func GenerateMarkdown(plan TestPlan) string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "# %s\n\n", plan.Feature)
	fmt.Fprintf(&builder, "- **Type:** %s\n", plan.Type)
	fmt.Fprintf(&builder, "- **Status:** %s\n", plan.Status)
	fmt.Fprintf(&builder, "- **Risk:** %s\n", plan.Risk)
	if plan.Issues != nil {
		fmt.Fprintf(&builder, "- **Issues:** %s\n", *plan.Issues)
	}
	if plan.Docs != nil {
		fmt.Fprintf(&builder, "- **Docs:** %s\n", *plan.Docs)
	}
	builder.WriteString("\n")

	if plan.Description != nil {
		builder.WriteString("## Description\n\n")
		fmt.Fprintf(&builder, "%s\n\n", *plan.Description)
	}

	if plan.Background != nil {
		builder.WriteString("## Background\n\n")
		for _, line := range strings.Split(strings.TrimSpace(*plan.Background), "\n") {
			fmt.Fprintf(&builder, "- %s\n", strings.TrimSpace(line))
		}
		builder.WriteString("\n")
	}

	hasExamples := len(plan.Examples) > 0
	if hasExamples {
		builder.WriteString("## Scenario Outlines\n\n")
	} else {
		builder.WriteString("## Scenarios\n\n")
	}
	for _, scenario := range plan.Scenarios {
		parts := strings.SplitN(scenario, "\n", 2)
		fmt.Fprintf(&builder, "### %s\n\n", strings.TrimSpace(parts[0]))
		if len(parts) > 1 {
			for _, step := range strings.Split(strings.TrimSpace(parts[1]), "\n") {
				step = strings.Replace(step, "Given ", "**Given** ", 1)
				step = strings.Replace(step, "When ", "**When** ", 1)
				step = strings.Replace(step, "Then ", "**Then** ", 1)
				step = strings.Replace(step, "And ", "**And** ", 1)
				fmt.Fprintf(&builder, "- %s\n", strings.TrimSpace(step))
			}
		}
		builder.WriteString("\n")
	}

	if len(plan.Examples) > 0 {
		builder.WriteString("## Examples\n\n")
		// Extract parameter names from scenarios
		paramRegex := regexp.MustCompile(`<([^>]+)>`)
		var params []string
		seen := make(map[string]bool)
		for _, scenario := range plan.Scenarios {
			for _, match := range paramRegex.FindAllStringSubmatch(scenario, -1) {
				if !seen[match[1]] {
					seen[match[1]] = true
					params = append(params, match[1])
				}
			}
		}
		if len(params) > 0 {
			builder.WriteString("| " + strings.Join(params, " | ") + " |\n")
			builder.WriteString("| " + strings.Repeat("--- | ", len(params)) + "\n")
		}
		for _, row := range plan.Examples {
			builder.WriteString("| " + strings.Join(row, " | ") + " |\n")
		}
		builder.WriteString("\n")
	}

	return builder.String()
}
