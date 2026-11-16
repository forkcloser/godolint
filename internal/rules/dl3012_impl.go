package rules //nolint:dupl // intentional near-copy of DL4003/DL4004: the port keeps one file per hadolint rule

import (
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

type healthcheckState int

const (
	noHealthcheck healthcheckState = iota
	hasHealthcheck
)

// DL3012Rule is a stateful rule that tracks HEALTHCHECK instructions.
// Ported from Hadolint.Rule.DL3012.
type DL3012Rule struct {
	rule.StatefulRuleBase
}

// DL3012 creates the rule for checking multiple HEALTHCHECK instructions.
func DL3012() rule.Rule {
	return &DL3012Rule{StatefulRuleBase: rule.NewStatefulRuleBase(DL3012Meta)}
}

// InitialState returns the initial state for this rule.
func (*DL3012Rule) InitialState() rule.State {
	return rule.EmptyState(noHealthcheck)
}

// Check implements the stateful check for multiple HEALTHCHECK instructions.
// Ported from the check function in DL3012.hs.
func (r *DL3012Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	// Reset state on each FROM (new stage)
	if _, ok := instruction.(*syntax.From); ok {
		return state.ReplaceData(noHealthcheck)
	}

	// Check for HEALTHCHECK instruction
	if _, ok := instruction.(*syntax.Healthcheck); ok {
		currentState := rule.Data[healthcheckState](state)

		if currentState == noHealthcheck {
			// First HEALTHCHECK in this stage - remember it
			return state.ReplaceData(hasHealthcheck)
		}

		// Second HEALTHCHECK in this stage - add failure
		return state.AddFailure(rule.Fail(r, line))
	}

	return state
}
