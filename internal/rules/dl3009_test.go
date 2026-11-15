package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3009 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3009Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3009(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3009(),
	}

	t.Run(
		"apt cleanup",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch
RUN apt update && apt install python && rm -rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"apt no cleanup",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch
RUN apt update && apt install python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"apt-get cleanup",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch
RUN apt-get update && apt-get install python && rm -rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"apt-get cleanup in stage image",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM ubuntu as foo
RUN apt-get update && apt-get install python
FROM scratch
RUN echo hey!`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"apt-get no cleanup",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch
RUN apt-get update && apt-get install python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"apt-get no cleanup in intermediate stage",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM ubuntu as foo
RUN apt-get update && apt-get install python
FROM foo
RUN hey!`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"apt-get no cleanup in last stage",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM ubuntu as foo
RUN hey!
FROM scratch
RUN apt-get update && apt-get install python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"aptitude cleanup",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch
RUN aptitude update && aptitude install python && rm -rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"aptitude no cleanup",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch
RUN aptitude update && aptitude install python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"don't warn: cache mount to apt cache and lists directory",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN \
  --mount=type=cache,target=/var/cache/apt \
  --mount=type=cache,target=/var/lib/apt \
  apt-get update`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"don't warn: cache mount to apt cache and tmpfs mount to lists directory",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN \
  --mount=type=cache,target=/var/cache/apt \
  --mount=type=tmpfs,target=/var/lib/apt \
  apt-get update`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"don't warn: tmpfs mount to apt cache and cache mount to lists directory",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN \
  --mount=type=tmpfs,target=/var/cache/apt \
  --mount=type=cache,target=/var/lib/apt \
  apt-get update`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"don't warn: tmpfs mount to apt cache and lists directory",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN \
  --mount=type=tmpfs,target=/var/cache/apt \
  --mount=type=tmpfs,target=/var/lib/apt \
  apt-get update`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"no warn apt-get cleanup in intermediate stage that cleans lists",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM ubuntu as foo
RUN apt-get update && apt-get install python && rm -rf /var/lib/apt/lists/*
FROM foo
RUN hey!`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"no warn apt-get cleanup in intermediate stage when stage not used later",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM ubuntu as foo
RUN apt-get update && apt-get install python
FROM scratch
RUN hey!`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"not ok with a non-recursive removal of the apt lists",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install python && rm /var/lib/apt/lists/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3009")
		},
	)
}
