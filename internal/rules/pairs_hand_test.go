package rules_test

import (
	"slices"
	"testing"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestQuotedPairsHand: the rules that read ENV and LABEL see the keys and
// values with their quotes gone, and every pair of the line. Each row is
// hadolint's verdict, read from its parser (language-docker's
// Parser/Pairs.hs) and the rule.
func TestQuotedPairsHand(t *testing.T) {
	t.Parallel()

	schema := func(label string, kind config.LabelType, strict bool) *config.Config {
		return &config.Config{LabelSchema: map[string]config.LabelType{label: kind}, StrictLabels: strict}
	}

	for _, tc := range []struct {
		name       string
		rule       rule.Rule
		dockerfile string
		lines      []int
	}{
		{
			"pip's switch after a PATH quoted in part",
			rules.DL3042(),
			"FROM a\nENV PATH=\"/opt/venv/bin\":$PATH PIP_NO_CACHE_DIR=1\nRUN pip install foo==1.0\n",
			nil,
		},
		{"pip's switch in single quotes", rules.DL3042(), "FROM a\nENV PIP_NO_CACHE_DIR='1'\nRUN pip install foo==1.0\n", nil},
		{"pip's switch off, in single quotes", rules.DL3042(), "FROM a\nENV PIP_NO_CACHE_DIR='0'\nRUN pip install foo==1.0\n", []int{3}},
		{"a secret after a PATH quoted in part", rules.DL3064(), "FROM a\nENV PATH=\"/opt/bin\":$PATH API_TOKEN=x\n", []int{2}},
		{"a quoted key that names a file", rules.DL3064(), "FROM a\nENV \"DB_PASSWORD_FILE\"=/run/secrets/db\n", nil},
		{"a quoted key used by the next pair", rules.DL3044(), "FROM a\nENV \"A\"=1 B=$A\n", []int{2}},
		{
			"a required label under a quoted key",
			rules.DL3049WithConfig(schema("maintainer", config.LabelTypeRawText, false)),
			"FROM a\nLABEL \"maintainer\"=\"x\"\n",
			nil,
		},
		{
			"a quoted key under strict labels",
			rules.DL3050WithConfig(schema("maintainer", config.LabelTypeRawText, true)),
			"FROM a\nLABEL \"maintainer\"=\"x\"\n",
			nil,
		},
		{"a quoted key is a valid key", rules.DL3048(), "FROM a\nLABEL \"com.example.vendor\"=\"ACME\"\n", nil},
		{
			"a typed label under a quoted key",
			rules.DL3058WithConfig(schema("maintainer", config.LabelTypeEmail, false)),
			"FROM a\nLABEL \"maintainer\"=\"not-an-email\"\n",
			[]int{2},
		},
		{
			"a version in single quotes",
			rules.DL3056WithConfig(schema("version", config.LabelTypeSemVer, false)),
			"FROM a\nLABEL version='1.0.0'\n",
			nil,
		},
		{
			"an empty value in single quotes",
			rules.DL3051WithConfig(schema("a", config.LabelTypeRawText, false)),
			"FROM a\nLABEL a=''\n",
			[]int{2},
		},
		{
			"a required label after a value quoted in part",
			rules.DL3049WithConfig(schema("d", config.LabelTypeRawText, false)),
			"FROM a\nLABEL a=\"b\"c d=e\n",
			nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var lines []int
			for _, v := range testutils.LintDockerfile(tc.dockerfile, []rule.Rule{tc.rule}) {
				lines = append(lines, v.Line)
			}

			if !slices.Equal(lines, tc.lines) {
				t.Errorf("reported at lines %v, want %v", lines, tc.lines)
			}
		})
	}
}
