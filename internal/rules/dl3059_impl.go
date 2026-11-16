package rules

import (
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3059 creates a rule for checking multiple consecutive RUN instructions.
func DL3059() rule.Rule {
	return &DL3059Rule{StatefulRuleBase: rule.NewStatefulRuleBase(DL3059Meta)}
}

// DL3059Rule checks for multiple consecutive RUN instructions.
// Ported from Hadolint.Rule.DL3059.
type DL3059Rule struct {
	rule.StatefulRuleBase
}

// dl3059State tracks the previous RUN instruction state.
type dl3059State struct {
	Seen  bool     // A RUN came before, with nothing but comments since
	Flags []string // Flags from previous RUN
	Count int      // Number of commands in previous RUN
}

// InitialState returns the initial state for this rule.
func (*DL3059Rule) InitialState() rule.State {
	return rule.EmptyState(dl3059State{})
}

// Check detects multiple consecutive RUN instructions.
func (r *DL3059Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	prevState := rule.Data[dl3059State](state)

	switch instr := instruction.(type) {
	case *syntax.Run:
		// Parse the shell command to count commands
		parsed, err := shell.ParseShell(instr.Command)
		if err != nil {
			// If parsing fails, reset state and continue
			return state.ReplaceData(dl3059State{})
		}

		commandCount := shell.CountCommands(parsed)
		currentFlags := instr.Flags

		// Check if we should fail:
		// 1. Previous state exists (not first RUN)
		// 2. Flags are the same
		// 3. NEITHER current NOR previous have chained commands (2+)
		// Per hadolint: "ignores the case where multiple commands are chained together"
		if prevState.Seen && // Previous RUN exists, with a command in it or not
			flagsEqual(prevState.Flags, currentFlags) &&
			commandCount < 2 && // Current is not chained
			prevState.Count < 2 { // Previous is not chained
			// Fail: consecutive simple RUN instructions with same flags
			return state.AddFailure(rule.Fail(r, line)).ReplaceData(dl3059State{
				Seen:  true,
				Flags: currentFlags,
				Count: commandCount,
			})
		}

		// Update state with current RUN
		return state.ReplaceData(dl3059State{
			Seen:  true,
			Flags: currentFlags,
			Count: commandCount,
		})

	case *syntax.Comment:
		// Comments don't break consecutive RUN detection
		return state

	default:
		// Any other instruction resets the state
		return state.ReplaceData(dl3059State{})
	}
}

// flagsEqual compares two flag slices for equality.
func flagsEqual(flags1, flags2 []string) bool {
	if len(flags1) != len(flags2) {
		return false
	}

	// Simple comparison - flags should be in same order
	for i := range flags1 {
		if flags1[i] != flags2[i] {
			return false
		}
	}

	return true
}
