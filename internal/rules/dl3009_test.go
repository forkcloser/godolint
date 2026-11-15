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
		"apt cleanup (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch
RUN apt update && apt install python && rm -rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

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
		"apt no cleanup (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch
RUN apt update && apt install python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3009", 1)
			testutils.AssertOnLines(t, violations)
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
		"apt-get cleanup (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch
RUN apt-get update && apt-get install python && rm -rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

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
		"apt-get cleanup in stage image (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM ubuntu as foo
RUN apt-get update && apt-get install python
FROM scratch
RUN echo hey!`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3009", 1)
			testutils.AssertOnLines(t, violations)
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
		"apt-get no cleanup (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch
RUN apt-get update && apt-get install python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3009", 1)
			testutils.AssertOnLines(t, violations)
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
		"apt-get no cleanup in intermediate stage (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM ubuntu as foo
RUN apt-get update && apt-get install python
FROM foo
RUN hey!`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3009", 1)
			testutils.AssertOnLines(t, violations)
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
		"apt-get no cleanup in last stage (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM ubuntu as foo
RUN hey!
FROM scratch
RUN apt-get update && apt-get install python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3009", 1)
			testutils.AssertOnLines(t, violations)
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
		"aptitude cleanup (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch
RUN aptitude update && aptitude install python && rm -rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

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
		"aptitude no cleanup (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM scratch
