package delete

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gherkinator/internal/common"
)

const multiDocYAML = `feature: "Job Submission"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "Submit a job"
---
feature: "GPU Job Submission"
type: "performance"
status: "implemented"
risk: "stable"
scenarios:
  - "Submit a GPU job"
---
feature: "Network Config"
type: "security"
status: "planned"
risk: "stable"
scenarios:
  - "Configure network"
`

func TestDeleteTestPlans_DeleteOne(t *testing.T) {
	plans := []common.TestPlan{
		{Feature: "Job Submission", Type: "functional", Status: "planned", Risk: "stable"},
		{Feature: "GPU Job Submission", Type: "performance", Status: "implemented", Risk: "stable"},
		{Feature: "Network Config", Type: "security", Status: "planned", Risk: "stable"},
	}

	remaining, deleted := DeleteTestPlans(plans, []string{"Job Submission"})
	assert.Len(t, remaining, 2)
	assert.Len(t, deleted, 1)
	assert.Equal(t, "Job Submission", deleted[0])
	assert.Equal(t, "GPU Job Submission", remaining[0].Feature)
	assert.Equal(t, "Network Config", remaining[1].Feature)
}

func TestDeleteTestPlans_DeleteMultiple(t *testing.T) {
	plans := []common.TestPlan{
		{Feature: "Job Submission", Type: "functional", Status: "planned", Risk: "stable"},
		{Feature: "GPU Job Submission", Type: "performance", Status: "implemented", Risk: "stable"},
		{Feature: "Network Config", Type: "security", Status: "planned", Risk: "stable"},
	}

	remaining, deleted := DeleteTestPlans(plans, []string{"Job Submission", "GPU Job Submission"})
	assert.Len(t, remaining, 1)
	assert.Len(t, deleted, 2)
	assert.Equal(t, "Network Config", remaining[0].Feature)
}

func TestDeleteTestPlans_CaseInsensitive(t *testing.T) {
	plans := []common.TestPlan{
		{Feature: "Job Submission", Type: "functional", Status: "planned", Risk: "stable"},
		{Feature: "Network Config", Type: "security", Status: "planned", Risk: "stable"},
	}

	remaining, deleted := DeleteTestPlans(plans, []string{"job submission"})
	assert.Len(t, remaining, 1)
	assert.Len(t, deleted, 1)
	assert.Equal(t, "Job Submission", deleted[0])
}

func TestDeleteTestPlans_NoMatch(t *testing.T) {
	plans := []common.TestPlan{
		{Feature: "Job Submission", Type: "functional", Status: "planned", Risk: "stable"},
	}

	remaining, deleted := DeleteTestPlans(plans, []string{"nonexistent"})
	assert.Len(t, remaining, 1)
	assert.Empty(t, deleted)
}

func TestDeleteTestPlans_DeleteAll(t *testing.T) {
	plans := []common.TestPlan{
		{Feature: "Job Submission", Type: "functional", Status: "planned", Risk: "stable"},
		{Feature: "Network Config", Type: "security", Status: "planned", Risk: "stable"},
	}

	remaining, deleted := DeleteTestPlans(plans, []string{"Job Submission", "Network Config"})
	assert.Empty(t, remaining)
	assert.Len(t, deleted, 2)
}

func TestConfirmDeletion_UserConfirmsY(t *testing.T) {
	reader := strings.NewReader("Y\n")
	result := ConfirmDeletion([]string{"Job Submission"}, reader)
	assert.True(t, result)
}

func TestConfirmDeletion_UserDeclinesN(t *testing.T) {
	reader := strings.NewReader("n\n")
	result := ConfirmDeletion([]string{"Job Submission"}, reader)
	assert.False(t, result)
}

func TestConfirmDeletion_UserDeclinesEmpty(t *testing.T) {
	reader := strings.NewReader("\n")
	result := ConfirmDeletion([]string{"Job Submission"}, reader)
	assert.False(t, result)
}

func TestConfirmDeletion_UserDeclinesLowercaseY(t *testing.T) {
	reader := strings.NewReader("y\n")
	result := ConfirmDeletion([]string{"Job Submission"}, reader)
	assert.False(t, result)
}

func TestConfirmDeletion_EmptyInput(t *testing.T) {
	reader := strings.NewReader("")
	result := ConfirmDeletion([]string{"Job Submission"}, reader)
	assert.False(t, result)
}

func TestConfirmDeletion_MultipleNames(t *testing.T) {
	reader := strings.NewReader("Y\n")
	result := ConfirmDeletion([]string{"Job Submission", "GPU Job Submission"}, reader)
	assert.True(t, result)
}

func TestQuoteNames(t *testing.T) {
	result := quoteNames([]string{"hello", "world"})
	assert.Equal(t, []string{`"hello"`, `"world"`}, result)
}

// TestDeleteRoundTrip exercises the full delete workflow using the public
// common package helpers to load/write plans, ensuring the YAML I/O layer
// plays well with DeleteTestPlans.
func TestDeleteRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(multiDocYAML), 0644)
	require.NoError(t, err)

	plans, err := common.LoadTestPlans(inputFile)
	require.NoError(t, err)
	require.Len(t, plans, 3)

	remaining, deleted := DeleteTestPlans(plans, []string{"Job Submission"})
	require.Len(t, remaining, 2)
	require.Len(t, deleted, 1)

	err = common.WriteTestPlans(inputFile, remaining)
	require.NoError(t, err)

	loaded, err := common.LoadTestPlans(inputFile)
	require.NoError(t, err)
	assert.Len(t, loaded, 2)
	for _, p := range loaded {
		assert.NotEqual(t, "Job Submission", p.Feature)
	}
}
