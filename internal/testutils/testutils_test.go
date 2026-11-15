package testutils_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
	"github.com/forkcloser/godolint/internal/testutils"
)

// reporting is a rule that declares one thing and reports, on every FROM,
// the failure it is given.
type reporting struct {
	message string
	failure rule.CheckFailure
}

func (reporting) Code() rule.Code         { return "DL9999" }
func (reporting) Severity() rule.Severity { return rule.Warning }
func (r reporting) Message() string       { return r.message }
func (reporting) InitialState() rule.State {
	return rule.EmptyState(nil)
}

func (r reporting) Check(_ int, state rule.State, instruction syntax.Instruction) rule.State {
	if _, ok := instruction.(*syntax.From); ok {
		return state.AddFailure(r.failure)
	}

	return state
}

func (reporting) Finalize(state rule.State) rule.State { return state }

// TestLintDockerfileHoldsFailuresToTheirRule: a failure under a rule's code
// carries that rule's severity and message, the label in the place of
// `<label>`, or the helper panics; a failure under another code is not the
// rule's to answer for.
func TestLintDockerfileHoldsFailuresToTheirRule(t *testing.T) {
	t.Parallel()

	const plain, labelled = "Do not do this.", "Label `<label>` is missing."

	tests := []struct {
		name     string
		declared string
		failure  rule.CheckFailure
		refused  bool
	}{
		{"the rule's own", plain, rule.CheckFailure{Code: "DL9999", Severity: rule.Warning, Message: plain}, false},
		{"another severity", plain, rule.CheckFailure{Code: "DL9999", Severity: rule.Error, Message: plain}, true},
		{
			"a longer message",
			plain,
			rule.CheckFailure{Code: "DL9999", Severity: rule.Warning, Message: plain + " Ever."},
			true,
		},
		{"an empty message", plain, rule.CheckFailure{Code: "DL9999", Severity: rule.Warning, Message: ""}, true},
		{
			"the label in its place", labelled,
			rule.CheckFailure{Code: "DL9999", Severity: rule.Warning, Message: "Label `maintainer` is missing."},
			false,
		},
		{
			"the label and another end", labelled,
			rule.CheckFailure{Code: "DL9999", Severity: rule.Warning, Message: "Label `maintainer` is empty."},
			true,
		},
		{
			"the label and another start", labelled,
			rule.CheckFailure{Code: "DL9999", Severity: rule.Warning, Message: "label `maintainer` is missing."},
			true,
		},
		{
			"the start and the end sharing their bytes", "a`<label>`a",
			rule.CheckFailure{Code: "DL9999", Severity: rule.Warning, Message: "a`a"},
			true,
		},
		{"another code", plain, rule.CheckFailure{Code: "SC2086", Severity: rule.Info, Message: "Quote this."}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				if refused := recover() != nil; refused != tt.refused {
					t.Errorf("refused: %v; want %v", refused, tt.refused)
				}
			}()

			failures := testutils.LintDockerfile(
				"FROM scratch\n", []rule.Rule{reporting{message: tt.declared, failure: tt.failure}},
			)
			if len(failures) != 1 {
				t.Errorf("%d failures; want the one reported", len(failures))
			}
		})
	}
}
