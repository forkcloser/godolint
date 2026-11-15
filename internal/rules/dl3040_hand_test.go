package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3040Hand holds the forms of dnf's cache path hadolint's own cases
// leave open: each of the four accepted spellings on its own, and a path that
// only starts like the cache.
func TestDL3040Hand(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{rules.DL3040()}

	for _, tc := range []struct {
		name      string
		cleanup   string
		violation bool
	}{
		{"the directory", `rm -rf /var/cache/libdnf5`, false},
		{"the directory with a trailing slash", `rm -rf /var/cache/libdnf5/`, false},
		{"its contents", `rm -rf /var/cache/libdnf5/*`, false},
		{"the directory by glob", `rm -rf /var/cache/libdnf5*`, false},
		{"a directory inside it", `rm -rf /var/cache/libdnf5/repos`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			violations := testutils.LintDockerfile("RUN dnf -y install bash && "+tc.cleanup, allRules)
			if tc.violation {
				testutils.AssertContainsViolation(t, violations, "DL3040")
			} else {
				testutils.AssertNoViolation(t, violations, "DL3040")
			}
		})
	}
}
