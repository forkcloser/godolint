package pragma_test

import (
	"slices"
	"testing"

	"github.com/forkcloser/godolint/internal/parser"
	"github.com/forkcloser/godolint/internal/pragma"
	"github.com/forkcloser/godolint/internal/process"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestGrammar is hadolint's pragma grammar, one parser at a time: what each
// one reads, and what it refuses as a whole.
func TestGrammar(t *testing.T) {
	t.Parallel()

	type parse func(string) ([]rule.Code, bool)

	parsers := map[string]parse{
		"ignore": pragma.ParseIgnore,
		"stage":  pragma.ParseStageIgnore,
		"global": pragma.ParseGlobalIgnore,
	}

	tests := []struct {
		name   string
		text   string
		parser string
		want   []rule.Code // nil: not a pragma for that parser
	}{
		{"one code", "hadolint ignore=DL3002", "ignore", []rule.Code{"DL3002"}},
		{"two codes", "hadolint ignore=DL3002,SC2086", "ignore", []rule.Code{"DL3002", "SC2086"}},
		{"spaces everywhere", "  hadolint \t ignore = DL3023 , DL3021  ", "ignore", []rule.Code{"DL3023", "DL3021"}},
		{"inline comment", "hadolint ignore=DL3002 # why", "ignore", []rule.Code{"DL3002"}},
		{"inline comment unspaced", "hadolint ignore=DL3002#why", "ignore", []rule.Code{"DL3002"}},
		{"trailing space", "hadolint ignore=DL3002 ", "ignore", []rule.Code{"DL3002"}},
		{"a name outside the alphabet refuses the whole list", "hadolint ignore=crazy,DL3002", "ignore", nil},
		{"trailing text is not a comment", "hadolint ignore=DL3002 foo", "ignore", nil},
		{"empty list", "hadolint ignore=", "ignore", nil},
		{"trailing comma", "hadolint ignore=DL3002,", "ignore", nil},
		{"no space after hadolint", "hadolintignore=DL3002", "ignore", nil},
		{"not a pragma", "just a comment", "ignore", nil},
		{"stage is not a line pragma", "hadolint stage ignore=DL3002", "ignore", nil},
		{"global is not a line pragma", "hadolint global ignore=DL3002", "ignore", nil},
		{"shell is not an ignore", "hadolint shell=powershell", "ignore", nil},
		{"stage", "hadolint stage ignore=DL3011", "stage", []rule.Code{"DL3011"}},
		{"stage spaced", "hadolint  stage  ignore = DL3011 ,DL3002 # x", "stage", []rule.Code{"DL3011", "DL3002"}},
		{"stage needs a space after the keyword", "hadolint stageignore=DL3011", "stage", nil},
		{"line is not a stage pragma", "hadolint ignore=DL3011", "stage", nil},
		{"global", "hadolint global ignore=DL3002", "global", []rule.Code{"DL3002"}},
		{"global spaced", "hadolint global ignore = DL3023 , DL3021", "global", []rule.Code{"DL3023", "DL3021"}},
		{"line is not a global pragma", "hadolint ignore=DL3002", "global", nil},
	}

	for _, tt := range tests {
		t.Run(tt.parser+": "+tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := parsers[tt.parser](tt.text)
			if ok != (tt.want != nil) {
				t.Fatalf("%s(%q): ok = %v, want %v (codes %v)", tt.parser, tt.text, ok, tt.want != nil, got)
			}

			if ok && !slices.Equal(got, tt.want) {
				t.Fatalf("%s(%q) = %v, want %v", tt.parser, tt.text, got, tt.want)
			}
		})
	}
}

