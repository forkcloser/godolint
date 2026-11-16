package rules

import (
	"maps"
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// dl3002State tracks root USER instructions per stage.
// Ported from Acc in DL3002.hs.
type dl3002State struct {
	currentStage int         // line number of current FROM
	rootUsers    map[int]int // map[stageLine]userLine - tracks last root USER per stage
}

// DL3002Rule checks that the last USER in each stage is not root.
// Ported from Hadolint.Rule.DL3002.
type DL3002Rule struct {
	rule.StatefulRuleBase
}

// DL3002 creates the rule for checking last USER should not be root.
func DL3002() rule.Rule {
	return &DL3002Rule{StatefulRuleBase: rule.NewStatefulRuleBase(DL3002Meta)}
}

// InitialState returns the initial state for this rule.
func (*DL3002Rule) InitialState() rule.State {
	return rule.EmptyState(dl3002State{
		currentStage: -1,
		rootUsers:    make(map[int]int),
	})
}

// Check tracks USER instructions and remembers which stages end with root.
// Ported from the check function in DL3002.hs.
func (*DL3002Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	currentState := rule.Data[dl3002State](state)

	if _, ok := instruction.(*syntax.From); ok {
		currentState.currentStage = line

		return state.ReplaceData(currentState)
	}

	if user, ok := instruction.(*syntax.User); ok {
		if isRoot(user.User) {
			newRootUsers := make(map[int]int)
			maps.Copy(newRootUsers, currentState.rootUsers)

			newRootUsers[currentState.currentStage] = line
			currentState.rootUsers = newRootUsers

			return state.ReplaceData(currentState)
		}

		newRootUsers := make(map[int]int)

		for k, v := range currentState.rootUsers {
			if k != currentState.currentStage {
				newRootUsers[k] = v
			}
		}

		currentState.rootUsers = newRootUsers

		return state.ReplaceData(currentState)
	}

	return state
}

// Finalize adds failures for all stages that end with a root USER.
// Ported from markFailures in DL3002.hs.
// Finalize performs final checks after processing all instructions.
func (r *DL3002Rule) Finalize(state rule.State) rule.State {
	currentState := rule.Data[dl3002State](state)

	finalState := state
	for _, userLine := range currentState.rootUsers {
		finalState = finalState.AddFailure(rule.Fail(r, userLine))
	}

	return finalState
}

// isRoot checks if a USER instruction specifies root.
// Ported from isRoot in DL3002.hs.
func isRoot(user string) bool {
	return strings.HasPrefix(user, "root:") ||
		strings.HasPrefix(user, "0:") ||
		user == "root" ||
		user == "0"
}
