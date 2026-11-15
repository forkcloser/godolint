package testutils_test

import (
	"slices"
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

// onBuilds reports every ONBUILD it is shown.
type onBuilds struct{}

func (onBuilds) Code() rule.Code          { return "DL9998" }
func (onBuilds) Severity() rule.Severity  { return rule.Info }
func (onBuilds) Message() string          { return "An ONBUILD." }
func (onBuilds) InitialState() rule.State { return rule.EmptyState(nil) }

func (r onBuilds) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	if _, ok := instruction.(*syntax.OnBuild); ok {
		return state.AddFailure(rule.Fail(r, line))
	}

	return state
}

func (onBuilds) Finalize(state rule.State) rule.State { return state }

// TestLintDockerfileOnBuildWrapsEveryRun: every RUN is linted as an ONBUILD
// RUN, at its own line, and nothing else is; LintDockerfile wraps nothing.
func TestLintDockerfileOnBuildWrapsEveryRun(t *testing.T) {
	t.Parallel()

	const dockerfile = "FROM scratch\nRUN a\nCOPY b c\nRUN d\n"

	if plain := testutils.LintDockerfile(dockerfile, []rule.Rule{onBuilds{}}); len(plain) != 0 {
		t.Errorf("LintDockerfile shows an ONBUILD: %v", plain)
	}

	var lines []int
	for _, failure := range testutils.LintDockerfileOnBuild(dockerfile, []rule.Rule{onBuilds{}}) {
		lines = append(lines, failure.Line)
	}

	if !slices.Equal(lines, []int{2, 4}) {
		t.Errorf("ONBUILD at lines %v, want [2 4]", lines)
	}
}

// TestAssertions: the assertions fail on what hadolint's fail on, and on
// nothing else: AssertContainsViolation without the code or with a failure
// on no line; AssertViolationCount on another count; AssertOnLines on a
// line of 0 or less; AssertNoViolation on the code.
func TestAssertions(t *testing.T) {
	t.Parallel()

	at := func(code string, line int) rule.CheckFailure {
		return rule.CheckFailure{Code: rule.Code(code), Line: line}
	}

	one, two := []rule.CheckFailure{at("DL1", 1)}, []rule.CheckFailure{at("DL1", 1), at("DL2", 1)}

	for _, tc := range []struct {
		name      string
		assertion string
		failures  []rule.CheckFailure
		count     int
		fails     bool
	}{
		{"contains: there", "contains", one, 0, false},
		{"contains: another code", "contains", []rule.CheckFailure{at("DL2", 1)}, 0, true},
		{"contains: none", "contains", nil, 0, true},
		{"contains: there, on line 0", "contains", []rule.CheckFailure{at("DL1", 0)}, 0, true},
		{"contains: there, another on no line", "contains", []rule.CheckFailure{at("DL1", 1), at("DL2", -1)}, 0, true},
		{"count: as many", "count", []rule.CheckFailure{at("DL1", 1), at("DL2", 1), at("DL1", 2)}, 2, false},
		{"count: fewer", "count", two, 2, true},
		{"count: more", "count", []rule.CheckFailure{at("DL1", 1), at("DL1", 2), at("DL1", 3)}, 2, true},
		{"count: none wanted, none there", "count", nil, 0, false},
		{"count: none wanted, one there", "count", one, 0, true},
		{"lines: all on one", "lines", []rule.CheckFailure{at("DL1", 1), at("DL2", 7)}, 0, false},
		{"lines: one at 0", "lines", []rule.CheckFailure{at("DL1", 1), at("DL2", 0)}, 0, true},
		{"lines: one below 0", "lines", []rule.CheckFailure{at("DL1", -3)}, 0, true},
		{"none: another code", "none", []rule.CheckFailure{at("DL2", 1)}, 0, false},
		{"none: there", "none", one, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// A T of its own takes the assertion's verdict without failing
			// this test.
			probe := &testing.T{}

			switch tc.assertion {
			case "contains":
				testutils.AssertContainsViolation(probe, tc.failures, "DL1")
			case "count":
				testutils.AssertViolationCount(probe, tc.failures, "DL1", tc.count)
			case "lines":
				testutils.AssertOnLines(probe, tc.failures)
			case "none":
				testutils.AssertNoViolation(probe, tc.failures, "DL1")
			}

			if probe.Failed() != tc.fails {
				t.Errorf("failed: %v; want %v", probe.Failed(), tc.fails)
			}
		})
	}
}
