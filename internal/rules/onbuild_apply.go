package rules

import "github.com/forkcloser/godolint/internal/rule"

// InOnBuild reports whether a rule is also applied to the instruction an
// ONBUILD wraps, as hadolint composes it: one of the rules of the generated
// table, or a rule that says so of itself.
func InOnBuild(applied rule.Rule) bool {
	if own, says := applied.(interface{ InOnBuild() bool }); says {
		return own.InOnBuild()
	}

	return onBuild[applied.Code()]
}
