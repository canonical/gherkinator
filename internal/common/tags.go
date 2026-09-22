package common

import (
	"fmt"
	"regexp"
)

// validTagPattern matches tags that are non-empty and contain neither
// whitespace nor commas.  Whitespace would break Gherkin tag rendering,
// and commas are the --tag CSV separator, so a tag containing one could
// never be matched from the CLI.
var validTagPattern = regexp.MustCompile(`^[^\s,]+$`)

// ValidateTag reports whether tag satisfies the tag format rules: it
// must be non-empty and contain no whitespace or commas.  It is used
// both by ValidateSchema (for writes) and by the --tag flag check on
// the generate, serve, and diff subcommands.
func ValidateTag(tag string) error {
	if !validTagPattern.MatchString(tag) {
		return fmt.Errorf("invalid tag '%s': tags must be non-empty and contain no whitespace or commas", tag)
	}
	return nil
}

// FilterPlansByTags returns the subset of plans carrying at least one
// tag that matches any requested tag (union semantics).  An empty
// tagFilters slice returns the input unchanged, disabling tag
// filtering.  Plans without tags are excluded whenever tag filtering
// is active, and matching is exact and case-sensitive.
func FilterPlansByTags(plans []TestPlan, tagFilters []string) []TestPlan {
	if len(tagFilters) == 0 {
		return plans
	}
	requested := make(map[string]bool, len(tagFilters))
	for _, tag := range tagFilters {
		requested[tag] = true
	}
	var filtered []TestPlan
	for _, plan := range plans {
		for _, tag := range plan.Tags {
			if requested[tag] {
				filtered = append(filtered, plan)
				break
			}
		}
	}
	return filtered
}
