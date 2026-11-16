package rules //nolint:dupl // intentional near-copy of DL3012/DL4004: the port keeps one file per hadolint rule

import (
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

type cmdState int

const (
	noCmd cmdState = iota
	hasCmd
)

// DL4003Rule is a stateful rule that tracks CMD instructions.
// Ported from Hadolint.Rule.DL4003.
type DL4003Rule struct {
	rule.StatefulRuleBase
}

// DL4003 creates the rule for checking multiple CMD instructions.
func DL4003() rule.Rule {
	return &DL4003Rule{StatefulRuleBase: rule.NewStatefulRuleBase(DL4003Meta)}
}

// InitialState returns the initial state for this rule.
func (*DL4003Rule) InitialState() rule.State {
	return rule.EmptyState(noCmd)
}

// Check implements the stateful check for multiple CMD instructions.
// Ported from the check function in DL4003.hs.
func (r *DL4003Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	// Reset state on each FROM (new stage)
	if _, ok := instruction.(*syntax.From); ok {
		return state.ReplaceData(noCmd)
	}

	// Check for CMD instruction
	if _, ok := instruction.(*syntax.Cmd); ok {
		currentState := rule.Data[cmdState](state)

		if currentState == noCmd {
			// First CMD in this stage - remember it
			return state.ReplaceData(hasCmd)
		}

		// Second CMD in this stage - add failure
		return state.AddFailure(rule.Fail(r, line))
	}

	return state
}
