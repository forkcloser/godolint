package rules_test

import (
	"slices"
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestShellCommandsHand holds what hadolint's cases leave open about what a
// command of a RUN is: a declaration is one, and a wrapper is the command it
// runs.
func TestShellCommandsHand(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		rule       func() rule.Rule
		dockerfile string
		lines      []int
	}{
		{
			"an export chained to a command makes two",
			rules.DL3059,
			"FROM a\nRUN export DEBIAN_FRONTEND=noninteractive && make\nRUN make install\n",
			nil,
		},
		{
			"a declare chained to a command makes two",
			rules.DL3059,
			"FROM a\nRUN make\nRUN declare -x CC=clang; make install\n",
			nil,
		},
		{
			"an export alone is one command",
			rules.DL3059,
			"FROM a\nRUN export A=1\nRUN make\n",
			[]int{3},
		},
		{
			"a RUN with no command, then a RUN",
			rules.DL3059,
			"FROM a\nRUN A=1\nRUN make\n",
			[]int{3},
		},
		{
			"a command timed is one command",
			rules.DL3059,
			"FROM a\nRUN time make\nRUN make install\n",
			[]int{3},
		},
		{
			"busybox wget is wget",
			rules.DL3047,
			"FROM a\nRUN busybox wget https://example.com/app.tgz\n",
			[]int{2},
		},
		{
			"busybox wget with the progress bar off",
			rules.DL3047,
			"FROM a\nRUN busybox wget -q https://example.com/app.tgz\n",
			nil,
		},
		{
			"command pip install is a pip install",
			rules.DL3013,
			"FROM a\nRUN command pip install requests\n",
			[]int{2},
		},
		{
			"the wrapped command's name is not one of its arguments",
			rules.DL3008,
			"FROM a\nRUN command apt-get install -y curl=8.5.0\n",
			nil,
		},
		{
			"exec apt-get install is an apt-get install",
			rules.DL3008,
			"FROM a\nRUN exec apt-get install -y curl\n",
			[]int{2},
		},
		{
			"a wrapper before an option is the wrapper",
			rules.DL3013,
			"FROM a\nRUN command -v pip install\n",
			nil,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var lines []int
			for _, v := range testutils.LintDockerfile(tc.dockerfile, []rule.Rule{tc.rule()}) {
				lines = append(lines, v.Line)
			}

			if !slices.Equal(lines, tc.lines) {
				t.Errorf("reported at lines %v, want %v", lines, tc.lines)
			}
		})
	}
}
