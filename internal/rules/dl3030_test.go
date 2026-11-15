package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3030 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3030Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3030(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3030(),
	}

	t.Run(
		"not ok when not using `-y` switch",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yum install httpd-2.4.24 && yum clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3030")
		},
	)

	t.Run(
		"not ok when not using `-y` switch (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yum install httpd-2.4.24 && yum clean all`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3030", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"ok when using `-y` switch",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yum install -y httpd-2.4.24 && yum clean all`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3030")
		},
	)

	t.Run(
		"ok when using `-y` switch (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN yum install -y httpd-2.4.24 && yum clean all`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3030")
		},
	)

	t.Run(
		"ok when using `-y` switch (ONBUILD) [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN bash -c ` + "`" + `# not even a yum command` + "`" + ``
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3030")
		},
	)

	t.Run(
		"ok when using `-y` switch [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN bash -c ` + "`" + `# not even a yum command` + "`" + ``
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3030")
		},
	)
}
