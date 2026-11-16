package rules

import (
	"maps"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// dl3024State tracks seen FROM aliases to detect duplicates.
type dl3024State struct {
	aliases map[string]int // map[alias]line - tracks where each alias was defined
}

// DL3024Rule checks that FROM aliases are unique.
// Ported from Hadolint.Rule.DL3024.
type DL3024Rule struct {
	rule.StatefulRuleBase
}

// DL3024 creates the rule for checking FROM aliases are unique.
func DL3024() rule.Rule {
	return &DL3024Rule{StatefulRuleBase: rule.NewStatefulRuleBase(DL3024Meta)}
}

// InitialState returns the initial state for this rule.
func (*DL3024Rule) InitialState() rule.State {
	return rule.EmptyState(dl3024State{
		aliases: make(map[string]int),
	})
}

// Check tracks FROM aliases and reports duplicates.
// Ported from the check function in DL3024.hs.
func (r *DL3024Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	from, ok := instruction.(*syntax.From)
	if !ok {
		return state
	}

	if from.Image.Alias == nil {
		return state
	}

	currentState := rule.Data[dl3024State](state)
	alias := *from.Image.Alias

	if _, exists := currentState.aliases[alias]; exists {
		return state.AddFailure(rule.Fail(r, line))
	}

	newAliases := make(map[string]int)
	maps.Copy(newAliases, currentState.aliases)

	newAliases[alias] = line
	currentState.aliases = newAliases

	return state.ReplaceData(currentState)
}
