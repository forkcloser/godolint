package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3038Hand holds the dnf subcommands hadolint's own cases leave open:
// groupinstall and localinstall want -y as install does, for dnf and
// microdnf alike.
func TestDL3038Hand(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{rules.DL3038()}

	for _, cmd := range []string{"dnf groupinstall", "dnf localinstall", "microdnf groupinstall"} {
		t.Run(cmd, func(t *testing.T) {
			t.Parallel()

			violations := testutils.LintDockerfile(`RUN `+cmd+` httpd && dnf clean all`, allRules)
			testutils.AssertContainsViolation(t, violations, "DL3038")

			violations = testutils.LintDockerfile(`RUN `+cmd+` -y httpd && dnf clean all`, allRules)
			testutils.AssertNoViolation(t, violations, "DL3038")
		})
	}
}
