package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSchema_ValidPlans(t *testing.T) {
	validTypes := []string{"functional", "solution", "performance", "reliability", "security"}
	validStatuses := []string{"planned", "implemented", "deprecated"}

	for _, tp := range validTypes {
		for _, st := range validStatuses {
			t.Run(tp+"_"+st, func(t *testing.T) {
				plan := TestPlan{Type: tp, Status: st}
				err := ValidateSchema(plan)
				assert.NoError(t, err)
			})
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
	plan := TestPlan{Type: "functional", Status: "invalid_status"}
	err := ValidateSchema(plan)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status 'invalid_status'")
}

func TestValidateSchema_EmptyStatus(t *testing.T) {
	plan := TestPlan{Type: "functional", Status: ""}
	err := ValidateSchema(plan)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status ''")
}

func TestValidateSchema_BothInvalid(t *testing.T) {
	plan := TestPlan{Type: "bad", Status: "bad"}
	err := ValidateSchema(plan)
	assert.Error(t, err)
	// Should fail on type first
	assert.Contains(t, err.Error(), "invalid type 'bad'")
}
