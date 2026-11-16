package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3046Hand holds what hadolint's cases leave open: the ways a useradd
// is given its UID.
func TestDL3046Hand(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		command  string
		reported bool
	}{
		{"after -u", "useradd -u 123456 foo", true},
		{"after -u last in a cluster", "useradd -mu 123456 foo", true},
		{"after --uid", "useradd --uid 123456 foo", true},
		{"glued to --uid", "useradd --uid=123456 foo", true},
		{"after -u, with flags after it", "useradd -u 123456 -m -s /bin/sh foo", true},
		{"a short UID in a cluster", "useradd -mu 12345 foo", false},
		{"a short UID after --uid", "useradd --uid 12345 foo", false},
		{"-l in the cluster", "useradd -lu 123456 foo", false},
		{"--no-log-init beside --uid", "useradd --no-log-init --uid=123456 foo", false},
		{"-u not last in its cluster", "useradd -um 123456 foo", false},
		{"a long argument that is no UID", "useradd -m -s /bin/bash -u 1000 administrator", false},
		{"no UID given", "useradd -m administrator", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			failures := testutils.LintDockerfile("FROM a\nRUN "+tc.command+"\n", []rule.Rule{rules.DL3046()})
			if reported := len(failures) > 0; reported != tc.reported {
				t.Errorf("reported %t, want %t", reported, tc.reported)
			}
		})
	}
}
