package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3014 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3014Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3014(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3014(),
	}

	t.Run(
		"apt-get --quiet",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install --quiet python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get --quiet (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install -q python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3014", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"apt-get --quiet --quiet",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install --quiet --quiet python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get --quiet --quiet (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install --quiet --quiet python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get -q",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install -q python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get -q (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install -q python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3014", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"apt-get -q -q",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install -q -q python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get -q -q (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install -q -q python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get -q=2",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install -q=2 python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get -q=2 (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install -q=2 python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get -qq",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install -qq python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get -qq (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install -qq python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get auto yes",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get auto yes (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3014", 1)
			testutils.AssertOnLines(t, violations)
		},
	)

	t.Run(
		"apt-get with assume-yes",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get --assume-yes install python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get with assume-yes (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get --assume-yes install python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get with auto expanded yes",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get --yes install python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get with auto expanded yes (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get --yes install python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get with auto yes",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get -y install python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get with auto yes (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get -y install python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get yes different pos",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install -y python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get yes different pos (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install -y python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get yes shortflag",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install -yq python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)

	t.Run(
		"apt-get yes shortflag (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install -yq python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3014")
		},
	)
}
