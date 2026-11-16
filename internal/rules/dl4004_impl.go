package rules //nolint:dupl // intentional near-copy of DL3012/DL4003: the port keeps one file per hadolint rule

import (
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

type entrypointState int

const (
	noEntrypoint entrypointState = iota
	hasEntrypoint
)

// DL4004Rule is a stateful rule that tracks ENTRYPOINT instructions.
// Ported from Hadolint.Rule.DL4004.
type DL4004Rule struct {
	rule.StatefulRuleBase
}

// DL4004 creates the rule for checking multiple ENTRYPOINT instructions.
func DL4004() rule.Rule {
	return &DL4004Rule{StatefulRuleBase: rule.NewStatefulRuleBase(DL4004Meta)}
}

// InitialState returns the initial state for this rule.
func (*DL4004Rule) InitialState() rule.State {
	return rule.EmptyState(noEntrypoint)
}

// Check implements the stateful check for multiple ENTRYPOINT instructions.
// Ported from the check function in DL4004.hs.
func (r *DL4004Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	if _, ok := instruction.(*syntax.From); ok {
		return state.ReplaceData(noEntrypoint)
	}

	if _, ok := instruction.(*syntax.Entrypoint); ok {
		currentState := rule.Data[entrypointState](state)

		if currentState == noEntrypoint {
			return state.ReplaceData(hasEntrypoint)
		}

		return state.AddFailure(rule.Fail(r, line))
	}

	return state
}
