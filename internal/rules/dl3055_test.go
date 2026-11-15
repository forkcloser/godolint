package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3055 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3055Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3055(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		LabelSchema: map[string]config.LabelType{
			"githash": config.LabelTypeGitHash,
		},
	}
	allRules := []rule.Rule{
		rules.DL3055WithConfig(cfg),
	}

	t.Run(
		"not ok with label not containing git hash",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL githash="not-git-hash"`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3055")
		},
	)

	t.Run(
		"not ok with label not containing git hash (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL githash="not-git-hash"`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3055", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"ok with label containing long git hash",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL githash="43c572f1272b6b3171dd1db9e41b7027128ce080"`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3055")
		},
	)

	t.Run(
		"ok with label containing long git hash (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL githash="43c572f1272b6b3171dd1db9e41b7027128ce080"`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3055")
		},
	)

	t.Run(
		"ok with label containing short git hash",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL githash="2dbfae9"`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3055")
		},
	)

	t.Run(
		"ok with label containing short git hash (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL githash="2dbfae9"`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3055")
		},
	)

	t.Run(
		"ok with other label not containing git hash",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL other="foo"`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3055")
		},
	)

	t.Run(
		"ok with other label not containing git hash (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL other="bar"`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3055")
		},
	)
}
