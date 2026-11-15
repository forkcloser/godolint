package process_test

import (
	"slices"
	"testing"

	"github.com/forkcloser/godolint/internal/parser"
	"github.com/forkcloser/godolint/internal/process"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
)

// TestRunAppliesAPragmaFromItsOwnLine: an ignore pragma covers the line
// after the one it is written on and no other, a stage pragma the stage
// whose FROM stands on that line, and a global one the file wherever it
// stands. Each row is hadolint's verdict (Pragma.hs, and run in Process.hs).
func TestRunAppliesAPragmaFromItsOwnLine(t *testing.T) {
	t.Parallel()

	const install = "RUN apt-get install -y foo\n"

	for _, tc := range []struct {
		name       string
		rule       func() rule.Rule
		dockerfile string
		lines      []int
	}{
		{"a pragma right above", rules.DL3008, "FROM a:1\n# hadolint ignore=DL3008\n" + install, nil},
		{"a pragma and a blank line", rules.DL3008, "FROM a:1\n# hadolint ignore=DL3008\n\n" + install, []int{4}},
		{"a pragma and a comment", rules.DL3008, "FROM a:1\n# hadolint ignore=DL3008\n# why\n" + install, []int{4}},
		{"a pragma inside a continuation", rules.DL3008, "FROM a:1\nRUN true \\\n  # hadolint ignore=DL3008\n  && apt-get install -y foo\n", []int{2}},
		{"two hashes are no pragma", rules.DL3006, "## hadolint ignore=DL3006\nFROM a\n", []int{2}},
		{"a global pragma above an empty comment", rules.DL3006, "# hadolint global ignore=DL3006\n#\nFROM a\n", nil},
		{"a global pragma above a blank line", rules.DL3006, "# hadolint global ignore=DL3006\n\nFROM a\n", nil},
		{"a global pragma after the last instruction", rules.DL3006, "FROM a\n# hadolint global ignore=DL3006\n", nil},
		{"a stage pragma right above its FROM", rules.DL3008, "# hadolint stage ignore=DL3008\nFROM a:1\n" + install, nil},
		{
			"a stage pragma and a blank line, the FROM",
			rules.DL3006,
			"# hadolint stage ignore=DL3006,DL3008\n\nFROM a\n" + install,
			nil,
		},
		{
			"a stage pragma and a blank line, the stage",
			rules.DL3008,
			"# hadolint stage ignore=DL3006,DL3008\n\nFROM a\n" + install,
			[]int{4},
		},
		{"the pragma rule names the pragma's own line", rules.DL1001, "# hadolint ignore=DL3006\n\nFROM a\n", []int{1}},
		{"the pragma rule, after the last instruction", rules.DL1001, "FROM a:1\n# hadolint ignore=DL3006\n", []int{2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			instructions, err := parser.NewBuildkitParser().Parse([]byte(tc.dockerfile))
			if err != nil {
				t.Fatal(err)
			}

			var lines []int
			for _, failure := range process.NewProcessor([]rule.Rule{tc.rule()}).Run(instructions) {
				lines = append(lines, failure.Line)
			}

			if !slices.Equal(lines, tc.lines) {
				t.Errorf("reported at lines %v, want %v", lines, tc.lines)
			}
		})
	}
}
