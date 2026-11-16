package rules

import (
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3061Rule checks that Dockerfile begins with FROM, ARG, or comment.
// Ported from Hadolint.Rule.DL3061.
type DL3061Rule struct {
	rule.StatefulRuleBase
}

// DL3061 creates the rule for checking instruction order.
func DL3061() rule.Rule {
	return &DL3061Rule{StatefulRuleBase: rule.NewStatefulRuleBase(DL3061Meta)}
}

// InitialState returns the initial state for this rule.
func (*DL3061Rule) InitialState() rule.State {
	return rule.EmptyState(false) // false = haven't seen FROM yet
}

// Check ensures Dockerfile starts with FROM, ARG, or comment.
// Ported from the check function in DL3061.hs.
func (r *DL3061Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	seenFrom := rule.Data[bool](state)

	if seenFrom {
		return state
	}

	if _, ok := instruction.(*syntax.From); ok {
		return state.ReplaceData(true)
	}

	if _, ok := instruction.(*syntax.Arg); ok {
		return state
	}

	if _, ok := instruction.(*syntax.Comment); ok {
		return state
	}

	return state.AddFailure(rule.Fail(r, line))
}
