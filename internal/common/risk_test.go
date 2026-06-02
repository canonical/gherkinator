package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterPlansByRisk(t *testing.T) {
	edge := "edge"
	beta := "beta"
	candidate := "candidate"
	stable := "stable"

	tests := []struct {
		name       string
		plans      []TestPlan
		filterRisk string
		expected   int
	}{
		{
			name: "no filter returns all",
			plans: []TestPlan{
				{Feature: "Edge", Type: "functional", Status: "planned", Risk: edge},
				{Feature: "Beta", Type: "functional", Status: "planned", Risk: beta},
				{Feature: "Stable", Type: "security", Status: "implemented", Risk: stable},
			},
			filterRisk: "",
			expected:   3,
		},
		{
			name: "filter edge returns only edge",
			plans: []TestPlan{
				{Feature: "Edge", Type: "functional", Status: "planned", Risk: edge},
				{Feature: "Beta", Type: "functional", Status: "planned", Risk: beta},
				{Feature: "Stable", Type: "security", Status: "implemented", Risk: stable},
			},
			filterRisk: "edge",
			expected:   1,
		},
		{
			name: "filter beta returns edge and beta",
			plans: []TestPlan{
				{Feature: "Edge", Type: "functional", Status: "planned", Risk: edge},
				{Feature: "Beta", Type: "functional", Status: "planned", Risk: beta},
				{Feature: "Stable", Type: "security", Status: "implemented", Risk: stable},
			},
			filterRisk: "beta",
			expected:   2,
		},
		{
			name: "filter candidate returns edge, beta, candidate",
			plans: []TestPlan{
				{Feature: "Edge", Type: "functional", Status: "planned", Risk: edge},
				{Feature: "Beta", Type: "functional", Status: "planned", Risk: beta},
				{Feature: "Candidate", Type: "security", Status: "planned", Risk: candidate},
				{Feature: "Stable", Type: "security", Status: "implemented", Risk: stable},
			},
			filterRisk: "candidate",
			expected:   3,
		},
		{
			name: "filter stable returns all",
			plans: []TestPlan{
				{Feature: "Edge", Type: "functional", Status: "planned", Risk: edge},
				{Feature: "Beta", Type: "functional", Status: "planned", Risk: beta},
				{Feature: "Candidate", Type: "security", Status: "planned", Risk: candidate},
				{Feature: "Stable", Type: "security", Status: "implemented", Risk: stable},
			},
			filterRisk: "stable",
			expected:   4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterPlansByRisk(tt.plans, tt.filterRisk)
			assert.Len(t, result, tt.expected)
		})
	}
}

func TestFilterPlansByRisk_InvalidFilter(t *testing.T) {
	edge := "edge"
	plan := TestPlan{Feature: "Test", Type: "functional", Status: "planned", Risk: edge}

	// Invalid filter should return empty (no plans match)
	result := FilterPlansByRisk([]TestPlan{plan}, "invalid")
	assert.Empty(t, result)
}

func TestIsValidRisk(t *testing.T) {
	tests := []struct {
		risk  string
		valid bool
	}{
		{"edge", true},
		{"beta", true},
		{"candidate", true},
		{"stable", true},
		{"", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.risk, func(t *testing.T) {
			assert.Equal(t, tt.valid, IsValidRisk(tt.risk))
		})
	}
}
