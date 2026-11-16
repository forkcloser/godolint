package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestVersionVariablesOfAnAliasTakenTwice: when two stages go by one alias,
// a stage built on that alias has the variables of the last, the one a build
// takes the name for, and has them on every run.
func TestVersionVariablesOfAnAliasTakenTwice(t *testing.T) {
	t.Parallel()

	const stages = "FROM a:1 AS base\nENV FIRST_VERSION=1.0\nFROM a:1 AS base\nENV SECOND_VERSION=1.0\nFROM base\n"

	for _, tc := range []struct {
		name     string
		rule     func() rule.Rule
		run      string
		reported bool
	}{
		{"yum, the first stage's variable", rules.DL3033, "RUN yum install -y foo-${FIRST_VERSION} && yum clean all\n", true},
		{"yum, the second stage's variable", rules.DL3033, "RUN yum install -y foo-${SECOND_VERSION} && yum clean all\n", false},
		{"dnf, the first stage's variable", rules.DL3041, "RUN dnf install -y foo-${FIRST_VERSION} && dnf clean all\n", true},
		{"dnf, the second stage's variable", rules.DL3041, "RUN dnf install -y foo-${SECOND_VERSION} && dnf clean all\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Enough runs for an order left to a map to show.
			for range 200 {
				violations := testutils.LintDockerfile(stages+tc.run, []rule.Rule{tc.rule()})
				if reported := len(violations) != 0; reported != tc.reported {
					t.Fatalf("reported %v, want %v: %v", reported, tc.reported, violations)
				}
			}
		})
	}
}
