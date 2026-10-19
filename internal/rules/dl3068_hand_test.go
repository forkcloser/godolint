package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3068Hand holds two readings of hadolint's rule that its own cases
// leave open: the channel is dropped before the version is read, so a
// channel ending in a digit does not pin an unversioned name; and an http(s)
// URL is fixed as it stands, archive suffix or not.
func TestDL3068Hand(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{rules.DL3068()}

	t.Run("a channel ending in a digit pins nothing", func(t *testing.T) {
		t.Parallel()

		violations := testutils.LintDockerfile(`RUN pecl install foo-1/redis`, allRules)
		testutils.AssertContainsViolation(t, violations, "DL3068")
	})

	t.Run("a URL without an archive suffix is fixed", func(t *testing.T) {
		t.Parallel()

		violations := testutils.LintDockerfile(`RUN pecl install https://example.com/redis`, allRules)
		testutils.AssertNoViolation(t, violations, "DL3068")
	})
}
