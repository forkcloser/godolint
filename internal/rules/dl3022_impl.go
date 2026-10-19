package rules

import (
	"maps"
	"strconv"
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// dl3022State tracks FROM stages to validate COPY --from references.
type dl3022State struct {
	count  int             // Number of FROM stages seen
	stages map[string]bool // Named stages (aliases)
}

// DL3022Rule checks that COPY --from references valid stages.
// Ported from Hadolint.Rule.DL3022.
type DL3022Rule struct{}

// DL3022 creates the rule for checking COPY --from references.
func DL3022() rule.Rule {
	return &DL3022Rule{}
}

// Code returns the rule code.
func (*DL3022Rule) Code() rule.Code {
	return DL3022Meta.Code
}

// Severity returns the rule severity.
func (*DL3022Rule) Severity() rule.Severity {
	return DL3022Meta.Severity
}

// Message returns the rule message.
func (*DL3022Rule) Message() string {
	return DL3022Meta.Message
}

// InitialState returns the initial state for this rule.
func (*DL3022Rule) InitialState() rule.State {
	return rule.EmptyState(dl3022State{
		count:  0,
		stages: make(map[string]bool),
	})
}

// Check tracks FROM stages and validates COPY --from references.
// Ported from the check function in DL3022.hs.
func (*DL3022Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	currentState := rule.Data[dl3022State](state)

	// Track FROM stages
	if from, ok := instruction.(*syntax.From); ok {
		newStages := make(map[string]bool)
		maps.Copy(newStages, currentState.stages)

		if from.Image.Alias != nil {
			newStages[*from.Image.Alias] = true
		}

		currentState.count++
		currentState.stages = newStages

		return state.ReplaceData(currentState)
	}

	// Check COPY --from references
	if copyInstr, ok := instruction.(*syntax.Copy); ok {
		if copyInstr.From == nil {
			return state
		}

		if currentState.resolves(*copyInstr.From) {
			return state
		}

		return state.AddFailure(rule.CheckFailure{
			Code:     DL3022Meta.Code,
			Severity: DL3022Meta.Severity,
			Message:  DL3022Meta.Message,
			Line:     line,
			Column:   1, // Hardcoded to 1 (matches hadolint)
		})
	}

	return state
}

// Finalize performs final checks after processing all instructions.
func (*DL3022Rule) Finalize(state rule.State) rule.State {
	return state // No finalization needed
}

// resolves reports whether a COPY --from names something the rule accepts:
// an image reference (it contains `:`), a variable (resolved at build time,
// not to a stage seen here), a stage alias seen so far, or a stage index
// below the count of stages seen.
func (s dl3022State) resolves(fromRef string) bool {
	if strings.Contains(fromRef, ":") || strings.Contains(fromRef, "$") || s.stages[fromRef] {
		return true
	}

	idx, err := strconv.Atoi(fromRef)

	return err == nil && idx < s.count
}
