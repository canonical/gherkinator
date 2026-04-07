package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestInitialModel(t *testing.T) {
	m := initialModel()
	assert.Equal(t, 0, m.state)
	assert.False(t, m.canceled)
	assert.False(t, m.done)
	assert.Equal(t, "", m.testType)
	assert.Equal(t, "", m.status)
	assert.Empty(t, m.parsedScenarios)
	assert.Empty(t, m.parsedExamples)
}

func TestAddModel_Init(t *testing.T) {
	m := initialModel()
	cmd := m.Init()
	assert.NotNil(t, cmd)
}

func TestAddModel_View_State0(t *testing.T) {
	m := initialModel()
	view := m.View()
	assert.Contains(t, view, "Feature Title:")
	assert.Contains(t, view, "(Press Esc to quit)")
}

func TestAddModel_View_Canceled(t *testing.T) {
	m := initialModel()
	m.canceled = true
	view := m.View()
	assert.Contains(t, view, "Operation canceled.")
}

func TestAddModel_View_Done(t *testing.T) {
	m := initialModel()
	m.done = true
	view := m.View()
	assert.Contains(t, view, "Saving test plan...")
}

func TestAddModel_View_State1(t *testing.T) {
	m := initialModel()
	m.state = 1
	m.choices = []string{"functional", "solution", "performance", "reliability", "security"}
	m.cursor = 0
	view := m.View()
	assert.Contains(t, view, "Select Test Type:")
	assert.Contains(t, view, "> functional")
}

func TestAddModel_View_State2(t *testing.T) {
	m := initialModel()
	m.state = 2
	m.choices = []string{"planned", "implemented", "deprecated"}
	m.cursor = 1
	view := m.View()
	assert.Contains(t, view, "Select Implementation Status:")
	assert.Contains(t, view, "> implemented")
}

func TestAddModel_View_State3(t *testing.T) {
	m := initialModel()
	m.state = 3
	view := m.View()
	assert.Contains(t, view, "Description (optional):")
}

func TestAddModel_View_State4(t *testing.T) {
	m := initialModel()
	m.state = 4
	view := m.View()
	assert.Contains(t, view, "Background (optional)")
	assert.Contains(t, view, "Ctrl+D to finish")
}

func TestAddModel_View_State5(t *testing.T) {
	m := initialModel()
	m.state = 5
	view := m.View()
	assert.Contains(t, view, "Scenarios")
	assert.Contains(t, view, "Ctrl+D to finish")
}

func TestAddModel_View_State6(t *testing.T) {
	m := initialModel()
	m.state = 6
	view := m.View()
	assert.Contains(t, view, "Examples")
	assert.Contains(t, view, "Ctrl+D to finish")
}

func TestAddModel_View_HintNoCtrlD_State0(t *testing.T) {
	m := initialModel()
	m.state = 0
	view := m.View()
	assert.NotContains(t, view, "Ctrl+D")
}

func TestAddModel_View_HintNoCtrlD_State3(t *testing.T) {
	m := initialModel()
	m.state = 3
	view := m.View()
	assert.NotContains(t, view, "Ctrl+D")
}

func TestAddModel_EscCancels(t *testing.T) {
	m := initialModel()
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	result := newModel.(addModel)
	assert.True(t, result.canceled)
	assert.NotNil(t, cmd)
}

func TestAddModel_CtrlCCancels(t *testing.T) {
	m := initialModel()
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	result := newModel.(addModel)
	assert.True(t, result.canceled)
	assert.NotNil(t, cmd)
}

func TestAddModel_EnterOnEmptyFeature(t *testing.T) {
	m := initialModel()
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := newModel.(addModel)
	assert.Equal(t, 0, result.state)
}

func TestAddModel_NavigateDown_State1(t *testing.T) {
	m := initialModel()
	m.state = 1
	m.choices = []string{"functional", "solution", "performance", "reliability", "security"}
	m.cursor = 0

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	result := newModel.(addModel)
	assert.Equal(t, 1, result.cursor)
}

