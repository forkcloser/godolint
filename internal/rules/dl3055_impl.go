package rules

import (
	"fmt"
	"regexp"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3055Rule checks that git hash labels are valid.
type DL3055Rule struct {
	cfg *config.Config
}

// DL3055 creates the rule for checking git hash labels.
func DL3055() rule.Rule {
	return &DL3055Rule{
		cfg: config.Default(),
	}
}

// DL3055WithConfig creates the rule with custom configuration.
func DL3055WithConfig(cfg *config.Config) rule.Rule {
	return &DL3055Rule{
		cfg: cfg,
	}
}

// Code returns the rule code.
func (*DL3055Rule) Code() rule.Code {
	return DL3055Meta.Code
}

// Severity returns the rule severity.
func (*DL3055Rule) Severity() rule.Severity {
	return DL3055Meta.Severity
}

// Message returns the rule message.
func (*DL3055Rule) Message() string {
	return DL3055Meta.Message
}

// InitialState returns the initial state for this rule.
func (*DL3055Rule) InitialState() rule.State {
	return rule.EmptyState(nil)
}

// Check validates that git hash labels are valid.
func (r *DL3055Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	label, ok := instruction.(*syntax.Label)
	if !ok {
		return state
	}

	ofKind := func(key string) bool { return r.cfg.LabelSchema[key] == config.LabelTypeGitHash }

	for _, key := range failingLabels(label.Pairs, ofKind, isValidGitHash) {
		state = state.AddFailure(rule.CheckFailure{
			Code:     DL3055Meta.Code,
			Severity: DL3055Meta.Severity,
			//nolint:gocritic // sprintfQuotedString: hadolint prints the key raw; %#q would escape one holding a backquote
			Message: fmt.Sprintf("Label `%s` is not a valid git hash.", key),
			Line:    line,
			Column:  1, // Hardcoded to 1 (matches hadolint)
		})
	}

	return state
}

// Finalize performs final checks after processing all instructions.
func (*DL3055Rule) Finalize(state rule.State) rule.State {
	return state
}

// Git hash lengths: abbreviated (short) and full SHA-1.
const (
	shortGitHashLength = 7
	fullGitHashLength  = 40
)

// isValidGitHash checks if a string is a valid git hash.
// Valid git hash: 7 or 40 hexadecimal characters.
func isValidGitHash(hash string) bool {
	if len(hash) != shortGitHashLength && len(hash) != fullGitHashLength {
		return false
	}

	// Check if all characters are hexadecimal
	hexPattern := regexp.MustCompile(`^[0-9a-fA-F]+$`)

	return hexPattern.MatchString(hash)
}
