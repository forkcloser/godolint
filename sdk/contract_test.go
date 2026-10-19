// The SDK's contract from the outside: what the README promises of Lint,
// the pragmas and the options.
package sdk_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/forkcloser/godolint/sdk"
)

func codes(r *sdk.Result) []string {
	out := make([]string, 0, len(r.Violations))
	for _, v := range r.Violations {
		out = append(out, v.Code)
	}

	slices.Sort(out)

	return slices.Compact(out)
}

func lint(t *testing.T, dockerfile string, opts ...sdk.Option) *sdk.Result {
	t.Helper()

	r, err := sdk.New(opts...).Lint(t.Context(), []byte(dockerfile))
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}

	return r
}

// TestContractIgnorePragmas: `# hadolint ignore=` covers the next line,
// `stage ignore=` above a FROM the stage it starts, `global ignore=` the
// whole file; spaces around = and the commas and a trailing # comment are
// fine; a list with a name that is not a rule code is not a pragma at all;
// an ignore pragma is itself reported by DL1001 at the ignore severity,
// which never reaches the result.
func TestContractIgnorePragmas(t *testing.T) {
	t.Parallel()

	const untagged = "FROM ubuntu\n"

	if got := codes(lint(t, untagged)); !slices.Contains(got, "DL3006") {
		t.Fatalf("the premise: an untagged FROM reports DL3006, got %v", got)
	}

	for name, c := range map[string]struct {
		dockerfile string
		absent     []string
		present    []string
	}{
		"ignore the next line":       {"# hadolint ignore=DL3006\nFROM ubuntu\n", []string{"DL3006"}, nil},
		"ignore two codes":           {"# hadolint ignore=DL3006,DL3007\nFROM ubuntu:latest\n", []string{"DL3006", "DL3007"}, nil},
		"spaces and a comment":       {"#  hadolint  ignore = DL3006 , DL3007  # why\nFROM ubuntu:latest\n", []string{"DL3006", "DL3007"}, nil},
		"ignore covers one line":     {"# hadolint ignore=DL3006\nFROM debian:bookworm\nFROM ubuntu\n", nil, []string{"DL3006"}},
		"stage ignore":               {"# hadolint stage ignore=DL3002\nFROM alpine:3.7\nUSER root\nFROM debian:bookworm\nUSER root\n", nil, []string{"DL3002"}},
		"global ignore":              {"# hadolint global ignore=DL3006\nFROM ubuntu\nFROM debian\n", []string{"DL3006"}, nil},
		"a name that is not a code":  {"# hadolint ignore=crazy,DL3006\nFROM ubuntu\n", nil, []string{"DL3006"}},
		"DL1001 never in the result": {"# hadolint ignore=DL3006\nFROM ubuntu\n", []string{"DL1001"}, nil},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := codes(lint(t, c.dockerfile))

			for _, a := range c.absent {
				if slices.Contains(got, a) {
					t.Errorf("%s reported: %v", a, got)
				}
			}

			for _, p := range c.present {
				if !slices.Contains(got, p) {
					t.Errorf("%s not reported: %v", p, got)
				}
			}
		})
	}

	// stage ignore covers its stage alone: DL3002 (last USER root) once,
	// from the second stage.
	r := lint(t, "# hadolint stage ignore=DL3002\nFROM alpine:3.7\nUSER root\nFROM debian:bookworm\nUSER root\n")

	var lines []int

	for _, v := range r.Violations {
		if v.Code == "DL3002" {
			lines = append(lines, v.Line)
		}
	}

	if !slices.Equal(lines, []int{5}) {
		t.Errorf("stage ignore: DL3002 at lines %v, want [5]", lines)
	}
}

// TestContractViolationsCarryLineSeverityMessage: a violation names its
// rule, the severity the rule declares, a 1-based line of the instruction
// and the rule's message; Passed is the absence of violations; a clean
// file passes.
func TestContractViolationsCarryLineSeverityMessage(t *testing.T) {
	t.Parallel()

	r := lint(t, "FROM debian:bookworm\n\nFROM ubuntu\nMAINTAINER x\n")

	var untagged, maintainer *sdk.Violation

	for i := range r.Violations {
		switch r.Violations[i].Code {
		case "DL3006":
			untagged = &r.Violations[i]
		case "DL4000":
			maintainer = &r.Violations[i]
		}
	}

	if untagged == nil || untagged.Line != 3 || untagged.Severity != sdk.SeverityWarning || untagged.Message == "" {
		t.Errorf("DL3006: %+v, want line 3, warning, a message", untagged)
	}

	if maintainer == nil || maintainer.Line != 4 || maintainer.Severity != sdk.SeverityError {
		t.Errorf("DL4000: %+v, want line 4, error", maintainer)
	}

	if r.Passed {
		t.Error("Passed with violations")
	}

	if clean := lint(t, "FROM debian:bookworm\nCMD [\"true\"]\n"); !clean.Passed || len(clean.Violations) != 0 {
		t.Errorf("a clean file: Passed %v, %v", clean.Passed, clean.Violations)
	}
}

// TestContractOptionsSelectRules: WithDisabledRules drops the named codes;
// WithRuleSet picks a set, recommended being the error and warning rules
// and any name no set has being every rule; AllRules lists every rule
// once.
func TestContractOptionsSelectRules(t *testing.T) {
	t.Parallel()

	const file = "FROM ubuntu\nMAINTAINER x\n"

	if got := codes(
		lint(t, file, sdk.WithDisabledRules("DL3006")),
	); slices.Contains(got, "DL3006") ||
		!slices.Contains(got, "DL4000") {
		t.Errorf("WithDisabledRules(DL3006): %v", got)
	}

	for _, set := range []sdk.RuleSet{sdk.RuleSetAll, sdk.RuleSetRecommended, sdk.RuleSetStrict} {
		if len(sdk.GetRuleSet(set)) == 0 {
			t.Errorf("rule set %q is empty", set)
		}
	}

	if got := codes(lint(t, file, sdk.WithRuleSet(sdk.RuleSet("nonesuch")))); !slices.Contains(got, "DL3006") {
		t.Errorf("an unknown rule set, every rule: %v", got)
	}

	for _, r := range sdk.GetRuleSet(sdk.RuleSetRecommended) {
		if sev := r.Severity().String(); sev != "error" && sev != "warning" {
			t.Errorf("recommended holds %s at %s", r.Code(), sev)
		}
	}

	seen := map[string]bool{}

	for _, r := range sdk.AllRules() {
		code := string(r.Code())
		if seen[code] {
			t.Errorf("rule %s listed twice", code)
		}

		seen[code] = true

		if !strings.HasPrefix(code, "DL") {
			t.Errorf("rule %s is not a DL code", code)
		}
	}
}

// TestContractParseErrorAndCancellation: a file with no instruction
// (empty, blank, comments alone) is a ParseError wrapping the parser's,
// and no result; a cancelled context is context.Canceled.
func TestContractParseErrorAndCancellation(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"", "  \n\n", "# a comment alone\n"} {
		r, err := sdk.New().Lint(t.Context(), []byte(in))

		var perr *sdk.ParseError
		if !errors.As(err, &perr) || perr.Unwrap() == nil || r != nil {
			t.Errorf("%q: %v (result %v), want a ParseError with a cause and no result", in, err, r)
		}
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if _, err := sdk.New().Lint(ctx, []byte("FROM debian\n")); !errors.Is(err, context.Canceled) {
		t.Errorf("a cancelled context: %v", err)
	}
}
