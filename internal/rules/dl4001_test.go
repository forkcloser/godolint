package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL4001 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL4001Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL4001(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL4001(),
	}

	t.Run(
		"does not warn when using both curl and wget in different stages",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget my.xyz
FROM scratch
RUN curl localhost`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL4001")
		},
	)

	t.Run(
		"does not warn when using only wget",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget my.xyz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL4001")
		},
	)

	t.Run(
		"does not warns when using both, on a single stage",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget my.xyz
RUN curl localhost
FROM scratch
RUN curl localhost`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL4001")
		},
	)

	t.Run(
		"only warns on the relevant RUN instruction",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget my.xyz
RUN curl my.xyz
RUN echo hello`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL4001", 1)
		},
	)

	t.Run(
		"warns when using both wget and curl",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget my.xyz
RUN curl localhost`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL4001")
		},
	)

	t.Run(
		"warns when using both wget and curl in same instruction",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget my.xyz && curl localhost`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL4001")
		},
	)
}
