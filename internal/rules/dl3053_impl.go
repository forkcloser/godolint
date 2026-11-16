package rules

import (
	"fmt"
	"time"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3053Rule checks that RFC3339 timestamp labels are valid.
type DL3053Rule struct {
	cfg *config.Config
}

// DL3053 creates the rule for checking RFC3339 labels.
func DL3053() rule.Rule {
	return &DL3053Rule{
		cfg: config.Default(),
	}
}

// DL3053WithConfig creates the rule with custom configuration.
func DL3053WithConfig(cfg *config.Config) rule.Rule {
	return &DL3053Rule{
		cfg: cfg,
	}
}

// Code returns the rule code.
func (*DL3053Rule) Code() rule.Code {
	return DL3053Meta.Code
}

// Severity returns the rule severity.
func (*DL3053Rule) Severity() rule.Severity {
	return DL3053Meta.Severity
}

// Message returns the rule message.
func (*DL3053Rule) Message() string {
	return DL3053Meta.Message
}

// InitialState returns the initial state for this rule.
func (*DL3053Rule) InitialState() rule.State {
	return rule.EmptyState(nil)
}

// Check validates that RFC3339 timestamp labels are valid.
func (r *DL3053Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	label, ok := instruction.(*syntax.Label)
	if !ok {
		return state
	}

	ofKind := func(key string) bool { return r.cfg.LabelSchema[key] == config.LabelTypeRFC3339 }

	for _, key := range failingLabels(label.Pairs, ofKind, isValidRFC3339) {
		state = state.AddFailure(rule.CheckFailure{
			Code:     DL3053Meta.Code,
			Severity: DL3053Meta.Severity,
			//nolint:gocritic // sprintfQuotedString: hadolint prints the key raw; %#q would escape one holding a backquote
			Message: fmt.Sprintf("Label `%s` is not a valid time format - must conform to RFC3339.", key),
			Line:    line,
			Column:  1, // Hardcoded to 1 (matches hadolint)
		})
	}

	return state
}

// Finalize performs final checks after processing all instructions.
func (*DL3053Rule) Finalize(state rule.State) rule.State {
	return state
}

// isValidRFC3339 checks if a string is a valid RFC3339 timestamp.
func isValidRFC3339(timestamp string) bool {
	_, err := time.Parse(time.RFC3339, timestamp)

	return err == nil
}
