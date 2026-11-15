package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3047 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3047Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3047(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3047(),
	}

	t.Run(
		"does not warn when running with --append-output (append-output long option) and without --progress option",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget --append-output=/tmp/wget.log my.xyz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"does not warn when running with --append-output (append-output long option) and without --progress option (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget --append-output=/tmp/wget.log my.xyz`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"does not warn when running with --no-verbose (no-verbose long option) and without --progress option",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget --no-verbose my.xyz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"does not warn when running with --no-verbose (no-verbose long option) and without --progress option (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget --no-verbose my.xyz`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"does not warn when running with --output-file (output-file long option) and without --progress option",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget --output-file=/tmp/wget.log my.xyz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"does not warn when running with --output-file (output-file long option) and without --progress option (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget --output-file=/tmp/wget.log my.xyz`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"does not warn when running with --progress option",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget --progress=dot:giga my.xyz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"does not warn when running with --progress option (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget --progress=dot:giga my.xyz`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"does not warn when running with --quiet (quiet long option) and without --progress option",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget --quiet my.xyz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"does not warn when running with --quiet (quiet long option) and without --progress option (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget --quiet my.xyz`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"does not warn when running with -a (append-output long option) and without --progress option",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget -a /tmp/wget.log my.xyz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"does not warn when running with -a (append-output long option) and without --progress option (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget -a /tmp/wget.log my.xyz`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"does not warn when running with -nv (no-verbose short option) and without --progress option",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget -nv my.xyz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"does not warn when running with -nv (no-verbose short option) and without --progress option (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget -nv my.xyz`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"does not warn when running with -o (output-file long option) and without --progress option",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget -o /tmp/wget.log my.xyz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"does not warn when running with -o (output-file long option) and without --progress option (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget -o /tmp/wget.log my.xyz`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"does not warn when running with -q (quiet short option) and without --progress option",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget -q my.xyz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"does not warn when running with -q (quiet short option) and without --progress option (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget -q my.xyz`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"warns when using wget without --progress option",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget my.xyz`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3047")
		},
	)

	t.Run(
		"warns when using wget without --progress option (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node as foo
RUN wget my.xyz`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3047", 1)
			testutils.AssertOnLines(t, violations)
		},
	)
}
