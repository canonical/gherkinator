package common

// RiskOrder maps each recognised risk level to its cumulative position in the
// risk ladder. A filter at level N includes all plans at levels <= N.
var RiskOrder = map[string]int{
	"edge":      0,
	"beta":      1,
	"candidate": 2,
	"stable":    3,
}

// IsValidRisk reports whether risk is one of the recognised risk levels
// (edge, beta, candidate, stable). An empty string is not considered valid;
// callers should treat an empty risk as "no filter".
func IsValidRisk(risk string) bool {
	_, ok := RiskOrder[risk]
	return ok
}

// FilterPlansByRisk returns a subset of plans whose risk level is at or below
// filterRisk in the risk ladder. An empty filterRisk returns the input
// unchanged. An unrecognised filterRisk returns an empty slice.
func FilterPlansByRisk(plans []TestPlan, filterRisk string) []TestPlan {
	if filterRisk == "" {
		return plans
	}
	maxLevel, ok := RiskOrder[filterRisk]
	if !ok {
		return nil
	}
	var filtered []TestPlan
	for _, plan := range plans {
		level, ok := RiskOrder[plan.Risk]
		if ok && level <= maxLevel {
			filtered = append(filtered, plan)
		}
	}
	return filtered
}
