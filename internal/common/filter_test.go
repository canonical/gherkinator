package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterOptions_ZeroValueReturnsAllPlans(t *testing.T) {
	// The zero FilterOptions disables every filter dimension, so plans
	// are returned unfiltered.
	plans := []TestPlan{
		{Feature: "A", Type: "functional", Status: "planned", Risk: "edge"},
		{Feature: "B", Type: "security", Status: "implemented", Risk: "stable", Tags: []string{"multi-node"}},
	}
	assert.Equal(t, plans, FilterOptions{}.Apply(plans))
}

func TestFilterOptions_IntersectsAllDimensions(t *testing.T) {
	// A plan must satisfy every non-empty dimension: cumulative risk,
	// exact status, and union tag membership.
	plans := []TestPlan{
		{Feature: "EdgeTagged", Type: "functional", Status: "implemented", Risk: "edge", Tags: []string{"multi-node"}},
		{Feature: "EdgeUntagged", Type: "functional", Status: "implemented", Risk: "edge"},
		{Feature: "StableTagged", Type: "functional", Status: "implemented", Risk: "stable", Tags: []string{"multi-node"}},
		{Feature: "EdgeTaggedPlanned", Type: "functional", Status: "planned", Risk: "edge", Tags: []string{"multi-node"}},
	}
	filtered := FilterOptions{Risk: "candidate", Status: "implemented", Tags: []string{"multi-node"}}.Apply(plans)
	require.Len(t, filtered, 1)
	assert.Equal(t, "EdgeTagged", filtered[0].Feature)
}
