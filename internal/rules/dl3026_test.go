package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for DL3026 ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/DL3026Spec.hs
//
// To regenerate: go generate ./internal/rules

func TestDL3026(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{
		rules.DL3026(),
	}

	t.Run(
		"does not warn on empty allowed registries",
		func(t *testing.T) {
			t.Parallel()

			dockerfile := `FROM random.com/debian`
			violations := testutils.LintDockerfile(dockerfile, allRules)

			testutils.AssertNoViolation(t, violations, "DL3026")
		},
	)
}
