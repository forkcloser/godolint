package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3034 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3034Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3034(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3034(),
	}

	t.Run(
		"not ok without non-interactive switch",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install httpd=2.4.24 && zypper clean`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3034")
		},
	)

	t.Run(
		"not ok without non-interactive switch (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install httpd=2.4.24 && zypper clean`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3034", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"ok with non-interactive switch present",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -n httpd=2.4.24 && zypper clean`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3034")
		},
	)

	t.Run(
		"ok with non-interactive switch present (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -n httpd=2.4.24 && zypper clean`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3034")
		},
	)

	t.Run(
		"ok with non-interactive switch present (ONBUILD) [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install --non-interactive httpd=2.4.24 && zypper clean`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3034")
		},
	)

	t.Run(
		"ok with non-interactive switch present (ONBUILD) [3]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -y httpd=2.4.24 && zypper clean`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3034")
		},
	)

	t.Run(
		"ok with non-interactive switch present (ONBUILD) [4]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install --no-confirm httpd=2.4.24 && zypper clean`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3034")
		},
	)

	t.Run(
		"ok with non-interactive switch present [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install --non-interactive httpd=2.4.24 && zypper clean`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3034")
		},
	)

	t.Run(
		"ok with non-interactive switch present [3]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install -y httpd=2.4.24 && zypper clean`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3034")
		},
	)

	t.Run(
		"ok with non-interactive switch present [4]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN zypper install --no-confirm httpd=2.4.24 && zypper clean`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3034")
		},
	)
}
