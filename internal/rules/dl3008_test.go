package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3008 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3008Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3008(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3008(),
	}

	t.Run(
		"apt-get pinned",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get -y --no-install-recommends install nodejs=0.10`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3008")
		},
	)

	t.Run(
		"apt-get pinned (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get -y --no-install-recommends install nodejs=0.10`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3008")
		},
	)

	t.Run(
		"apt-get pinned chained",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update \
 && apt-get -yqq --no-install-recommends install nodejs=0.10 \
 && rm -rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3008")
		},
	)

	t.Run(
		"apt-get pinned chained (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update \
 && apt-get -yqq --no-install-recommends install nodejs=0.10 \
 && rm -rf /var/lib/apt/lists/*`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3008")
		},
	)

	t.Run(
		"apt-get pinned regression",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install --no-install-recommends -y \
python-demjson=2.2.2* \
wget=1.16.1* \
git=1:2.5.0* \
ruby=1:2.1.*`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3008")
		},
	)

	t.Run(
		"apt-get pinned regression (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install --no-install-recommends -y \
python-demjson=2.2.2* \
wget=1.16.1* \
git=1:2.5.0* \
ruby=1:2.1.*`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3008")
		},
	)

	t.Run(
		"apt-get version",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install -y python=1.2.2`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3008")
		},
	)

	t.Run(
		"apt-get version (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install -y python=1.2.2`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3008")
		},
	)

	t.Run(
		"apt-get version (ONBUILD) [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install ./wkhtmltox_0.12.5-1.bionic_amd64.deb`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3008")
		},
	)

	t.Run(
		"apt-get version [2]",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get install ./wkhtmltox_0.12.5-1.bionic_amd64.deb`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3008")
		},
	)

	t.Run(
		"apt-get version pinning",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install python`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertContainsViolation(t, violations, "DL3008")
		},
	)

	t.Run(
		"apt-get version pinning (ONBUILD)",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `RUN apt-get update && apt-get install python`
			violations := testutils.LintDockerfileOnBuild(dockerfile, allRules)

			testutils.AssertViolationCount(t, violations, "DL3008", 1)
			testutils.AssertOnLines(t, violations)
		},
	)
}
