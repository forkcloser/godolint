package rules

import (
	"maps"
	"slices"
	"strings"
	"unicode"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// dl3066State is the ARG and ENV names declared so far (a user id written as
// one is resolved at build time, not a name the host has to know), the line
// of the current stage's FROM, and the line of each stage's last unresolved
// USER, keyed by the stage's FROM line (0 before any FROM). Only the last
// USER of a stage reaches the host, and only the last stage's: a FROM forgets
// the stage before it, as hadolint's rule does.
type dl3066State struct {
	args    map[string]bool
	stage   int
	pending map[int]int
}

// DL3066Rule checks that the user a stage ends on names a numeric id.
// Ported from Hadolint.Rule.DL3066.
type DL3066Rule struct{}

// DL3066 creates the rule that a user id is numeric, or a declared ARG or ENV.
func DL3066() rule.Rule {
	return &DL3066Rule{}
}

// Code returns the rule code.
func (*DL3066Rule) Code() rule.Code {
	return DL3066Meta.Code
}

// Severity returns the rule severity.
func (*DL3066Rule) Severity() rule.Severity {
	return DL3066Meta.Severity
}

// Message returns the rule message.
func (*DL3066Rule) Message() string {
	return DL3066Meta.Message
}

// InitialState returns the initial state for this rule.
func (*DL3066Rule) InitialState() rule.State {
	return rule.EmptyState(dl3066State{args: map[string]bool{}, pending: map[int]int{}})
}

// Check registers each ARG and ENV name, starts a stage at each FROM
// (forgetting the previous stage's unresolved USER), and records a USER
// whose id (the part before any `:`) is neither all digits nor a reference
// to a registered name, or clears the record when one is.
func (*DL3066Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	current := rule.Data[dl3066State](state)

	switch instr := instruction.(type) {
	case *syntax.From:
		pending := maps.Clone(current.pending)
		delete(pending, current.stage)

		return state.ReplaceData(dl3066State{args: current.args, stage: line, pending: pending})
	case *syntax.Arg:
		return state.ReplaceData(current.withNames(instr.ArgName))
	case *syntax.Env:
		names := make([]string, 0, len(instr.Pairs))
		for _, pair := range instr.Pairs {
			names = append(names, pair.Key)
		}

		return state.ReplaceData(current.withNames(names...))
	case *syntax.User:
		uid, _, _ := strings.Cut(instr.User, ":")
		pending := maps.Clone(current.pending)

		if allDigits(uid) || uidIsDefinedArg(current, uid) {
			delete(pending, current.stage)
		} else {
			pending[current.stage] = line
		}

		return state.ReplaceData(dl3066State{args: current.args, stage: current.stage, pending: pending})
	default:
		return state
	}
}

// Finalize fails the unresolved USER each remaining stage ended on, in stage
// order.
func (*DL3066Rule) Finalize(state rule.State) rule.State {
	current := rule.Data[dl3066State](state)

	for _, stage := range slices.Sorted(maps.Keys(current.pending)) {
		state = state.AddFailure(rule.CheckFailure{
			Code:     DL3066Meta.Code,
			Severity: DL3066Meta.Severity,
			Message:  DL3066Meta.Message,
			Line:     current.pending[stage],
			Column:   1,
		})
	}

	return state
}

// withNames is the state with names registered.
func (s dl3066State) withNames(names ...string) dl3066State {
	args := maps.Clone(s.args)
	for _, name := range names {
		args[name] = true
	}

	return dl3066State{args: args, stage: s.stage, pending: s.pending}
}

// allDigits is hadolint's `Text.all Char.isDigit`: true of the empty string.
func allDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}

	return true
}

// uidIsDefinedArg reports whether uid refers to a registered name, as
// `${NAME}` or `$NAME`.
func uidIsDefinedArg(current dl3066State, uid string) bool {
	for arg := range current.args {
		if strings.Contains(uid, "${"+arg+"}") || strings.Contains(uid, "$"+arg) {
			return true
		}
	}

	return false
}
