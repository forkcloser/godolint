package rules

import (
	"regexp"
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// dl3044State tracks defined environment variables.
type dl3044State struct {
	definedVars map[string]bool
}

// DL3044Rule checks for ENV self-reference.
type DL3044Rule struct {
	rule.StatefulRuleBase
}

// DL3044 creates the rule for checking ENV self-reference.
func DL3044() rule.Rule {
	return &DL3044Rule{StatefulRuleBase: rule.NewStatefulRuleBase(DL3044Meta)}
}

// InitialState returns the initial state for this rule.
func (*DL3044Rule) InitialState() rule.State {
	return rule.EmptyState(dl3044State{
		definedVars: make(map[string]bool),
	})
}

// Check flags ENV statements that reference a variable defined in the same
// statement.
func (r *DL3044Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	currentState := rule.Data[dl3044State](state)

	switch inst := instruction.(type) {
	case *syntax.Arg:
		currentState.definedVars[inst.ArgName] = true

		return state.ReplaceData(currentState)

	case *syntax.Env:
		reported := referencesUndefinedSibling(inst.Pairs, currentState.definedVars)

		// The names are defined from here on whether the ENV is reported or
		// not: a later ENV that gives one of them again and reads it is in
		// order.
		for _, pair := range inst.Pairs {
			currentState.definedVars[pair.Key] = true
		}

		state = state.ReplaceData(currentState)

		if reported {
			return state.AddFailure(rule.Fail(r, line))
		}

		return state
	}

	return state
}

// referencesUndefinedSibling reports whether a pair of one ENV reads another
// variable that same ENV defines. Docker evaluates the pairs left to right
// against the environment as it was before the statement, so the reference
// resolves to nothing — unless an earlier instruction defined the name, which
// is what defined holds.
func referencesUndefinedSibling(pairs []syntax.EnvPair, defined map[string]bool) bool {
	for i, pair := range pairs {
		for j, otherPair := range pairs {
			if i == j {
				continue // Skip self
			}

			if referencesVar(pair.Value, otherPair.Key) && !defined[otherPair.Key] {
				return true
			}
		}
	}

	return false
}

// referencesVar checks if a value string references a variable.
// Matches ${var} or $var (where var is terminated by non-alphanumeric char).
func referencesVar(value, varName string) bool {
	if strings.Contains(value, "${"+varName+"}") {
		return true
	}

	pattern := regexp.MustCompile(`\$` + regexp.QuoteMeta(varName) + `(?:[^a-zA-Z0-9_]|$)`)

	return pattern.MatchString(value)
}
