package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL1001 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL1001Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL1001(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL1001(),
	}

	t.Run(
		"catches inline ignore pragma",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `# hadolint ignore=DL3003
RUN foo bar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL1001")
		},
	)

	t.Run(
		"does not catch other comment",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `# foobar
RUN foo bar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL1001")
		},
	)

	t.Run(
		"does not catch other pragma",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `# hadolint shell=powershell
RUN foo bar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL1001")
		},
	)

	t.Run(
		"does not catch when inline ignore pragma is absent",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM debian
RUN foo bar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL1001")
		},
	)
}
