package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status string
		valid  bool
	}{
		{"planned", true},
		{"implemented", true},
		{"deprecated", true},
		{"", false},
		{"invalid", false},
		{"PLANNED", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			assert.Equal(t, tt.valid, IsValidStatus(tt.status))
		})
	}
}

func TestFilterPlansByStatus_EmptyFilterReturnsAll(t *testing.T) {
	plans := []TestPlan{
		{Feature: "Planned Feature", Type: "functional", Status: "planned", Risk: "stable"},
		{Feature: "Implemented Feature", Type: "security", Status: "implemented", Risk: "stable"},
		{Feature: "Deprecated Feature", Type: "solution", Status: "deprecated", Risk: "stable"},
	}

	filtered := FilterPlansByStatus(plans, "")
	assert.Equal(t, plans, filtered)
}

func TestFilterPlansByStatus_PlannedOnly(t *testing.T) {
	plans := []TestPlan{
		{Feature: "Planned Feature", Type: "functional", Status: "planned", Risk: "stable"},
		{Feature: "Implemented Feature", Type: "security", Status: "implemented", Risk: "stable"},
		{Feature: "Deprecated Feature", Type: "solution", Status: "deprecated", Risk: "stable"},
		{Feature: "Another Planned", Type: "performance", Status: "planned", Risk: "beta"},
	}

	filtered := FilterPlansByStatus(plans, "planned")
	assert.Len(t, filtered, 2)
	for _, p := range filtered {
		assert.Equal(t, "planned", p.Status)
	}
}

func TestFilterPlansByStatus_ImplementedOnly(t *testing.T) {
	plans := []TestPlan{
		{Feature: "Planned Feature", Type: "functional", Status: "planned", Risk: "stable"},
		{Feature: "Implemented Feature", Type: "security", Status: "implemented", Risk: "stable"},
		{Feature: "Deprecated Feature", Type: "solution", Status: "deprecated", Risk: "stable"},
	}

	filtered := FilterPlansByStatus(plans, "implemented")
	assert.Len(t, filtered, 1)
	assert.Equal(t, "Implemented Feature", filtered[0].Feature)
}

func TestFilterPlansByStatus_DeprecatedOnly(t *testing.T) {
	plans := []TestPlan{
		{Feature: "Planned Feature", Type: "functional", Status: "planned", Risk: "stable"},
		{Feature: "Implemented Feature", Type: "security", Status: "implemented", Risk: "stable"},
		{Feature: "Deprecated Feature", Type: "solution", Status: "deprecated", Risk: "stable"},
	}

	filtered := FilterPlansByStatus(plans, "deprecated")
	assert.Len(t, filtered, 1)
	assert.Equal(t, "Deprecated Feature", filtered[0].Feature)
}

func TestFilterPlansByStatus_NoMatches(t *testing.T) {
	plans := []TestPlan{
		{Feature: "Planned Feature", Type: "functional", Status: "planned", Risk: "stable"},
		{Feature: "Implemented Feature", Type: "security", Status: "implemented", Risk: "stable"},
	}

	// No plan is deprecated
	filtered := FilterPlansByStatus(plans, "deprecated")
	assert.Empty(t, filtered)
}

func TestFilterPlansByStatus_InvalidFilter(t *testing.T) {
	plans := []TestPlan{
		{Feature: "Planned Feature", Type: "functional", Status: "planned", Risk: "stable"},
	}

	filtered := FilterPlansByStatus(plans, "bogus")
	assert.Empty(t, filtered)
}
