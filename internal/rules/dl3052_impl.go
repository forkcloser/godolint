package rules

import (
	"fmt"
	"net/url"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3052Rule checks that URL labels are valid.
type DL3052Rule struct {
	cfg *config.Config
}

// DL3052 creates the rule for checking URL labels.
func DL3052() rule.Rule {
	return &DL3052Rule{
		cfg: config.Default(),
	}
}

// DL3052WithConfig creates the rule with custom configuration.
func DL3052WithConfig(cfg *config.Config) rule.Rule {
	return &DL3052Rule{
		cfg: cfg,
	}
}

// Code returns the rule code.
func (*DL3052Rule) Code() rule.Code {
	return DL3052Meta.Code
}

// Severity returns the rule severity.
func (*DL3052Rule) Severity() rule.Severity {
	return DL3052Meta.Severity
}

// Message returns the rule message.
func (*DL3052Rule) Message() string {
	return DL3052Meta.Message
}

// InitialState returns the initial state for this rule.
func (*DL3052Rule) InitialState() rule.State {
	return rule.EmptyState(nil)
}

// Check validates that URL label values are valid URLs.
func (r *DL3052Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	label, ok := instruction.(*syntax.Label)
	if !ok {
		return state
	}

	ofKind := func(key string) bool { return r.cfg.LabelSchema[key] == config.LabelTypeURL }

	for _, key := range failingLabels(label.Pairs, ofKind, isValidURL) {
		state = state.AddFailure(rule.CheckFailure{
			Code:     DL3052Meta.Code,
			Severity: DL3052Meta.Severity,
			//nolint:gocritic // sprintfQuotedString: hadolint prints the key raw; %#q would escape one holding a backquote
			Message: fmt.Sprintf("Label `%s` is not a valid URL.", key),
			Line:    line,
			Column:  1, // Hardcoded to 1 (matches hadolint)
		})
	}

	return state
}

// Finalize performs final checks after processing all instructions.
func (*DL3052Rule) Finalize(state rule.State) rule.State {
	return state
}

// isValidURL checks if a string is a valid URL.
func isValidURL(urlStr string) bool {
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return false
	}

	// Must have a scheme (http, https, etc.)
	return parsedURL.Scheme != ""
}
