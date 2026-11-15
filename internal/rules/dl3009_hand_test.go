package rules_test

import (
	"slices"
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3009Hand holds hadolint's "ok with any form of recursive removal of
// the apt lists" cases, which the generator does not port: the spec maps one
// assertion over a list of Dockerfiles.
func TestDL3009Hand(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{rules.DL3009()}

	for _, tc := range []struct {
		name      string
		cleanup   string
		violation bool
	}{
		{"flags in the other order", `rm -fr /var/lib/apt/lists/*`, false},
		{"flags apart", `rm -r -f /var/lib/apt/lists/*`, false},
		{"upper-case recursive", `rm -Rf /var/lib/apt/lists/*`, false},
		{"long flags", `rm --recursive --force /var/lib/apt/lists/*`, false},
		{"the directory itself", `rm -rf /var/lib/apt/lists`, false},
		{"the directory with a trailing slash", `rm -rf /var/lib/apt/lists/`, false},
		{"another directory", `rm -rf /var/lib/apt`, true},
		{"force without recursive", `rm -f /var/lib/apt/lists/*`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dockerfile := "RUN apt-get update && apt-get install python && " + tc.cleanup

			violations := testutils.LintDockerfile(dockerfile, allRules)
			if tc.violation {
				testutils.AssertContainsViolation(t, violations, "DL3009")
			} else {
				testutils.AssertNoViolation(t, violations, "DL3009")
			}
		})
	}
}

// TestDL3009Stages holds what hadolint's cases leave open about the stages:
// a stage with no alias that nothing can be built on, two stages written
// alike, which are one to hadolint, a RUN before any FROM, and the order of
// the failures.
func TestDL3009Stages(t *testing.T) {
	t.Parallel()

	const apt = "RUN apt-get update && apt-get install python"

	allRules := []rule.Rule{rules.DL3009()}

	for _, tc := range []struct {
		name       string
		dockerfile string
		lines      []int
	}{
		{"no FROM at all", apt, []int{1}},
		{"a RUN before the first FROM", apt + "\nFROM debian\nRUN true", nil},
		{"a RUN before the first FROM, the last stage FROM scratch", apt + "\nFROM scratch\nRUN true", []int{1}},
		{
			"an image behind a registry is not the stage of that name",
			"FROM ubuntu AS base\n" + apt + "\nFROM reg.io/base\nRUN true", nil,
		},
		{"a stage with no alias, then another image", "FROM ubuntu\n" + apt + "\nFROM scratch\nRUN true", nil},
		{"two stages written alike", "FROM ubuntu\n" + apt + "\nFROM ubuntu\nRUN true", []int{2}},
		{"the same image under another alias", "FROM ubuntu AS one\n" + apt + "\nFROM ubuntu AS two\nRUN true", nil},
		{"a stage built on by its alias", "FROM ubuntu AS base\n" + apt + "\nFROM base\nRUN true", []int{2}},
		{
			"two RUNs of the last stage, in their order",
			"FROM ubuntu\n" + apt + "\nRUN true\n" + apt,
			[]int{2, 4},
		},
		{
			"an earlier stage built on and the last one",
			"FROM ubuntu AS base\n" + apt + "\nFROM base\n" + apt,
			[]int{2, 4},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var lines []int
			for _, v := range testutils.LintDockerfile(tc.dockerfile, allRules) {
				lines = append(lines, v.Line)
			}

			if !slices.Equal(lines, tc.lines) {
				t.Errorf("reported at lines %v, want %v", lines, tc.lines)
			}
		})
	}
}