func TestAddModel_NavigateUp_State1(t *testing.T) {
	m := initialModel()
	m.state = 1
	m.choices = []string{"functional", "solution", "performance", "reliability", "security"}
	m.cursor = 2

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	result := newModel.(addModel)
	assert.Equal(t, 1, result.cursor)
}

func TestAddModel_NavigateUp_AtZero(t *testing.T) {
	m := initialModel()
	m.state = 1
	m.choices = []string{"functional", "solution"}
	m.cursor = 0

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	result := newModel.(addModel)
	assert.Equal(t, 0, result.cursor)
}

func TestAddModel_NavigateDown_AtEnd(t *testing.T) {
	m := initialModel()
	m.state = 1
	m.choices = []string{"functional", "solution"}
	m.cursor = 1

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	result := newModel.(addModel)
	assert.Equal(t, 1, result.cursor)
}

func TestAddModel_NavigateDown_State2(t *testing.T) {
	m := initialModel()
	m.state = 2
	m.choices = []string{"planned", "implemented", "deprecated"}
	m.cursor = 0

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	result := newModel.(addModel)
	assert.Equal(t, 1, result.cursor)
}

func TestAddModel_NavigateUp_State2(t *testing.T) {
	m := initialModel()
	m.state = 2
	m.choices = []string{"planned", "implemented", "deprecated"}
	m.cursor = 2

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	result := newModel.(addModel)
	assert.Equal(t, 1, result.cursor)
}

func TestAddModel_SelectType_State1(t *testing.T) {
	m := initialModel()
	m.state = 1
	m.choices = []string{"functional", "solution", "performance", "reliability", "security"}
	m.cursor = 2

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := newModel.(addModel)
	assert.Equal(t, 2, result.state)
	assert.Equal(t, "performance", result.testType)
	assert.Equal(t, []string{"planned", "implemented", "deprecated"}, result.choices)
	assert.Equal(t, 0, result.cursor)
}

func TestAddModel_SelectStatus_State2(t *testing.T) {
	m := initialModel()
	m.state = 2
	m.choices = []string{"planned", "implemented", "deprecated"}
	m.cursor = 1

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := newModel.(addModel)
	assert.Equal(t, 3, result.state)
	assert.Equal(t, "implemented", result.status)
}

func TestAddModel_DescriptionToBackground_State3(t *testing.T) {
	m := initialModel()
	m.state = 3

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := newModel.(addModel)
	assert.Equal(t, 4, result.state)
	assert.False(t, result.done)
	assert.NotNil(t, cmd)
}

// Ctrl+D on state 4 (Background) advances to state 5 (Scenarios)
func TestAddModel_CtrlD_BackgroundToScenarios_State4(t *testing.T) {
	m := initialModel()
	m.state = 4

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	result := newModel.(addModel)
	assert.Equal(t, 5, result.state)
	assert.False(t, result.done)
	assert.NotNil(t, cmd)
}

// Ctrl+D on state 5 (Scenarios) advances to state 6 (Examples)
func TestAddModel_CtrlD_ScenariosToExamples_State5(t *testing.T) {
	m := initialModel()
	m.state = 5

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	result := newModel.(addModel)
	assert.Equal(t, 6, result.state)
	assert.False(t, result.done)
	assert.NotNil(t, cmd)
}

// Ctrl+D on state 6 (Examples) finishes
func TestAddModel_CtrlD_Done_State6(t *testing.T) {
	m := initialModel()
	m.state = 6

	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	result := newModel.(addModel)
	assert.True(t, result.done)
	assert.NotNil(t, cmd)
}

// Ctrl+D on non-textarea states has no effect
func TestAddModel_CtrlD_NoEffect_State0(t *testing.T) {
	m := initialModel()
	m.state = 0

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	result := newModel.(addModel)
	assert.Equal(t, 0, result.state)
	assert.False(t, result.done)
}

// Enter on textarea states (4,5,6) does NOT advance — it's consumed by textarea
func TestAddModel_Enter_NoAdvance_State4(t *testing.T) {
	m := initialModel()
	m.state = 4

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := newModel.(addModel)
	assert.Equal(t, 4, result.state)
	assert.False(t, result.done)
}

