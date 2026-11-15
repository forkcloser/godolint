package rules

import (
	"fmt"
	"net/mail"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3058Rule checks that email labels are valid.
type DL3058Rule struct {
	cfg *config.Config
}

// DL3058 creates the rule for checking email labels.
func DL3058() rule.Rule {
	return &DL3058Rule{
		cfg: config.Default(),
	}
}

// DL3058WithConfig creates the rule with custom configuration.
func DL3058WithConfig(cfg *config.Config) rule.Rule {
	return &DL3058Rule{
		cfg: cfg,
	}
}

// Code returns the rule code.
func (*DL3058Rule) Code() rule.Code {
	return DL3058Meta.Code
}

// Severity returns the rule severity.
func (*DL3058Rule) Severity() rule.Severity {
	return DL3058Meta.Severity
}

// Message returns the rule message.
func (*DL3058Rule) Message() string {
	return DL3058Meta.Message
}

// InitialState returns the initial state for this rule.
func (*DL3058Rule) InitialState() rule.State {
	return rule.EmptyState(nil)
}

// Check validates that email labels are valid email addresses.
func (r *DL3058Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	label, ok := instruction.(*syntax.Label)
	if !ok {
		return state
	}

	ofKind := func(key string) bool { return r.cfg.LabelSchema[key] == config.LabelTypeEmail }

	for _, key := range failingLabels(label.Pairs, ofKind, isValidEmail) {
		state = state.AddFailure(rule.CheckFailure{
			Code:     DL3058Meta.Code,
			Severity: DL3058Meta.Severity,
			//nolint:gocritic // sprintfQuotedString: hadolint prints the key raw; %#q would escape one holding a backquote
			Message: fmt.Sprintf("Label `%s` is not a valid email format - must conform to RFC5322.", key),
			Line:    line,
			Column:  1, // Hardcoded to 1 (matches hadolint)
		})
	}

	return state
}

// Finalize performs final checks after processing all instructions.
func (*DL3058Rule) Finalize(state rule.State) rule.State {
	return state
}

// isValidEmail reports whether email is an address and nothing more:
// john@example.com. A name with the address (John Doe <john@example.com>),
// or the address in angle brackets alone, is a mailbox, which the parser
// takes and the label's type does not.
func isValidEmail(email string) bool {
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed == nil {
		return false
	}

	return parsed.Name == "" && parsed.Address == email
}
