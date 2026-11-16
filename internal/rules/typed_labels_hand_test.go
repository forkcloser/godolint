package rules_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestTypedLabelsHand: each label of the schema that a LABEL gives a bad
// value is reported, once a label and by name, as hadolint's one rule a
// label has it: its cases give one label at a time.
func TestTypedLabelsHand(t *testing.T) {
	t.Parallel()

	for _, kind := range []struct {
		name      string
		rule      func(*config.Config) rule.Rule
		labelType config.LabelType
		good, bad string
	}{
		{"DL3051", rules.DL3051WithConfig, config.LabelTypeRawText, "x", `""`},
		{"DL3052", rules.DL3052WithConfig, config.LabelTypeURL, "https://example.com/a", "nope"},
		{"DL3053", rules.DL3053WithConfig, config.LabelTypeRFC3339, "2021-03-10T10:26:33Z", "yesterday"},
		{"DL3054", rules.DL3054WithConfig, config.LabelTypeSPDX, "MIT", "not a licence"},
		{"DL3055", rules.DL3055WithConfig, config.LabelTypeGitHash, "abcdef1", "xyz"},
		{"DL3056", rules.DL3056WithConfig, config.LabelTypeSemVer, "1.2.3", "one"},
		{"DL3058", rules.DL3058WithConfig, config.LabelTypeEmail, "john@example.com", "john"},
	} {
		cfg := &config.Config{LabelSchema: map[string]config.LabelType{
			"alpha": kind.labelType, "beta": kind.labelType, "gamma": kind.labelType,
		}}

		quoted := func(value string) string {
			if strings.HasPrefix(value, `"`) {
				return value
			}

			return `"` + value + `"`
		}

		good, bad := quoted(kind.good), quoted(kind.bad)

		for _, tc := range []struct {
			name   string
			label  string
			labels []string
		}{
			{"every label good", "alpha=" + good + " beta=" + good, nil},
			{"one bad among good ones", "alpha=" + good + " beta=" + bad + " gamma=" + good, []string{"beta"}},
			{"two bad", "gamma=" + bad + " alpha=" + bad + " beta=" + good, []string{"alpha", "gamma"}},
			{"all three bad", "beta=" + bad + " gamma=" + bad + " alpha=" + bad, []string{"alpha", "beta", "gamma"}},
			{"one label bad twice", "alpha=" + bad + " alpha=" + bad, []string{"alpha"}},
			{"a label outside the schema", "other=" + bad + " alpha=" + good, nil},
		} {
			t.Run(kind.name+"/"+tc.name, func(t *testing.T) {
				t.Parallel()

				var labels []string

				for _, v := range testutils.LintDockerfile("FROM a\nLABEL "+tc.label+"\n", []rule.Rule{kind.rule(cfg)}) {
					if v.Line != 2 {
						t.Errorf("a failure on line %d: %v", v.Line, v)
					}

					for _, name := range []string{"alpha", "beta", "gamma"} {
						if strings.Contains(v.Message, "`"+name+"`") {
							labels = append(labels, name)
						}
					}
				}

				if !slices.Equal(labels, tc.labels) {
					t.Errorf("LABEL %s: the labels reported are %q, want %q", tc.label, labels, tc.labels)
				}
			})
		}
	}
}
