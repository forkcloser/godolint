package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3052 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3052Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3052(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		LabelSchema: map[string]config.LabelType{
			"urllabel": config.LabelTypeURL,
		},
	}
	allRules := []rule.Rule{
		rules.DL3052WithConfig(cfg),
	}

	t.Run(
		"not ok with label not containing URL",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL urllabel="not-url"`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3052")
		},
	)

	t.Run(
		"not ok with label not containing URL (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL urllabel="not-url"`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3052", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"ok with label containing URL",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL urllabel="http://example.com"`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3052")
		},
	)

	t.Run(
		"ok with label containing URL (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL urllabel="http://example.com"`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3052")
		},
	)

	t.Run(
		"ok with other label not containing URL",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL other="foo"`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3052")
		},
	)

	t.Run(
		"ok with other label not containing URL (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL other="bar"`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3052")
		},
	)
}
