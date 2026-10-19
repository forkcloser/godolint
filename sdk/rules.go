package sdk

import (
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
)

// AllRules returns every implemented hadolint DL#### rule (pure Go).
// Shellcheck integration (validates RUN instruction shell scripts via external binary)
// is opt-in via WithShellcheck() and adds SC#### violations.
func AllRules() []rule.Rule {
	return []rule.Rule{
		// DL1xxx - Miscellaneous
		rules.DL1001(),
		// DL3xxx - Dockerfile best practices
		rules.DL3000(),
		rules.DL3001(),
		rules.DL3002(),
		rules.DL3003(),
		rules.DL3004(),
		rules.DL3006(),
		rules.DL3007(),
		rules.DL3008(),
		rules.DL3009(),
		rules.DL3010(),
		rules.DL3011(),
		rules.DL3012(),
		rules.DL3013(),
		rules.DL3014(),
		rules.DL3015(),
		rules.DL3016(),
		rules.DL3018(),
		rules.DL3019(),
		rules.DL3020(),
		rules.DL3021(),
		rules.DL3022(),
		rules.DL3023(),
		rules.DL3024(),
		rules.DL3025(),
		rules.DL3026(),
		rules.DL3027(),
		rules.DL3028(),
		rules.DL3029(),
		rules.DL3030(),
		rules.DL3032(),
		rules.DL3033(),
		rules.DL3034(),
		rules.DL3035(),
		rules.DL3036(),
		rules.DL3037(),
		rules.DL3038(),
		rules.DL3040(),
		rules.DL3041(),
		rules.DL3042(),
		rules.DL3043(),
		rules.DL3044(),
		rules.DL3045(),
		rules.DL3046(),
		rules.DL3047(),
		rules.DL3048(),
		rules.DL3049(),
		rules.DL3050(),
		rules.DL3051(),
		rules.DL3052(),
		rules.DL3053(),
		rules.DL3054(),
		rules.DL3055(),
		rules.DL3056(),
		rules.DL3057(),
		rules.DL3058(),
		rules.DL3059(),
		rules.DL3060(),
		rules.DL3061(),
		rules.DL3062(),
		rules.DL3063(),
		rules.DL3064(),
		rules.DL3065(),
		rules.DL3066(),
		rules.DL3067(),
		rules.DL3068(),
		// DL4xxx - Deprecated instructions
		rules.DL4000(),
		rules.DL4001(),
		rules.DL4003(),
		rules.DL4004(),
		rules.DL4005(),
		rules.DL4006(),
	}
}

// RuleSet represents a predefined set of rules.
type RuleSet string

const (
	// RuleSetAll includes all implemented rules.
	RuleSetAll RuleSet = "all"
	// RuleSetRecommended includes rules with Error and Warning severity.
	RuleSetRecommended RuleSet = "recommended"
	// RuleSetStrict includes all rules (same as All currently).
	RuleSetStrict RuleSet = "strict"
)

// GetRuleSet returns rules for a predefined rule set.
func GetRuleSet(set RuleSet) []rule.Rule {
	all := AllRules()

	// The explicit Strict/All cases document the mapping; every other (invalid)
	// value also falls back to the full set, so default mirrors them on purpose.
	//revive:disable:identical-switch-branches intentional, see above.
	switch set {
	case RuleSetRecommended:
		// Filter to Error and Warning severity only
		var filtered []rule.Rule

		for _, r := range all {
			if r.Severity() == rule.Error || r.Severity() == rule.Warning {
				filtered = append(filtered, r)
			}
		}

		return filtered
	case RuleSetStrict, RuleSetAll:
		return all
	default:
		return all
	}
}

// FilterRules returns a new rule set with specified rules disabled.
// disabledCodes should contain rule codes like "DL3000", "DL3007", etc.
func FilterRules(ruleSet []rule.Rule, disabledCodes []string) []rule.Rule {
	if len(disabledCodes) == 0 {
		return ruleSet
	}

	// Build lookup map
	disabled := make(map[string]bool)
	for _, code := range disabledCodes {
		disabled[code] = true
	}

	// Filter rules
	var filtered []rule.Rule

	for _, r := range ruleSet {
		if !disabled[string(r.Code())] {
			filtered = append(filtered, r)
		}
	}

	return filtered
}
