package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3061 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3061Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3061(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3061(),
	}

	t.Run(
		"don't warn: ARG then FROM then LABEL",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `ARG A=B
FROM foo
LABEL foo=bar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3061")
		},
	)

	t.Run(
		"don't warn: ARG then FROM then LABEL (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `ARG A=B
FROM foo
LABEL foo=bar`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3061")
		},
	)

	t.Run(
		"don't warn: FROM then ARG then RUN",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM foo
ARG A=B
RUN echo bla`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3061")
		},
	)

	t.Run(
		"don't warn: from before label",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM foo
LABEL foo=bar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3061")
		},
	)

	t.Run(
		"don't warn: from before label (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM foo
LABEL foo=bar`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3061")
		},
	)

	t.Run(
		"don't warn: pragma before FROM",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `# syntax = docker/dockerfile:1.0-experimental
FROM node:16-alpine3.13`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3061")
		},
	)

	t.Run(
		"don't warn: pragma before FROM (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `# syntax = docker/dockerfile:1.0-experimental
FROM node:16-alpine3.13`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3061")
		},
	)

	t.Run(
		"warn: label before from",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL foo=bar
FROM foo`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3061")
		},
	)

	t.Run(
		"warn: label before from (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `LABEL foo=bar
FROM foo`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3061", 1)
			testutils.AssertOnLines(t, violations)
		},
	)
}
