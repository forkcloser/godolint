package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3037 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3037Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3037(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3037(),
	}

	t.Run(
		"not ok without zypper version pinning",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -y tomcat && zypper clean`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3037")
		},
	)

	t.Run(
		"not ok without zypper version pinning (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -y tomcat && zypper clean`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3037", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"ok with different variants of zypper version pinning",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -y tomcat=9.0.39 && zypper clean`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3037")
		},
	)

	t.Run(
		"ok with different variants of zypper version pinning (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -y tomcat=9.0.39 && zypper clean`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3037")
		},
	)

	t.Run(
		"ok with different variants of zypper version pinning (ONBUILD) [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -y tomcat\>=9.0 && zypper clean`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3037")
		},
	)

	t.Run(
		"ok with different variants of zypper version pinning (ONBUILD) [3]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -y tomcat\>9.0 && zypper clean`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3037")
		},
	)

	t.Run(
		"ok with different variants of zypper version pinning (ONBUILD) [4]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -y tomcat\<=9.0 && zypper clean`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3037")
		},
	)

	t.Run(
		"ok with different variants of zypper version pinning (ONBUILD) [5]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -y tomcat\<9.0 && zypper clean`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3037")
		},
	)

	t.Run(
		"ok with different variants of zypper version pinning (ONBUILD) [6]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -y tomcat-9.0.39-1.rpm && zypper clean`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3037")
		},
	)

	t.Run(
		"ok with different variants of zypper version pinning [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -y tomcat\>=9.0 && zypper clean`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3037")
		},
	)

	t.Run(
		"ok with different variants of zypper version pinning [3]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -y tomcat\>9.0 && zypper clean`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3037")
		},
	)

	t.Run(
		"ok with different variants of zypper version pinning [4]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -y tomcat\<=9.0 && zypper clean`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3037")
		},
	)

	t.Run(
		"ok with different variants of zypper version pinning [5]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -y tomcat\<9.0 && zypper clean`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3037")
		},
	)

	t.Run(
		"ok with different variants of zypper version pinning [6]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -y tomcat-9.0.39-1.rpm && zypper clean`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3037")
		},
	)
}
