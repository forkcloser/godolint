package sdk

import (
	"slices"
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
)

// RuleSet represents a predefined set of rules.
type RuleSet string

const (
	// RuleSetAll includes all implemented rules.
	RuleSetAll RuleSet = "all"
	// RuleSetRecommended includes rules with Error and Warning severity.
	RuleSetRecommended RuleSet = "recommended"
)

// ruleSet is the rules of a set: the error and warning rules for
// RuleSetRecommended, every rule for RuleSetAll and for any other value.
func ruleSet(set RuleSet) []rule.Rule {
	all := rules.All()
	if set != RuleSetRecommended {
		return all
	}

	var recommended []rule.Rule

	for _, r := range all {
		if r.Severity() == rule.Error || r.Severity() == rule.Warning {
			recommended = append(recommended, r)
		}
	}

	return recommended
}

// Rules describes every rule a Linter has by default, ordered by code. The slice is
// a new one on every call, the caller's to change. The codes shellcheck
// reports (SC####) are not rules of this set.
func Rules() []RuleInfo {
	all := rules.All()
	infos := make([]RuleInfo, 0, len(all))

	for _, r := range all {
		infos = append(infos, RuleInfo{
			Code:     string(r.Code()),
			Severity: convertSeverity(r.Severity()),
			Message:  r.Message(),
		})
	}

	slices.SortFunc(infos, func(a, b RuleInfo) int { return strings.Compare(a.Code, b.Code) })

	return infos
}

// Rule describes the rule with the given code, and reports whether there is
// one. The code is matched as written: "DL3006", not "dl3006".
func Rule(code string) (RuleInfo, bool) {
	for _, info := range Rules() {
		if info.Code == code {
			return info, true
		}
	}

	return RuleInfo{}, false
}
