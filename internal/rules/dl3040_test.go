package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3040 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3040Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3040(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3040(),
	}

	t.Run(
		"different install command variants",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y install`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"different install command variants (10)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y in && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"different install command variants (11)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y upgrade && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"different install command variants (12)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y up && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"different install command variants (13)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y upgrade-minimal && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"different install command variants (14)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y up-min && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"different install command variants (15)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y reinstall && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"different install command variants (16)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y rei && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"different install command variants (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y in`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"different install command variants (3)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y upgrade`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"different install command variants (4)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y up`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"different install command variants (5)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y upgrade-minimal`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"different install command variants (6)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y up-min`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"different install command variants (7)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y reinstall`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"different install command variants (8)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y rei`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"different install command variants (9)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y install && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"no ok without dnf clean all",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y mariadb-10.4`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"no ok without dnf clean all (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN microdnf install -y mariadb-10.4`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"no ok without dnf clean all (3)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf in -y mariadb-10.4`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"not ok with `dnf upgrade`",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y upgrade`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"not ok with `dnf upgrade` (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf -y up`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"not ok with clean before install",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN microdnf clean all && dnf install -y mariadb-10.4`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"not ok with clean before install (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN microdnf clean all && dnf in -y mariadb-10.4`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"not ok with clean before install (3)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN rm -rf /var/cache/libdnf5 && dnf install -y mariadb-10.4`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"not ok with clean before install (4)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN rm -rf /var/cache/libdnf5 && dnf in -y mariadb-10.4`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"not ok with clean before install (5)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN rm -rf /var/cache/libdnf5 && microdnf install -y mariadb-10.4`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"not ok with rm /var/cache/yum",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y mariadb-10.4 && rm -rf /var/cache/yum/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"not ok with rm /var/cache/yum (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf in -y mariadb-10.4 && rm -rf /var/cache/yum/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"not ok with rm /var/cache/yum (3)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN microdnf install -y mariadb-10.4 && rm -rf /var/cache/yum/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"ok with cache mount at /var/cache/yum",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=cache,target=/var/cache/libdnf5 dnf install -y mariadb-10.4`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"ok with cache mount at /var/cache/yum (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=cache,target=/var/cache/libdnf5 dnf in -y mariadb-10.4`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"ok with cache mount at /var/cache/yum (3)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=cache,target=/var/cache/libdnf5 microdnf install -y mariadb-10.4`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"ok with dnf clean all",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y mariadb-10.4 && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"ok with dnf clean all (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf in -y mariadb-10.4 && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"ok with dnf clean all (3)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN microdnf install -y mariadb-10.4 && microdnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"ok with dnf clean all (4)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN notdnf install mariadb`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"ok with tmpfs mount at /var/cache/yum",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=tmpfs,target=/var/cache/libdnf5 dnf install -y mariadb-10.4`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"ok with tmpfs mount at /var/cache/yum (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=tmpfs,target=/var/cache/libdnf5 dnf in -y mariadb-10.4`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"ok with tmpfs mount at /var/cache/yum (3)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN --mount=type=tmpfs,target=/var/cache/libdnf5 microdnf install -y mariadb-10.4`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"strict all-args matching for dnf cleanup",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y bash && dnf clean`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"strict all-args matching for dnf cleanup (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y bash && dnf clean cache`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"strict all-args matching for dnf cleanup (3)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y bash && dnf clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"strict all-args matching for dnf cleanup (4)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y bash && rm -rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"strict all-args matching for dnf cleanup (5)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y bash && rm -rf /var/cache/libdnf5`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"strict all-args matching for dnf cleanup (6)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN dnf install -y bash && rm -rf /var/cache/yum/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3040")
		},
	)

	t.Run(
		"strict all-args matching for dnf cleanup (7)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN rm -rf /foo/bar && dnf install -y bash && rm -rf /var/cache/libdnf5`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3040")
		},
	)
}
