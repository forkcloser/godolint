package rules

import (
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3023Rule checks that COPY --from doesn't reference its own FROM alias.
// Ported from Hadolint.Rule.DL3023.
type DL3023Rule struct {
	rule.StatefulRuleBase
}

// DL3023 creates the rule for checking COPY --from self-reference.
func DL3023() rule.Rule {
	return &DL3023Rule{StatefulRuleBase: rule.NewStatefulRuleBase(DL3023Meta)}
}

// InitialState returns the initial state for this rule.
func (*DL3023Rule) InitialState() rule.State {
	return rule.EmptyState("") // Empty string = no current stage alias
}

// Check tracks current stage alias and validates COPY --from doesn't self-reference.
// Ported from the check function in DL3023.hs.
func (r *DL3023Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	currentAlias := rule.Data[string](state)

	if from, ok := instruction.(*syntax.From); ok {
		if from.Image.Alias != nil {
			return state.ReplaceData(*from.Image.Alias)
		}

		return state.ReplaceData("") // No alias
	}

	if copyInstr, ok := instruction.(*syntax.Copy); ok {
		if copyInstr.From != nil && currentAlias != "" {
			if *copyInstr.From == currentAlias {
				return state.AddFailure(rule.Fail(r, line))
			}
		}
	}

	return state
}
