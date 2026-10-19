package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3032 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3032Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3032(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3032(),
	}

	t.Run(
		"not ok with clean before install",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yum install -y mariadb-10.4 && yum clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3032")
		},
	)

	t.Run(
		"not ok with clean before install (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yum clean all && yum install -y`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3032")
		},
	)

	t.Run(
		"not ok with no clean all",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yum install -y mariadb-10.4`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3032")
		},
	)

	t.Run(
		"ok with rm -rf /var/cache/yum/*",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yum install -y mariadb-10.4 && rm -rf /var/cache/yum/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3032")
		},
	)

	t.Run(
		"ok with yum clean all ",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yum install -y mariadb-10.4 && yum clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3032")
		},
	)

	t.Run(
		"ok with yum clean all  (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN bash -c ` + "`" + `# not even a yum command` + "`" + ``
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3032")
		},
	)

	t.Run(
		"strict all-args matching for yum cleanup",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yum install -y bash && yum clean`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3032")
		},
	)

	t.Run(
		"strict all-args matching for yum cleanup (2)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yum install -y bash && yum clean cache`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3032")
		},
	)

	t.Run(
		"strict all-args matching for yum cleanup (3)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yum install -y bash && yum clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3032")
		},
	)

	t.Run(
		"strict all-args matching for yum cleanup (4)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yum install -y bash && rm -rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3032")
		},
	)

	t.Run(
		"strict all-args matching for yum cleanup (5)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yum install -y bash && rm -rf /var/cache/yum/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3032")
		},
	)

	t.Run(
		"strict all-args matching for yum cleanup (6)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yum install -y bash && rm -rf /var/cache/libdnf5`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3032")
		},
	)

	t.Run(
		"strict all-args matching for yum cleanup (7)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN rm -rf /foo/bar && yum install -y bash && rm -rf /var/cache/yum/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3032")
		},
	)
}
