package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type addModel struct {
	state           int // 0: Feature, 1: Type, 2: Status, 3: Description, 4: Background, 5: Scenarios, 6: Examples
	feature         textinput.Model
	description     textinput.Model
	background      textarea.Model
	scenarios       textarea.Model
	examples        textarea.Model
	testType        string
	status          string
	parsedScenarios []string
	parsedExamples  [][]string
	choices         []string
	cursor          int
	canceled        bool
	done            bool
}

func initialModel() addModel {
	fi := textinput.New()
	fi.Placeholder = "Enter feature title"
	fi.Focus()

	di := textinput.New()
	di.Placeholder = "Enter description (optional)"

	bi := textarea.New()
	bi.Placeholder = "Enter background / test environment setup (optional)"
	bi.MaxHeight = 6
	bi.ShowLineNumbers = false

	scenariosInput := textarea.New()
	scenariosInput.Placeholder = "Enter Scenarios (separate each scenario with a blank line)"
	scenariosInput.MaxHeight = 8
	scenariosInput.ShowLineNumbers = false

	examplesInput := textarea.New()
	examplesInput.Placeholder = "Enter Examples as CSV (header row first, then data rows)"
	examplesInput.MaxHeight = 8
	examplesInput.ShowLineNumbers = false

	return addModel{
		state:           0,
		feature:         fi,
		description:     di,
		background:      bi,
		scenarios:       scenariosInput,
		examples:        examplesInput,
		parsedScenarios: []string{},
		parsedExamples:  [][]string{},
	}
}

func (m addModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m addModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true
			return m, tea.Quit
		case tea.KeyCtrlD:
			// Ctrl+D to finish textarea states
			switch m.state {
			case 4:
				m.state++
				m.scenarios.Focus()
				return m, textarea.Blink
			case 5:
				m.parsedScenarios = parseScenarios(m.scenarios.Value())
				m.state++
				m.examples.Focus()
				return m, textarea.Blink
			case 6:
				m.parsedExamples = parseExamples(m.examples.Value())
				m.done = true
				return m, tea.Quit
			}
		case tea.KeyEnter:
			switch {
			case m.state == 0 && m.feature.Value() != "":
				m.state++
				m.choices = []string{"functional", "solution", "performance", "reliability", "security"}
				m.cursor = 0
				return m, nil
			case m.state == 1:
				m.testType = m.choices[m.cursor]
				m.state++
				m.choices = []string{"planned", "implemented", "deprecated"}
				m.cursor = 0
				return m, nil
			case m.state == 2:
				m.status = m.choices[m.cursor]
				m.state++
				m.description.Focus()
				return m, textinput.Blink
			case m.state == 3:
				m.state++
				m.background.Focus()
				return m, textarea.Blink
			}
		case tea.KeyUp:
			if (m.state == 1 || m.state == 2) && m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if (m.state == 1 || m.state == 2) && m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		}
	}

	switch m.state {
	case 0:
		m.feature, cmd = m.feature.Update(msg)
	case 3:
		m.description, cmd = m.description.Update(msg)
	case 4:
		m.background, cmd = m.background.Update(msg)
	case 5:
		m.scenarios, cmd = m.scenarios.Update(msg)
	case 6:
		m.examples, cmd = m.examples.Update(msg)
	}
	return m, cmd
}

func (m addModel) View() string {
	if m.canceled {
		return "Operation canceled.\n"
	}
	if m.done {
		return "Saving test plan...\n"
	}

	s := "\n"
	switch m.state {
	case 0:
		s += "Feature Title:\n" + m.feature.View()
	case 1, 2:
		prompt := "Select Test Type:\n"
		if m.state == 2 {
			prompt = "Select Implementation Status:\n"
		}
		s += prompt
		for i, choice := range m.choices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
	case 3:
		s += "Description (optional):\n" + m.description.View()
	case 4:
		s += "Background (optional) - Press Ctrl+D to finish:\n" + m.background.View()
	case 5:
		s += "Scenarios (Separate each scenario with a blank line) - Press Ctrl+D to finish:\n" + m.scenarios.View()
	case 6:
		s += "Examples (CSV format: header row first, then data rows) - Press Ctrl+D to finish:\n" + m.examples.View()
	}

	hint := "(Press Esc to quit)"
	if m.state >= 4 {
		hint += " | Ctrl+D to finish"
	}
	return s + "\n\n" + hint + "\n"
}

// parseScenarios splits multi-line text into individual scenarios
// separated by blank lines (double newlines).
func parseScenarios(text string) []string {
	if text == "" {
		return []string{}
	}
	blocks := strings.Split(text, "\n\n")
	var scenarios []string
	for _, block := range blocks {
		trimmed := strings.TrimSpace(block)
		if trimmed != "" {
			scenarios = append(scenarios, trimmed)
		}
	}
	return scenarios
}

// parseExamples parses CSV-formatted text into a slice of string slices.
// The first line defines headers, subsequent lines define data rows.
func parseExamples(text string) [][]string {
	if text == "" {
		return [][]string{}
	}
	lines := strings.Split(text, "\n")
	var result [][]string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		for i, field := range fields {
			fields[i] = strings.TrimSpace(field)
		}
		result = append(result, fields)
	}
	return result
}