// TestScopes is hadolint's PragmaSpec: each pragma kind's reach, through
// the rules the spec reads them with.
func TestScopes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		dockerfile string
		rules      []rule.Rule
		caught     []string
		notCaught  []string
	}{
		// Rules can be ignored with inline comments.
		{
			name: "ignores single rule",
			dockerfile: `FROM ubuntu
# hadolint ignore=DL3002
USER root`,
			rules:     []rule.Rule{rules.DL3002()},
			notCaught: []string{"DL3002"},
		},
		{
			name: "ignores only the given rule",
			dockerfile: `FROM scratch
# hadolint ignore=DL3001
USER root`,
			rules:  []rule.Rule{rules.DL3002()},
			caught: []string{"DL3002"},
		},
		{
			name: "ignores only the given rule, when multiple passed",
			dockerfile: `FROM scratch
# hadolint ignore=DL3001,DL3002
USER root`,
			rules:     []rule.Rule{rules.DL3002()},
			notCaught: []string{"DL3002"},
		},
		{
			name: "ignores the rule only if directly above the instruction",
			dockerfile: `# hadolint ignore=DL3001,DL3002
FROM ubuntu
USER root`,
			rules:  []rule.Rule{rules.DL3002()},
			caught: []string{"DL3002"},
		},
		{
			name: "won't ignore the rule if passed invalid rule names",
			dockerfile: `FROM scratch
# hadolint ignore=crazy,DL3002
USER root`,
			rules:  []rule.Rule{rules.DL3002()},
			caught: []string{"DL3002"},
		},
		{
			name: "ignores multiple rules correctly, even with some extra whitespace",
			dockerfile: `FROM node as foo
# hadolint ignore=DL3023, DL3021
COPY --from=foo bar baz .`,
			rules:     []rule.Rule{rules.DL3023(), rules.DL3021()},
			notCaught: []string{"DL3023", "DL3021"},
		},
		// Rules can be ignored per stage with stage ignore pragma.
		{
			name: "stage: ignore single rule",
			dockerfile: `# hadolint stage ignore=DL3011
FROM ubuntu
EXPOSE 80000`,
			rules:     []rule.Rule{rules.DL3011()},
			notCaught: []string{"DL3011"},
		},
		{
			name: "stage ignore rule, but catch it in different stage 1",
			dockerfile: `# hadolint stage ignore=DL3011
FROM ubuntu
EXPOSE 80000
FROM ubuntu
EXPOSE 80000`,
			rules:  []rule.Rule{rules.DL3011()},
			caught: []string{"DL3011"},
		},
		{
			name: "stage ignore rule, but catch it in different stage 2",
			dockerfile: `FROM ubuntu
EXPOSE 80000
# hadolint stage ignore=DL3011
FROM ubuntu
EXPOSE 80000`,
			rules:  []rule.Rule{rules.DL3011()},
			caught: []string{"DL3011"},
		},
		{
			name: "stage: ignore multiple rules",
			dockerfile: `# hadolint stage ignore=DL3011,DL3002
FROM ubuntu
USER root
EXPOSE 80000`,
			rules:     []rule.Rule{rules.DL3011(), rules.DL3012(), rules.DL3002()},
			notCaught: []string{"DL3012", "DL3011", "DL3002"},
		},
		// Rules can be ignored globally with global ignore pragma.
		{
			name: "global: ignores single rule",
			dockerfile: `# hadolint global ignore=DL3002
FROM ubuntu
USER root`,
			rules:     []rule.Rule{rules.DL3002()},
			notCaught: []string{"DL3002"},
		},
		{
			name: "global: ignores multiple rules correctly, even with some extra whitespace",
			dockerfile: `# hadolint global ignore = DL3023 , DL3021
FROM node as foo
COPY --from=foo bar baz .`,
			rules:     []rule.Rule{rules.DL3023(), rules.DL3021()},
			notCaught: []string{"DL3023", "DL3021"},
		},
		{
			name: "global: anywhere in the file, before or after the line",
			dockerfile: `FROM ubuntu
USER root
# hadolint global ignore=DL3002
EXPOSE 80`,
			rules:     []rule.Rule{rules.DL3002()},
			notCaught: []string{"DL3002"},
		},
		// A comment can follow in the same line as a pragma.
		{
			name: "pragma followed just by space",
			dockerfile: `FROM ubuntu
# hadolint ignore=DL3002
USER root`,
			rules:     []rule.Rule{rules.DL3002()},
			notCaught: []string{"DL3002"},
		},
		{
			name: "ignore pragma with comment",
			dockerfile: `FROM ubuntu
# hadolint ignore=DL3002 # foobar
USER root`,
			rules:     []rule.Rule{rules.DL3002()},
			notCaught: []string{"DL3002"},
		},
		{
			name: "shell pragma with comment",
			dockerfile: `# hadolint shell=powershell # foobar
FROM ubuntu`,
			rules:     []rule.Rule{rules.DL1001()},
			notCaught: []string{"DL1001"},
		},
		// DL1001Spec: the rule reads the line pragma, and only that one.
		{
			name: "DL1001 catches inline ignore pragma",
			dockerfile: `# hadolint ignore=DL3003
RUN foo bar`,
			rules:  []rule.Rule{rules.DL1001()},
			caught: []string{"DL1001"},
		},
		{
			name: "DL1001 does not catch other pragma",
			dockerfile: `# hadolint shell=powershell
RUN foo bar`,
			rules:     []rule.Rule{rules.DL1001()},
			notCaught: []string{"DL1001"},
		},
		{
			name: "DL1001 does not catch a stage or global pragma",
			dockerfile: `# hadolint global ignore=DL3003
# hadolint stage ignore=DL3003
RUN foo bar`,
			rules:     []rule.Rule{rules.DL1001()},
			notCaught: []string{"DL1001"},
		},
		{
			name: "DL1001 does not catch other comment",
			dockerfile: `# foobar
RUN foo bar`,
			rules:     []rule.Rule{rules.DL1001()},
			notCaught: []string{"DL1001"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			violations := testutils.LintDockerfile(tt.dockerfile, tt.rules)

			for _, code := range tt.caught {
				testutils.AssertContainsViolation(t, violations, code)
			}

			for _, code := range tt.notCaught {
				testutils.AssertNoViolation(t, violations, code)
			}
		})
	}
}

