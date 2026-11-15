package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3058HandAddress: the email type wants an address and nothing more,
// as hadolint's does: a name with it, or angle brackets around it, is not
// one.
func TestDL3058HandAddress(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{LabelSchema: map[string]config.LabelType{"maintainer": config.LabelTypeEmail}}

	for value, reported := range map[string]bool{
		"john@example.com":                   false,
		"john.doe+tag@sub.example.com":       false,
		"John Doe <john@example.com>":        true,
		"<john@example.com>":                 true,
		"john@example.com, jane@example.com": true,
		"john":                               true,
		"":                                   true,
	} {
		violations := testutils.LintDockerfile(
			"FROM a\nLABEL maintainer=\""+value+"\"\n",
			[]rule.Rule{rules.DL3058WithConfig(cfg)},
		)
		if got := len(violations) != 0; got != reported {
			t.Errorf("maintainer=%q: reported %v, want %v", value, got, reported)
		}
	}
}
