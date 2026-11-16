package rules_test

import (
	"slices"
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3044Hand holds what hadolint's cases leave open: the names an ENV
// gives are defined after it whether or not the ENV itself was reported.
func TestDL3044Hand(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		dockerfile string
		lines      []int
	}{
		{
			"a name given again after an ENV that was reported",
			"FROM a\nENV APP_HOME=/app PATH=$APP_HOME/bin:$PATH\nENV APP_HOME=/srv LOG_DIR=$APP_HOME/log\n",
			[]int{2},
		},
		{
			"a name given again after an ENV that was in order",
			"FROM a\nENV APP_HOME=/app\nENV APP_HOME=/srv LOG_DIR=$APP_HOME/log\n",
			nil,
		},
		{
			"a name first given beside its use, twice",
			"FROM a\nENV A=1 B=$A\nENV C=1 D=$C\n",
			[]int{2, 3},
		},
		{
			"a name an ARG declared",
			"FROM a\nARG A\nENV A=1 B=$A\n",
			nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var lines []int
			for _, v := range testutils.LintDockerfile(tc.dockerfile, []rule.Rule{rules.DL3044()}) {
				lines = append(lines, v.Line)
			}

			if !slices.Equal(lines, tc.lines) {
				t.Errorf("reported at lines %v, want %v", lines, tc.lines)
			}
		})
	}
}
