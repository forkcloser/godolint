package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3011Hand holds the upper bound hadolint's own cases leave open: 65535
// is a port, 65536 is not, at either end of a range.
func TestDL3011Hand(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{rules.DL3011()}

	for _, tc := range []struct {
		name      string
		expose    string
		violation bool
	}{
		{"the last port", `EXPOSE 65535`, false},
		{"one past it", `EXPOSE 65536`, true},
		{"a range ending at the last port", `EXPOSE 1024-65535/tcp`, false},
		{"a range ending one past it", `EXPOSE 1024-65536`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			violations := testutils.LintDockerfile(tc.expose, allRules)
			if tc.violation {
				testutils.AssertContainsViolation(t, violations, "DL3011")
			} else {
				testutils.AssertNoViolation(t, violations, "DL3011")
			}
		})
	}
}
