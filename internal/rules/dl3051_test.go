package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3051 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3051Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3051(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		LabelSchema: map[string]config.LabelType{
			"emptylabel": config.LabelTypeRawText,
		},
	}
	allRules := []rule.Rule{
		rules.DL3051WithConfig(cfg),
	}

	t.Run(
		"not ok with label empty",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL emptylabel=""`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3051")
		},
	)

	t.Run(
		"not ok with label empty (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL emptylabel=""`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3051", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"ok with label not empty",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL emptylabel="foo"`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3051")
		},
	)

	t.Run(
		"ok with label not empty (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL emptylabel="bar"`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3051")
		},
	)

	t.Run(
		"ok with other label empty",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL other=""`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3051")
		},
	)

	t.Run(
		"ok with other label empty (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL other=""`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3051")
		},
	)
}
