package rules_test

import (
	"slices"
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3057Hand holds what hadolint's cases leave open about the stages a
// HEALTHCHECK makes good: those the stage is built on, by name, and a stage
// with no alias is named by its image, so two stages of one image are each
// built on the other and on themselves. The rule ends on every such shape.
func TestDL3057Hand(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{rules.DL3057()}

	for _, tc := range []struct {
		name       string
		dockerfile string
		lines      []int
	}{
		{"one stage, no HEALTHCHECK", "FROM debian\nRUN true", []int{1}},
		{"one stage with one", "FROM debian\nHEALTHCHECK CMD true", nil},
		{"two stages of one image, neither with one", "FROM debian\nRUN true\nFROM debian\nRUN true", []int{1, 3}},
		{"two stages of one image, the second with one", "FROM debian\nRUN true\nFROM debian\nHEALTHCHECK CMD true", nil},
		{"two stages of one image, the first with one", "FROM debian\nHEALTHCHECK CMD true\nFROM debian\nRUN true", nil},
		{
			"three stages of one image, the last with one",
			"FROM debian\nFROM debian\nFROM debian\nHEALTHCHECK CMD true",
			nil,
		},
		{
			"three stages of one image, the second with one",
			"FROM debian\nFROM debian\nHEALTHCHECK CMD true\nFROM debian",
			nil,
		},
		{"a stage aliased as its own image, then one of it", "FROM debian AS debian\nFROM debian\nHEALTHCHECK CMD true", nil},
		{
			"two stages each named as the other's image, then one built on them",
			"FROM one AS two\nFROM two AS one\nFROM one\nHEALTHCHECK CMD true",
			nil,
		},
		{
			"two images, the second with one",
			"FROM debian\nRUN true\nFROM alpine\nHEALTHCHECK CMD true",
			[]int{1},
		},
		{
			"a stage built on by its alias, the later one with one",
			"FROM debian AS base\nRUN true\nFROM base\nHEALTHCHECK CMD true",
			nil,
		},
		{
			"a chain of three by alias, the last with one, and a stage apart",
			"FROM debian AS one\nFROM one AS two\nFROM alpine AS apart\nFROM two\nHEALTHCHECK CMD true",
			[]int{3},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var lines []int
			for _, v := range testutils.LintDockerfile(tc.dockerfile, allRules) {
				lines = append(lines, v.Line)
			}

			slices.Sort(lines)

			if !slices.Equal(lines, tc.lines) {
				t.Errorf("reported at lines %v, want %v", lines, tc.lines)
			}
		})
	}
}
