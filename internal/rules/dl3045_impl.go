package rules

import (
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3045 checks for COPY to relative destination without WORKDIR set.
func DL3045() rule.Rule {
	return &DL3045Rule{StatefulRuleBase: rule.NewStatefulRuleBase(DL3045Meta)}
}

// DL3045Rule implements the rule for checking COPY with relative paths.
type DL3045Rule struct {
	rule.StatefulRuleBase
}

// InitialState returns the initial state for this rule.
func (*DL3045Rule) InitialState() rule.State {
	return rule.EmptyState(dl3045State{
		WorkdirSet: make(map[string]bool),
	})
}

type dl3045State struct {
	CurrentStage string
	WorkdirSet   map[string]bool
}

// Check validates COPY instructions use absolute paths or have WORKDIR set.
func (r *DL3045Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	currentState := rule.Data[dl3045State](state)

	switch inst := instruction.(type) {
	case *syntax.From:
		stageName := inst.Image.Image
		if inst.Image.Alias != nil {
			stageName = *inst.Image.Alias
		}

		inheritWorkdir := false
		if currentState.WorkdirSet[inst.Image.Image] {
			inheritWorkdir = true
		}

		currentState.CurrentStage = stageName
		currentState.WorkdirSet[stageName] = inheritWorkdir

		return state.ReplaceData(currentState)

	case *syntax.Workdir:
		currentState.WorkdirSet[currentState.CurrentStage] = true

		return state.ReplaceData(currentState)

	case *syntax.Copy:
		dest := inst.Destination
		dest = strings.Trim(dest, "\"'")

		if currentState.WorkdirSet[currentState.CurrentStage] {
			return state.ReplaceData(currentState)
		}

		if strings.HasPrefix(dest, "/") {
			return state.ReplaceData(currentState)
		}

		if len(dest) >= 2 && dest[1] == ':' && isLetter(rune(dest[0])) {
			return state.ReplaceData(currentState)
		}

		if strings.HasPrefix(dest, "$") {
			return state.ReplaceData(currentState)
		}

		return state.
			ReplaceData(currentState).
			AddFailure(rule.Fail(r, line))
	}

	return state.ReplaceData(currentState)
}

func isLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}