RUN aptitude update && aptitude install python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3009", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"don't warn: BuildKit cache mount to apt directories 1",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=cache,target=/var/lib/apt \    --mount=type=cache,target=/var/cache/apt \    rm -f /etc/apt/apt.conf.d/docker-clean && \    apt-get update && apt-get install python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"don't warn: BuildKit cache mount to apt directories 1 (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=cache,target=/var/lib/apt \    --mount=type=cache,target=/var/cache/apt \    rm -f /etc/apt/apt.conf.d/docker-clean && \    apt-get update && apt-get install python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"don't warn: BuildKit cache mount to apt directories 2",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN rm -f /etc/apt/apt.conf.d/docker-clean
RUN --mount=type=cache,target=/var/cache/apt \
    --mount=type=cache,target=/var/lib/apt \
    apt-get update && apt-get install foo`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"don't warn: BuildKit cache mount to apt directories 2 (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN rm -f /etc/apt/apt.conf.d/docker-clean
RUN --mount=type=cache,target=/var/cache/apt \
    --mount=type=cache,target=/var/lib/apt \
    apt-get update && apt-get install foo`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"don't warn: BuildKit cache mount to apt lists directory",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=cache,target=/var/lib/apt/lists \    apt-get update && apt-get install python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"don't warn: BuildKit cache mount to apt lists directory (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=cache,target=/var/lib/apt/lists \    apt-get update && apt-get install python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
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
		"don't warn: cache mount to apt cache and lists directory (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN \
  --mount=type=cache,target=/var/cache/apt \
  --mount=type=cache,target=/var/lib/apt \
  apt-get update`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

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
		"don't warn: cache mount to apt cache and tmpfs mount to lists directory (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN \
  --mount=type=cache,target=/var/cache/apt \
  --mount=type=tmpfs,target=/var/lib/apt \
  apt-get update`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

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
		"don't warn: tmpfs mount to apt cache and cache mount to lists directory (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN \
  --mount=type=tmpfs,target=/var/cache/apt \
  --mount=type=cache,target=/var/lib/apt \
  apt-get update`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

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
		"don't warn: tmpfs mount to apt cache and lists directory (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN \
  --mount=type=tmpfs,target=/var/cache/apt \
  --mount=type=tmpfs,target=/var/lib/apt \
  apt-get update`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

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
		"no warn apt-get cleanup in intermediate stage that cleans lists (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM ubuntu as foo
RUN apt-get update && apt-get install python && rm -rf /var/lib/apt/lists/*
FROM foo
RUN hey!`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

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
		"no warn apt-get cleanup in intermediate stage when stage not used later (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM ubuntu as foo
RUN apt-get update && apt-get install python
FROM scratch
RUN hey!`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3009", 1)
			testutils.AssertOnLines(t, violations)
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

	t.Run(
		"not ok with a non-recursive removal of the apt lists (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install python && rm /var/lib/apt/lists/*`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3009", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"not ok with strict all-args matching for aptitude cleanup",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN aptitude update && aptitude install python && rm -rf /var/cache/libdnf5`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"not ok with strict all-args matching for aptitude cleanup (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN aptitude update && aptitude install python && rm -rf /var/cache/libdnf5`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3009", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"not ok with strict all-args matching for aptitude cleanup (ONBUILD) [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN aptitude update && aptitude install python && rm -rf /var/cache/yum/*`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3009", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"not ok with strict all-args matching for aptitude cleanup [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN aptitude update && aptitude install python && rm -rf /var/cache/yum/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with any form of recursive removal of the apt lists",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install python && rm -fr /var/lib/apt/lists/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with any form of recursive removal of the apt lists (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install python && rm -fr /var/lib/apt/lists/*`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with any form of recursive removal of the apt lists (ONBUILD) [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install python && rm -r -f /var/lib/apt/lists/*`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with any form of recursive removal of the apt lists (ONBUILD) [3]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install python && rm -Rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with any form of recursive removal of the apt lists (ONBUILD) [4]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install python && rm --recursive --force /var/lib/apt/lists/*`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with any form of recursive removal of the apt lists (ONBUILD) [5]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install python && rm -rf /var/lib/apt/lists`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with any form of recursive removal of the apt lists (ONBUILD) [6]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install python && rm -rf /var/lib/apt/lists/`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with any form of recursive removal of the apt lists [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install python && rm -r -f /var/lib/apt/lists/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with any form of recursive removal of the apt lists [3]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install python && rm -Rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with any form of recursive removal of the apt lists [4]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install python && rm --recursive --force /var/lib/apt/lists/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with any form of recursive removal of the apt lists [5]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install python && rm -rf /var/lib/apt/lists`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with any form of recursive removal of the apt lists [6]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install python && rm -rf /var/lib/apt/lists/`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with strict all-args matching for aptitude cleanup",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN aptitude update && aptitude install python && rm -rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with strict all-args matching for aptitude cleanup (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN aptitude update && aptitude install python && rm -rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with strict all-args matching for aptitude cleanup (ONBUILD) [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN rm -rf /foo/bar && aptitude update && aptitude install python && rm -rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with strict all-args matching for aptitude cleanup (ONBUILD) [3]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN aptitude update && rm -rf /foo/bar && aptitude install python && rm -rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with strict all-args matching for aptitude cleanup [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN rm -rf /foo/bar && aptitude update && aptitude install python && rm -rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"ok with strict all-args matching for aptitude cleanup [3]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN aptitude update && rm -rf /foo/bar && aptitude install python && rm -rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"warn: BuildKit cache mount to apt cache directory only",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=cache,target=/var/cache/apt \
    rm -f /etc/apt/apt.conf.d/docker-clean && \
    apt-get update && apt-get install python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"warn: BuildKit cache mount to apt cache directory only (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=cache,target=/var/cache/apt \
    rm -f /etc/apt/apt.conf.d/docker-clean && \
    apt-get update && apt-get install python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3009", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"warn: BuildKit cache mount to apt lists directory only",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN rm -f /etc/apt/apt.conf.d/docker-clean
RUN --mount=type=cache,target=/var/lib/apt \
    apt-get update && apt-get install python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3009")
		},
	)

	t.Run(
		"warn: BuildKit cache mount to apt lists directory only (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN rm -f /etc/apt/apt.conf.d/docker-clean
RUN --mount=type=cache,target=/var/lib/apt \
    apt-get update && apt-get install python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3009", 1)
			testutils.AssertOnLines(t, violations)
		},
	)
}
