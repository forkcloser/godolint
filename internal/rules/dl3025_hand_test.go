package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3025Hand holds what hadolint's cases leave open about HEALTHCHECK:
// NONE has no command, in either case of the word, and the options before
// CMD change nothing of its form.
func TestDL3025Hand(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{rules.DL3025()}

	for _, tc := range []struct {
		name       string
		dockerfile string
		violation  bool
	}{
		{"HEALTHCHECK NONE", "HEALTHCHECK NONE", false},
		{"healthcheck none", "healthcheck none", false},
		{"options, then the shell form", "HEALTHCHECK --interval=5s --timeout=3s CMD curl -f http://localhost/", true},
		{"options, then the JSON form", `HEALTHCHECK --interval=5s CMD ["curl", "-f", "http://localhost/"]`, false},
		{"a command named NONE in the shell form", "HEALTHCHECK CMD NONE", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			violations := testutils.LintDockerfile(tc.dockerfile, allRules)
			if tc.violation {
				testutils.AssertContainsViolation(t, violations, "DL3025")
			} else {
				testutils.AssertNoViolation(t, violations, "DL3025")
			}
		})
	}
}
