package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3064Hand holds what hadolint's own cases leave open: the _FILE suffix
// is matched in any case, as the sensitive words are, and privatekey in one
// word is a secret as private_key is.
func TestDL3064Hand(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{rules.DL3064()}

	for _, tc := range []struct {
		name       string
		dockerfile string
		violation  bool
	}{
		{"a lower-case _file suffix", "FROM debian:12\nARG password_file=/run/secrets/db", false},
		{"a mixed-case _File suffix", "FROM debian:12\nENV Api_Token_File=/run/secrets/token", false},
		{"_file inside the name, not at its end", "FROM debian:12\nENV PASSWORD_FILE_CONTENT=hunter2", true},
		{"privatekey in one word", "FROM debian:12\nENV APP_PRIVATEKEY=abc", true},
		{"private alone", "FROM debian:12\nENV NPM_CONFIG_PRIVATE=true", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			violations := testutils.LintDockerfile(tc.dockerfile, allRules)
			if tc.violation {
				testutils.AssertContainsViolation(t, violations, "DL3064")
			} else {
				testutils.AssertNoViolation(t, violations, "DL3064")
			}
		})
	}
}
