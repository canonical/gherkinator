package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateTag_AcceptsValidTags(t *testing.T) {
	// Any non-empty string without whitespace or commas is a valid tag,
	// including hyphens, underscores, dots, and uppercase letters.
	for _, tag := range []string{"single-node", "multi-node", "minimal", "gpu", "a", "UPPER", "with_underscore", "with.dot"} {
		assert.NoError(t, ValidateTag(tag), "tag %q should be valid", tag)
	}
}

func TestValidateTag_RejectsEmpty(t *testing.T) {
	err := ValidateTag("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tag ''")
}

func TestValidateTag_RejectsWhitespace(t *testing.T) {
	err := ValidateTag("foo bar")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tag 'foo bar'")
	assert.Contains(t, err.Error(), "no whitespace or commas")
}

func TestValidateTag_RejectsInternalWhitespace(t *testing.T) {
	// The rejection is not limited to the space character.
	assert.Error(t, ValidateTag("foo\tbar"))
	assert.Error(t, ValidateTag("foo\nbar"))
}

func TestValidateTag_RejectsComma(t *testing.T) {
	// Commas are forbidden because they are the --tag CSV separator; a
	// comma-containing tag could never be matched from the CLI.
	err := ValidateTag("a,b")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tag 'a,b'")
}

func TestFilterPlansByTags_NoFilterReturnsInputUnchanged(t *testing.T) {
	// An absent (nil) or empty tag filter disables tag filtering, so
	// untagged plans pass through untouched.
	plans := []TestPlan{
		{Feature: "Tagged", Tags: []string{"single-node"}},
		{Feature: "Untagged"},
	}
	assert.Equal(t, plans, FilterPlansByTags(plans, nil))
	assert.Equal(t, plans, FilterPlansByTags(plans, []string{}))
}

func TestFilterPlansByTags_SingleTag(t *testing.T) {
	plans := []TestPlan{
		{Feature: "Single", Tags: []string{"single-node", "minimal"}},
		{Feature: "Multi", Tags: []string{"multi-node"}},
		{Feature: "Untagged"},
	}
	filtered := FilterPlansByTags(plans, []string{"single-node"})
	require.Len(t, filtered, 1)
	assert.Equal(t, "Single", filtered[0].Feature)
}

func TestFilterPlansByTags_UnionAcrossMultipleFilters(t *testing.T) {
	// Multiple requested tags use union semantics: a plan matching any
	// requested tag is included.
	plans := []TestPlan{
		{Feature: "Single", Tags: []string{"single-node"}},
		{Feature: "Multi", Tags: []string{"multi-node"}},
		{Feature: "Both", Tags: []string{"single-node", "multi-node"}},
		{Feature: "Minimal", Tags: []string{"minimal"}},
	}
	filtered := FilterPlansByTags(plans, []string{"single-node", "minimal"})
	require.Len(t, filtered, 3)
	names := []string{filtered[0].Feature, filtered[1].Feature, filtered[2].Feature}
	assert.Contains(t, names, "Single")
	assert.Contains(t, names, "Both")
	assert.Contains(t, names, "Minimal")
	assert.NotContains(t, names, "Multi")
}

func TestFilterPlansByTags_UntaggedPlansExcludedWhenFilterActive(t *testing.T) {
	// Once a tag filter is active, only plans explicitly classified with
	// a matching tag render; untagged (config-agnostic) plans are
	// excluded by design.
	plans := []TestPlan{
		{Feature: "Untagged"},
		{Feature: "Tagged", Tags: []string{"multi-node"}},
	}
	filtered := FilterPlansByTags(plans, []string{"multi-node"})
	require.Len(t, filtered, 1)
	assert.Equal(t, "Tagged", filtered[0].Feature)
}

func TestFilterPlansByTags_MatchingIsCaseSensitive(t *testing.T) {
	// Tag matching is exact and case-sensitive, consistent with the
	// --risk and --status enum filters.
	plans := []TestPlan{{Feature: "Tagged", Tags: []string{"multi-node"}}}
	assert.Empty(t, FilterPlansByTags(plans, []string{"Multi-Node"}))
}

func TestFilterPlansByTags_NoMatchesReturnsEmpty(t *testing.T) {
	plans := []TestPlan{{Feature: "Single", Tags: []string{"single-node"}}}
	assert.Empty(t, FilterPlansByTags(plans, []string{"gpu"}))
}
