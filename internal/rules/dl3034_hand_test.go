package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3034Hand holds the zypper subcommand aliases hadolint's own cases
// leave open: in, remove, rm, source-install, si and patch want the
// non-interactive switch as install does.
func TestDL3034Hand(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{rules.DL3034()}

	for _, sub := range []string{"in", "remove", "rm", "source-install", "si", "patch"} {
		t.Run(sub, func(t *testing.T) {
			t.Parallel()

			violations := testutils.LintDockerfile(`RUN zypper `+sub+` httpd && zypper clean`, allRules)
			testutils.AssertContainsViolation(t, violations, "DL3034")

			violations = testutils.LintDockerfile(`RUN zypper `+sub+` -n httpd && zypper clean`, allRules)
			testutils.AssertNoViolation(t, violations, "DL3034")
		})
	}
}