func TestAddModel_TextInput_State0(t *testing.T) {
	m := initialModel()
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("A")})
	result := newModel.(addModel)
	assert.Equal(t, 0, result.state)
}

func TestAddModel_TextInput_State3(t *testing.T) {
	m := initialModel()
	m.state = 3
	m.description.Focus()
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("D")})
	result := newModel.(addModel)
	assert.Equal(t, 3, result.state)
}

func TestAddModel_TextInput_State4(t *testing.T) {
	m := initialModel()
	m.state = 4
	m.background.Focus()
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("B")})
	result := newModel.(addModel)
	assert.Equal(t, 4, result.state)
}

func TestAddModel_TextInput_State5(t *testing.T) {
	m := initialModel()
	m.state = 5
	m.scenarios.Focus()
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("S")})
	result := newModel.(addModel)
	assert.Equal(t, 5, result.state)
}

func TestAddModel_TextInput_State6(t *testing.T) {
	m := initialModel()
	m.state = 6
	m.examples.Focus()
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("E")})
	result := newModel.(addModel)
	assert.Equal(t, 6, result.state)
}

func TestAddModel_NavigateUpDown_State0_NoEffect(t *testing.T) {
	m := initialModel()
	m.state = 0

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	result := newModel.(addModel)
	assert.Equal(t, 0, result.cursor)

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	result = newModel.(addModel)
	assert.Equal(t, 0, result.cursor)
}

// --- parseScenarios tests ---

func TestParseScenarios_Empty(t *testing.T) {
	result := parseScenarios("")
	assert.Empty(t, result)
}

func TestParseScenarios_SingleScenario(t *testing.T) {
	result := parseScenarios("Given a user exists\nWhen they log in\nThen they see dashboard")
	assert.Len(t, result, 1)
	assert.Contains(t, result[0], "Given a user exists")
}

func TestParseScenarios_MultipleScenarios(t *testing.T) {
	input := "Given a user exists\nWhen they log in\nThen they see dashboard\n\nGiven admin is active\nWhen they delete user\nThen user is removed"
	result := parseScenarios(input)
	assert.Len(t, result, 2)
	assert.Contains(t, result[0], "Given a user exists")
	assert.Contains(t, result[1], "Given admin is active")
}

func TestParseScenarios_IgnoresEmptyBlocks(t *testing.T) {
	input := "Given x\nWhen y\nThen z\n\n\n\nGiven a\nWhen b\nThen c"
	result := parseScenarios(input)
	assert.Len(t, result, 2)
}

func TestParseScenarios_TrimsWhitespace(t *testing.T) {
	input := "  Given x  \n  When y  \n\n  Given a  "
	result := parseScenarios(input)
	assert.Len(t, result, 2)
	assert.Equal(t, "Given x  \n  When y", result[0])
	assert.Equal(t, "Given a", result[1])
}

// --- parseExamples tests ---

func TestParseExamples_Empty(t *testing.T) {
	result := parseExamples("")
	assert.Empty(t, result)
}

func TestParseExamples_SingleRow(t *testing.T) {
	result := parseExamples("username,status,expected")
	assert.Len(t, result, 1)
	assert.Equal(t, []string{"username", "status", "expected"}, result[0])
}

func TestParseExamples_MultipleRows(t *testing.T) {
	input := "username,status,expected\njohn_doe,active,success\njane_smith,inactive,blocked"
	result := parseExamples(input)
	assert.Len(t, result, 3)
	assert.Equal(t, []string{"username", "status", "expected"}, result[0])
	assert.Equal(t, []string{"john_doe", "active", "success"}, result[1])
	assert.Equal(t, []string{"jane_smith", "inactive", "blocked"}, result[2])
}

func TestParseExamples_TrimsWhitespace(t *testing.T) {
	result := parseExamples("  a , b , c  ")
	assert.Len(t, result, 1)
	assert.Equal(t, []string{"a", "b", "c"}, result[0])
}

func TestParseExamples_SkipsEmptyLines(t *testing.T) {
	input := "a,b\n\nc,d"
	result := parseExamples(input)
	assert.Len(t, result, 2)
}
