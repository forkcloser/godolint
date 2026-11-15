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

// TestContractOnBuild: a rule reads the instruction an ONBUILD wraps when
// hadolint applies it there, and then apart from the stage the ONBUILD
// stands in; the other rules do not look inside an ONBUILD.
func TestContractOnBuild(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		dockerfile string
		code       string
		reported   bool
	}{
		// DL3059 (consecutive RUNs) and DL4000 (MAINTAINER) are not applied
		// inside ONBUILD; DL3008 (apt-get versions) is.
		"MAINTAINER as written":      {"FROM debian:12\nMAINTAINER x\n", "DL4000", true},
		"MAINTAINER inside ONBUILD":  {"FROM debian:12\nONBUILD MAINTAINER x\n", "DL4000", false},
		"an unpinned package":        {"FROM debian:12\nRUN apt-get install -y curl\n", "DL3008", true},
		"the same inside ONBUILD":    {"FROM debian:12\nONBUILD RUN apt-get install -y curl\n", "DL3008", true},
		"two RUNs as written":        {"FROM debian:12\nRUN a\nRUN b\n", "DL3059", true},
		"two RUNs inside ONBUILD":    {"FROM debian:12\nONBUILD RUN a\nONBUILD RUN b\n", "DL3059", false},
		"an ONBUILD RUN, then a RUN": {"FROM debian:12\nONBUILD RUN a\nRUN b\n", "DL3059", false},
		"apt lists left in a stage nothing is built on": {
			"FROM debian:12 AS build\nRUN apt-get update && apt-get install -y curl=1\nFROM scratch\n", "DL3009", false,
		},
		"the same RUN inside ONBUILD, which is of no stage": {
			"FROM debian:12 AS build\nONBUILD RUN apt-get update && apt-get install -y curl=1\nFROM scratch\n", "DL3009", true,
		},
	} {
		if got := slices.Contains(codes(lint(t, tc.dockerfile)), tc.code); got != tc.reported {
			t.Errorf("%s: %s reported %v, want %v", name, tc.code, got, tc.reported)
		}
	}
}

// TestContractRuleMessagesAreWhole: a rule's message is its whole
// description, as the README says of the metadata: never empty, and every
// backtick span it opens is closed. A message cut at the first literal of a
// joined one ("Label `", where hadolint goes on with the label's name) fails
// the second.
func TestContractRuleMessagesAreWhole(t *testing.T) {
	t.Parallel()

	for _, r := range sdk.AllRules() {
		msg := r.Message()
		if msg == "" || strings.Count(msg, "`")%2 != 0 {
			t.Errorf("%s: message %q is not a whole description", r.Code(), msg)
		}
	}
}

// TestContractRulesDescribeEveryRule: Rules describes each rule AllRules
// returns once, ordered by code, with that rule's severity and message; the slice is
// the caller's, so changing it changes no later answer; Rule finds a rule by
// its code as written and nothing else, shellcheck's codes included.
func TestContractRulesDescribeEveryRule(t *testing.T) {
	t.Parallel()

	infos := sdk.Rules()
	if len(infos) != len(sdk.AllRules()) {
		t.Fatalf("Rules describes %d rules, AllRules returns %d", len(infos), len(sdk.AllRules()))
	}

	severities := []sdk.Severity{
		sdk.SeverityError, sdk.SeverityWarning, sdk.SeverityInfo, sdk.SeverityStyle, sdk.SeverityIgnore,
	}

	declared := map[string]sdk.RuleInfo{}
	for _, r := range sdk.AllRules() {
		declared[string(r.Code())] = sdk.RuleInfo{
			Code: string(r.Code()), Severity: sdk.Severity(r.Severity().String()), Message: r.Message(),
		}
	}

	for i, info := range infos {
		if info != declared[info.Code] {
			t.Errorf("%s: described as %+v, declared as %+v", info.Code, info, declared[info.Code])
		}

		if len(info.Code) != 6 || !strings.HasPrefix(info.Code, "DL") {
			t.Errorf("code %q is not DL and four digits", info.Code)
		}

		if i > 0 && infos[i-1].Code >= info.Code {
			t.Errorf("%s comes after %s: not ordered by code, or twice", info.Code, infos[i-1].Code)
		}

		if info.Message == "" || !slices.Contains(severities, info.Severity) {
			t.Errorf("%s: severity %q, message %q", info.Code, info.Severity, info.Message)
		}

		if got, ok := sdk.Rule(info.Code); !ok || got != info {
			t.Errorf("Rule(%q) = %+v, %v; want %+v", info.Code, got, ok, info)
		}
	}

	first := infos[0]
	infos[0] = sdk.RuleInfo{Code: "changed"}

	if again := sdk.Rules(); again[0] != first {
		t.Errorf("after the caller changed its slice, Rules starts with %+v, want %+v", again[0], first)
	}

	for _, code := range []string{"", "dl3006", "DL30060", "SC2086"} {
		if got, ok := sdk.Rule(code); ok || got != (sdk.RuleInfo{}) {
			t.Errorf("Rule(%q) = %+v, %v; want no rule", code, got, ok)
		}
	}
}

// TestContractRuleInfoIsWhatAViolationCarries: a violation has its rule's
// severity and message; a rule about a label says `<label>` where its
// violation names the label, and any other name in angle brackets is the
// violation's text too; a rule at SeverityIgnore reports nothing a Result
// holds.
func TestContractRuleInfoIsWhatAViolationCarries(t *testing.T) {
	t.Parallel()

	r := lint(t, "FROM ubuntu\nMAINTAINER x\n")
	if len(r.Violations) == 0 {
		t.Fatal("no violation to compare")
	}

	for _, v := range r.Violations {
		info, ok := sdk.Rule(v.Code)
		if !ok || info.Severity != v.Severity || info.Message != v.Message {
			t.Errorf("violation %+v; its rule: %+v, %v", v, info, ok)
		}
	}

	if info, _ := sdk.Rule("DL3049"); info.Message != "Label `<label>` is missing." {
		t.Errorf("DL3049's message %q; want `<label>` in the label's place", info.Message)
	}

	// No HEALTHCHECK, which DL3057 is about, and DL3057 is off.
	if info, ok := sdk.Rule("DL3057"); !ok || info.Severity != sdk.SeverityIgnore {
		t.Fatalf("DL3057: %+v, %v; want SeverityIgnore", info, ok)
	}

	if slices.Contains(codes(r), "DL3057") {
		t.Errorf("a rule at SeverityIgnore reached the result: %v", codes(r))
	}
}
