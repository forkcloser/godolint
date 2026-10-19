package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3066 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3066Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3066(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3066(),
	}

	t.Run(
		"not ok: UID is a variable that isn't defined",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `ENV APP_UID=1000
USER $OTHER_UID`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3066")
		},
	)

	t.Run(
		"not ok: UID is numeric but USER with non-numeric UID follows",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM foo:bar
USER 1234
USER foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3066")
		},
	)

	t.Run(
		"not ok: UID is numeric not in last build stage",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM foo:bar AS build
USER 1234
FROM foobar:barfoo
USER foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3066")
		},
	)

	t.Run(
		"not ok: non-numeric UID",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM foo:bar
USER foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3066")
		},
	)

	t.Run(
		"not ok: non-numeric UID and GID",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM foo:bar
USER foobar:barfoo`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3066")
		},
	)

	t.Run(
		"ok: UID and GID are non-numeric, but are defined ENVs",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `ENV APP_UID=1000
USER $APP_UID`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3066")
		},
	)

	t.Run(
		"ok: UID and GID are non-numeric, but are defined ENVs (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `ENV USER_UID=1000 USER_GID=1000
USER ${USER_UID}:${USER_GID}`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3066")
		},
	)

	t.Run(
		"ok: UID is non-numeric but USER with numeric UID follows",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM foo:bar
USER foobar
USER 1234`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3066")
		},
	)

	t.Run(
		"ok: UID is non-numeric in any but last build stage",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM foo:bar AS build
USER foobar
FROM foobar:barfoo
USER 1234`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3066")
		},
	)

	t.Run(
		"ok: UID is non-numeric, but is a defined ARG",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `ARG APP_UID
FROM foobar:barfoo
USER ${APP_UID}`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3066")
		},
	)

	t.Run(
		"ok: numeric UID",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM foo:bar
USER 1234`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3066")
		},
	)

	t.Run(
		"ok: numeric UID and GID",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM foo:bar
USER 1234:5678`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3066")
		},
	)
}
