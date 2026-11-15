package rules

import (
	"fmt"
	"regexp"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3056Rule checks that a label the schema declares as a semantic version
// holds one. Ported from Hadolint.Rule.DL3056.
type DL3056Rule struct {
	cfg *config.Config
}

// DL3056 creates the rule with the default configuration.
func DL3056() rule.Rule {
	return &DL3056Rule{cfg: config.Default()}
}

// DL3056WithConfig creates the rule with custom configuration.
func DL3056WithConfig(cfg *config.Config) rule.Rule {
	return &DL3056Rule{cfg: cfg}
}

// Code returns the rule code.
func (*DL3056Rule) Code() rule.Code {
	return DL3056Meta.Code
}

// Severity returns the rule severity.
func (*DL3056Rule) Severity() rule.Severity {
	return DL3056Meta.Severity
}

// Message returns the rule message.
func (*DL3056Rule) Message() string {
	return DL3056Meta.Message
}

// InitialState returns the initial state for this rule.
func (*DL3056Rule) InitialState() rule.State {
	return rule.EmptyState(nil)
}

// Check fails a LABEL whose key the schema types as a semantic version and
// whose value is not one.
func (r *DL3056Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	label, ok := instruction.(*syntax.Label)
	if !ok {
		return state
	}

	ofKind := func(key string) bool { return r.cfg.LabelSchema[key] == config.LabelTypeSemVer }

	for _, key := range failingLabels(label.Pairs, ofKind, semanticVersion.MatchString) {
		state = state.AddFailure(rule.CheckFailure{
			Code:     DL3056Meta.Code,
			Severity: DL3056Meta.Severity,
			//nolint:gocritic // sprintfQuotedString: hadolint prints the key raw; %#q would escape one holding a backquote
			Message: fmt.Sprintf("Label `%s` does not conform to semantic versioning.", key),
			Line:    line,
			Column:  1, // Hardcoded to 1 (matches hadolint)
		})
	}

	return state
}

// Finalize performs no final check.
func (*DL3056Rule) Finalize(state rule.State) rule.State {
	return state
}

// semanticVersion is semver.org's grammar for a version: three numeric
// identifiers without leading zeros, an optional pre-release and an optional
// build, the whole string and nothing else.
var semanticVersion = regexp.MustCompile(
	`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)` +
		`(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?` +
		`(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`,
)
