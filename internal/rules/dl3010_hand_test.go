package rules_test

import (
	"slices"
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3010Hand holds what hadolint's cases leave open: which COPY is one
// an ADD could stand for, and how many times a line is reported.
func TestDL3010Hand(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		dockerfile string
		lines      []int
	}{
		{
			"an archive copied and unpacked",
			"FROM a\nCOPY app.tar.gz /tmp/\nRUN tar -xzf /tmp/app.tar.gz -C /opt\n",
			[]int{2},
		},
		{
			"an archive copied from another stage",
			"FROM a AS build\nFROM b\nCOPY --from=build /out/app.tar.gz /tmp/\nRUN tar -xzf /tmp/app.tar.gz -C /opt\n",
			nil,
		},
		{
			"two archives of one COPY, both unpacked",
			"FROM a\nCOPY one.tar.gz two.tar.gz /tmp/\nRUN tar -xzf /tmp/one.tar.gz && tar -xzf /tmp/two.tar.gz\n",
			[]int{2},
		},
		{
			"two archives of two COPYs, both unpacked",
			"FROM a\nCOPY one.tar.gz /tmp/\nCOPY two.tar.gz /tmp/\nRUN tar -xzf /tmp/one.tar.gz && tar -xzf /tmp/two.tar.gz\n",
			[]int{2, 3},
		},
		{
			"an archive copied and left packed",
			"FROM a\nCOPY app.tar.gz /tmp/\nRUN ls /tmp\n",
			nil,
		},
		{
			"an archive unpacked in a stage that is not the last",
			"FROM a AS build\nCOPY app.tar.gz /tmp/\nRUN tar -xzf /tmp/app.tar.gz -C /opt\nFROM b\nCOPY --from=build /opt /opt\n",
			[]int{2},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var lines []int
			for _, v := range testutils.LintDockerfile(tc.dockerfile, []rule.Rule{rules.DL3010()}) {
				lines = append(lines, v.Line)
			}

			slices.Sort(lines)

			if !slices.Equal(lines, tc.lines) {
				t.Errorf("reported at lines %v, want %v", lines, tc.lines)
			}
		})
	}
}
