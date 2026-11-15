package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3009Hand holds hadolint's "ok with any form of recursive removal of
// the apt lists" cases, which the generator does not port: the spec maps one
// assertion over a list of Dockerfiles.
func TestDL3009Hand(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{rules.DL3009()}

	for _, tc := range []struct {
		name      string
		cleanup   string
		violation bool
	}{
		{"flags in the other order", `rm -fr /var/lib/apt/lists/*`, false},
		{"flags apart", `rm -r -f /var/lib/apt/lists/*`, false},
		{"upper-case recursive", `rm -Rf /var/lib/apt/lists/*`, false},
		{"long flags", `rm --recursive --force /var/lib/apt/lists/*`, false},
		{"the directory itself", `rm -rf /var/lib/apt/lists`, false},
		{"the directory with a trailing slash", `rm -rf /var/lib/apt/lists/`, false},
		{"another directory", `rm -rf /var/lib/apt`, true},
		{"force without recursive", `rm -f /var/lib/apt/lists/*`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dockerfile := "RUN apt-get update && apt-get install python && " + tc.cleanup

			violations := testutils.LintDockerfile(dockerfile, allRules)
			if tc.violation {
				testutils.AssertContainsViolation(t, violations, "DL3009")
			} else {
				testutils.AssertNoViolation(t, violations, "DL3009")
			}
		})
	}
}
