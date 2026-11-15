package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3049 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3049Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3049(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		LabelSchema: map[string]config.LabelType{
			"foo": config.LabelTypeRawText,
		},
	}
	allRules := []rule.Rule{
		rules.DL3049WithConfig(cfg),
	}

	t.Run(
		"don't warn: two stages, inheritance, label defined in second stage",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM baseimage as base
FROM base
LABEL foo="bar"`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3049")
		},
	)

	t.Run(
		"don't warn: two stages, no inheritance, label in last stage, with copy",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM baseimage as base
FROM anotherimage
LABEL foo="bar"
COPY --from=base /bar /baz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3049")
		},
	)

	t.Run(
		"ok: single stage, label present",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM baseimage
LABEL foo="bar"`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3049")
		},
	)

	t.Run(
		"ok: single stage, label present (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM baseimage
LABEL foo="bar"`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3049")
		},
	)

	t.Run(
		"warn once: multiple results, inheritance, label in one result",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM baseimage as base
FROM base
LABEL foo="bar"

FROM base`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3049", 1)
		},
	)

	t.Run(
		"warn once: multiple results, no inheritance, label in one result",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM baseimage as base
FROM newimage
LABEL foo="bar"
COPY --from=base /bar /baz

FROM newimage
COPY --from=base /bar /baz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3049", 1)
		},
	)

	t.Run(
		"warn once: two stages, label present in first only",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM baseimage
LABEL foo="bar"
FROM newimage`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3049", 1)
		},
	)
}
