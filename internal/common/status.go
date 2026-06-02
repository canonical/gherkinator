package common

// validStatuses is the set of recognized implementation statuses for a
// test plan. Used by ValidateSchema (for writes) and FilterPlansByStatus
// (for reads).
var validStatuses = map[string]bool{
	"planned":     true,
	"implemented": true,
	"deprecated":  true,
}

// IsValidStatus reports whether status is one of the recognized
// implementation statuses (planned, implemented, deprecated). An empty
// string is not considered valid; callers should treat an empty status
// as "no filter".
func IsValidStatus(status string) bool {
	return validStatuses[status]
}

// FilterPlansByStatus returns the subset of plans whose status exactly
// matches the statusFilter. An empty statusFilter returns the input
// unchanged. An unrecognised statusFilter returns an empty slice.
//
// Status filtering is exact-match, not cumulative: a plan is included
// only when its status field equals statusFilter byte-for-byte.
func FilterPlansByStatus(plans []TestPlan, statusFilter string) []TestPlan {
	if statusFilter == "" {
		return plans
	}
	if !validStatuses[statusFilter] {
		return nil
	}
	var filtered []TestPlan
	for _, plan := range plans {
		if plan.Status == statusFilter {
			filtered = append(filtered, plan)
		}
	}
	return filtered
}
