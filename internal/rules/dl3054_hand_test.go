package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3054Hand holds what hadolint's cases leave open: which values are an
// identifier of the SPDX License List.
func TestDL3054Hand(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{LabelSchema: map[string]config.LabelType{"license": config.LabelTypeSPDX}}

	for _, tc := range []struct {
		name     string
		value    string
		reported bool
	}{
		{"an identifier", "MIT", false},
		{"an identifier with a version", "Apache-2.0", false},
		{"an identifier that starts with a digit", "0BSD", false},
		{"an identifier with no version", "Zlib", false},
		{"an identifier that holds another", "MIT-0", false},
		{"the only and or-later forms", "GPL-3.0-only", false},
		{"an identifier the list deprecates", "GPL-3.0", true},
		{"a name of the right shape that is on no list", "Foo-1.0", true},
		{"an identifier in another case", "mit", true},
		{"an expression", "MIT OR Apache-2.0", true},
		{"an identifier and an exception", "GPL-2.0-only WITH Classpath-exception-2.0", true},
		{"white space after the identifier", "MIT \t", false},
		{"white space before the identifier", " MIT", true},
		{"nothing", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			failures := testutils.LintDockerfile(
				"FROM a\nLABEL license=\""+tc.value+"\"\n",
				[]rule.Rule{rules.DL3054WithConfig(cfg)},
			)
			if reported := len(failures) > 0; reported != tc.reported {
				t.Errorf("reported %t, want %t", reported, tc.reported)
			}
		})
	}
}
