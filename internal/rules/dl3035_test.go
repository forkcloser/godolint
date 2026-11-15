package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3035 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3035Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3035(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3035(),
	}

	t.Run(
		"not ok: zypper dist-upgrade",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper dist-upgrade`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3035")
		},
	)

	t.Run(
		"not ok: zypper dist-upgrade (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper dist-upgrade`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3035", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"not ok: zypper dist-upgrade (ONBUILD) [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper dup`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3035", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"not ok: zypper dist-upgrade [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper dup`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3035")
		},
	)
}
