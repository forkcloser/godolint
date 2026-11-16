package rules

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3054Rule checks that SPDX license labels are valid.
type DL3054Rule struct {
	cfg *config.Config
}

// DL3054 creates the rule for checking SPDX labels.
func DL3054() rule.Rule {
	return &DL3054Rule{
		cfg: config.Default(),
	}
}

// DL3054WithConfig creates the rule with custom configuration.
func DL3054WithConfig(cfg *config.Config) rule.Rule {
	return &DL3054Rule{
		cfg: cfg,
	}
}

// Code returns the rule code.
func (*DL3054Rule) Code() rule.Code {
	return DL3054Meta.Code
}

// Severity returns the rule severity.
func (*DL3054Rule) Severity() rule.Severity {
	return DL3054Meta.Severity
}

// Message returns the rule message.
func (*DL3054Rule) Message() string {
	return DL3054Meta.Message
}

// InitialState returns the initial state for this rule.
func (*DL3054Rule) InitialState() rule.State {
	return rule.EmptyState(nil)
}

// Check validates that SPDX license labels are valid.
func (r *DL3054Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	label, ok := instruction.(*syntax.Label)
	if !ok {
		return state
	}

	ofKind := func(key string) bool { return r.cfg.LabelSchema[key] == config.LabelTypeSPDX }

	for _, key := range failingLabels(label.Pairs, ofKind, isValidSPDX) {
		state = state.AddFailure(rule.CheckFailure{
			Code:     DL3054Meta.Code,
			Severity: DL3054Meta.Severity,
			//nolint:gocritic // sprintfQuotedString: hadolint prints the key raw; %#q would escape one holding a backquote
			Message: fmt.Sprintf("Label `%s` is not a valid SPDX identifier.", key),
			Line:    line,
			Column:  1, // Hardcoded to 1 (matches hadolint)
		})
	}

	return state
}

// Finalize performs final checks after processing all instructions.
func (*DL3054Rule) Finalize(state rule.State) rule.State {
	return state
}

// isValidSPDX reports whether license is one identifier of the SPDX License
// List, as it is written there and not deprecated; white space may follow
// it. An expression (MIT OR Apache-2.0) is not one identifier.
func isValidSPDX(license string) bool {
	_, listed := spdxLicenses[strings.TrimRightFunc(license, unicode.IsSpace)]

	return listed
}
