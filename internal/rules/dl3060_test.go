package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3060 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3060Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3060(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3060(),
	}

	t.Run(
		"not ok when cache mount is in wrong location",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=cache,target=/var/lib/foobar yarn install foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"not ok when cache mount is in wrong location (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=cache,target=/var/lib/foobar yarn install foobar`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3060", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"not ok when inheriting from stage with yarn install w/o yarn clean",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node:lts-alpine as foo
RUN yarn install
FROM foo
RUN hey!`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"not ok when inheriting from stage with yarn install w/o yarn clean (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node:lts-alpine as foo
RUN yarn install
FROM foo
RUN hey!`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3060", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"not ok when tmpfs mount is in wrong location",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=tmpfs,target=/var/lib/foobar yarn install foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"not ok when tmpfs mount is in wrong location (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=tmpfs,target=/var/lib/foobar yarn install foobar`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3060", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"not ok when yarn install is in last stage w/o yarn clean",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node:lts-alpine as foo
RUN hey!
FROM scratch
RUN yarn install`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"not ok when yarn install is in last stage w/o yarn clean (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node:lts-alpine as foo
RUN hey!
FROM scratch
RUN yarn install`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3060", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"not ok with clean before install",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yarn cache clean && yarn install foo`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"not ok with no cache clean",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yarn install foo`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"not ok with no cache clean (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yarn install foo`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3060", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"not ok with strict all-args matching for yarn cache clean",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yarn install foo && yarn cache`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"not ok with strict all-args matching for yarn cache clean (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yarn install foo && yarn cache`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3060", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"not ok with strict all-args matching for yarn cache clean (ONBUILD) [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yarn install foo && yarn cache dir`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3060", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"not ok with strict all-args matching for yarn cache clean [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yarn install foo && yarn cache dir`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"ok when cache mount is used",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=cache,target=/root/.cache/yarn yarn install foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"ok when cache mount is used (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=cache,target=/root/.cache/yarn yarn install foobar`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"ok when inheriting from stage with yarn cache clear",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node:lts-alpine as foo
RUN yarn install && yarn cache clean
FROM foo
RUN hey!`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"ok when inheriting from stage with yarn cache clear (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node:lts-alpine as foo
RUN yarn install && yarn cache clean
FROM foo
RUN hey!`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"ok when omitting yarn cache clean in stage that is not reused later (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM node:lts-alpine as foo
RUN yarn install foo
FROM scratch
RUN hey!`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3060", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"ok when tmpfs mount is used",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=tmpfs,target=/root/.cache/yarn yarn install foobar`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"ok when tmpfs mount is used (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=tmpfs,target=/root/.cache/yarn yarn install foobar`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"ok with cache clean",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yarn install bar && yarn cache clean`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"ok with cache clean (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yarn install bar && yarn cache clean`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"ok with non-yarn commands",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN foo`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"ok with non-yarn commands (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN foo`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"ok with strict all-args matching for yarn cache clean",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yarn install foo && yarn cache clean`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3060")
		},
	)

	t.Run(
		"ok with strict all-args matching for yarn cache clean (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yarn install foo && yarn cache clean`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3060")
		},
	)
}
