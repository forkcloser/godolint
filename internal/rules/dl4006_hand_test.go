package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL4006Hand holds what hadolint's cases leave open: which SHELL has no
// pipefail to ask for.
func TestDL4006Hand(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		shell    string
		reported bool
	}{
		{"cmd, as Docker's Windows default has it", `SHELL ["cmd", "/S", "/C"]`, false},
		{"cmd.exe", `SHELL ["cmd.exe", "/S", "/C"]`, false},
		{"cmd under its Windows path", `SHELL ["C:\\Windows\\System32\\cmd.exe", "/S", "/C"]`, false},
		{"cmd in capitals", `SHELL ["CMD", "/S", "/C"]`, false},
		{"powershell", `SHELL ["powershell", "-Command"]`, false},
		{"pwsh under a path", `SHELL ["/usr/bin/pwsh", "-Command"]`, false},
		{"fish", `SHELL ["/usr/bin/fish", "-c"]`, false},
		{"bash with pipefail", `SHELL ["/bin/bash", "-o", "pipefail", "-c"]`, false},
		{"bash without pipefail", `SHELL ["/bin/bash", "-c"]`, true},
		{"sh", `SHELL ["/bin/sh", "-c"]`, true},
		{"a shell whose name only holds cmd", `SHELL ["/bin/cmdsh", "-c"]`, true},
		{"bash, with a shell that has no pipefail among its arguments", `SHELL ["/bin/bash", "-c", "fish"]`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			failures := testutils.LintDockerfile(
				"FROM a\n"+tc.shell+"\nRUN wget -O - https://example.com/a | tar -xz\n",
				[]rule.Rule{rules.DL4006()},
			)
			if reported := len(failures) > 0; reported != tc.reported {
				t.Errorf("reported %t, want %t", reported, tc.reported)
			}
		})
	}
}
