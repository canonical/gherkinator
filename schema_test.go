package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSchema_ValidPlans(t *testing.T) {
	validTypes := []string{"functional", "solution", "performance", "reliability", "security"}
	validStatuses := []string{"planned", "implemented", "deprecated"}
	validRisks := []string{"edge", "beta", "candidate", "stable"}

	for _, tp := range validTypes {
		for _, st := range validStatuses {
			for _, r := range validRisks {
				t.Run(tp+"_"+st+"_"+r, func(t *testing.T) {
					plan := TestPlan{Type: tp, Status: st, Risk: r}
					err := ValidateSchema(plan)
					assert.NoError(t, err)
				})
			}
		}
	}
}

func TestValidateSchema_InvalidType(t *testing.T) {
	plan := TestPlan{Type: "invalid_type", Status: "implemented"}
	err := ValidateSchema(plan)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid type 'invalid_type'")
}

func TestValidateSchema_EmptyType(t *testing.T) {
	plan := TestPlan{Type: "", Status: "planned"}
	err := ValidateSchema(plan)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid type ''")
}

func TestValidateSchema_InvalidStatus(t *testing.T) {
	plan := TestPlan{Type: "functional", Status: "invalid_status", Risk: "stable"}
	err := ValidateSchema(plan)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status 'invalid_status'")
}

func TestValidateSchema_InvalidRisk(t *testing.T) {
	plan := TestPlan{Type: "functional", Status: "planned", Risk: "invalid_risk"}
	err := ValidateSchema(plan)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid risk 'invalid_risk'")
}

func TestValidateSchema_EmptyRisk(t *testing.T) {
	plan := TestPlan{Type: "functional", Status: "planned", Risk: ""}
	err := ValidateSchema(plan)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid risk ''")
}

func TestValidateSchema_BothInvalid(t *testing.T) {
	plan := TestPlan{Type: "bad", Status: "bad", Risk: "bad"}
	err := ValidateSchema(plan)
	assert.Error(t, err)
	// Should fail on type first
	assert.Contains(t, err.Error(), "invalid type 'bad'")
}

func TestValidateSchema_EmptyStatus(t *testing.T) {
	plan := TestPlan{Type: "functional", Status: "", Risk: "stable"}
	err := ValidateSchema(plan)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status ''")
}