// TestStageReach: a stage pragma covers the lines after it up to the next
// FROM, and nothing before it. Line 4, right after the pragma, is hadolint's
// reach exactly: the pragma is keyed on that line and the lookup wants a key
// below the failure's line, so the first line after a pragma is not covered
// — written above a FROM, as documented, the FROM takes that line.
func TestStageReach(t *testing.T) {
	t.Parallel()

	dockerfile := `FROM ubuntu
EXPOSE 80000
# hadolint stage ignore=DL3011
EXPOSE 80001
EXPOSE 80002
FROM ubuntu
EXPOSE 80003`

	violations := testutils.LintDockerfile(dockerfile, []rule.Rule{rules.DL3011()})

	var lines []int

	for _, v := range violations {
		if v.Code == "DL3011" {
			lines = append(lines, v.Line)
		}
	}

	if want := []int{2, 4, 7}; !slices.Equal(lines, want) {
		t.Fatalf("DL3011 reported at lines %v, want %v", lines, want)
	}
}

// TestDisabled: with the pragmas disabled, all three kinds are plain
// comments.
func TestDisabled(t *testing.T) {
	t.Parallel()

	dockerfile := `# hadolint global ignore=DL3011
# hadolint stage ignore=DL3002
FROM ubuntu
# hadolint ignore=DL3002
USER root
EXPOSE 80000`

	instructions, err := parser.NewBuildkitParser().Parse([]byte(dockerfile))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	allRules := []rule.Rule{rules.DL3002(), rules.DL3011()}

	silenced, err := process.NewProcessor(allRules).Run(instructions)
	if err != nil {
		t.Fatal(err)
	}

	testutils.AssertNoViolation(t, silenced, "DL3002")
	testutils.AssertNoViolation(t, silenced, "DL3011")

	loud, err := process.NewProcessor(allRules).WithDisableIgnorePragmas(true).Run(instructions)
	if err != nil {
		t.Fatal(err)
	}

	testutils.AssertContainsViolation(t, loud, "DL3002")
	testutils.AssertContainsViolation(t, loud, "DL3011")
}
