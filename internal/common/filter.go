package common

// FilterOptions carries the plan-selection filters shared by the
// generate, serve, and diff subcommands.  The zero value disables all
// filtering.
type FilterOptions struct {
	// Risk is the cumulative risk ladder bound (see FilterPlansByRisk).
	Risk string
	// Status is the exact-match status filter (see FilterPlansByStatus).
	Status string
	// Tags is the union-match tag filter (see FilterPlansByTags).
	Tags []string
}

// Apply returns the subset of plans that satisfy every non-empty filter
// dimension: status (exact), risk (cumulative), and tags (union).
func (o FilterOptions) Apply(plans []TestPlan) []TestPlan {
	plans = FilterPlansByStatus(plans, o.Status)
	plans = FilterPlansByRisk(plans, o.Risk)
	plans = FilterPlansByTags(plans, o.Tags)
	return plans
}
