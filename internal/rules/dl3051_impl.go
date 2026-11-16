package rules

import (
	"fmt"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3051Rule checks that labels are not empty.
type DL3051Rule struct {
	cfg *config.Config
}

// DL3051 creates the rule for checking empty labels.
func DL3051() rule.Rule {
	return &DL3051Rule{
		cfg: config.Default(),
	}
}

// DL3051WithConfig creates the rule with custom configuration.
func DL3051WithConfig(cfg *config.Config) rule.Rule {
	return &DL3051Rule{
		cfg: cfg,
	}
}

// Code returns the rule code.
func (*DL3051Rule) Code() rule.Code {
	return DL3051Meta.Code
}

// Severity returns the rule severity.
func (*DL3051Rule) Severity() rule.Severity {
	return DL3051Meta.Severity
}

// Message returns the rule message.
func (*DL3051Rule) Message() string {
	return DL3051Meta.Message
}

// InitialState returns the initial state for this rule.
func (*DL3051Rule) InitialState() rule.State {
	return rule.EmptyState(nil)
}

// Check validates that label values are not empty.
func (r *DL3051Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	label, ok := instruction.(*syntax.Label)
	if !ok {
		return state
	}

	inSchema := func(key string) bool {
		_, exists := r.cfg.LabelSchema[key]

		return exists
	}

	for _, key := range failingLabels(label.Pairs, inSchema, func(value string) bool { return value != "" }) {
		state = state.AddFailure(rule.CheckFailure{
			Code:     DL3051Meta.Code,
			Severity: DL3051Meta.Severity,
			//nolint:gocritic // sprintfQuotedString: hadolint prints the key raw; %#q would escape one holding a backquote
			Message: fmt.Sprintf("label `%s` is empty.", key),
			Line:    line,
			Column:  1, // Hardcoded to 1 (matches hadolint)
		})
	}

	return state
}

// Finalize performs final checks after processing all instructions.
func (*DL3051Rule) Finalize(state rule.State) rule.State {
	return state
}
