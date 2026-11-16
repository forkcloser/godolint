package parser_test

import (
	"slices"
	"testing"

	"github.com/forkcloser/godolint/internal/parser"
	"github.com/forkcloser/godolint/internal/syntax"
)

// TestParseReadsPairsUnderTheEscapeDirective: the character that escapes a
// quote or a blank in an ENV or a LABEL is the Dockerfile's, a backslash
// unless a directive names another, and the other is then no escape.
func TestParseReadsPairsUnderTheEscapeDirective(t *testing.T) {
	t.Parallel()

	const backtick = "# escape=`\nFROM a\n"

	for _, tc := range []struct {
		name       string
		dockerfile string
		pairs      []string
	}{
		{"a quote escaped by a backtick", backtick + "ENV GREETING=\"say `\"hi`\"\"\n", []string{`GREETING=say "hi"`}},
		{"a blank escaped by a backtick", backtick + "ENV Q=a` b R=2\n", []string{"Q=a b", "R=2"}},
		{"a backslash is no escape under a backtick", backtick + "ENV P=C:\\dir\\ Q=\"a\\\"\n", []string{`P=C:\dir\`, `Q=a\`}},
		{"a LABEL", backtick + "LABEL a=\"x `\"y`\"\" `\n  b=2\n", []string{`a=x "y"`, "b=2"}},
		{"an ENV under ONBUILD", backtick + "ONBUILD ENV A=\"x `\" y\" B=2\n", []string{`A=x " y`, "B=2"}},
		{"a quote escaped by a backslash, with no directive", "FROM a\nENV GREETING=\"say \\\"hi\\\"\" Q=a\\ b\n", []string{`GREETING=say "hi"`, "Q=a b"}},
		{"a backtick is no escape with no directive", "FROM a\nENV A=\"x `\" B=2\n", []string{"A=x `", "B=2"}},
		{"an ENV under ONBUILD, with no directive", "FROM a\nONBUILD ENV A=\"x \\\" y\" B=2\n", []string{`A=x " y`, "B=2"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			instructions, err := parser.NewBuildkitParser().Parse([]byte(tc.dockerfile))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}

			var pairs []string

			for _, pos := range instructions {
				instruction := pos.Instruction
				if onBuild, wrapped := instruction.(*syntax.OnBuild); wrapped {
					instruction = onBuild.Inner
				}

				switch typed := instruction.(type) {
				case *syntax.Env:
					for _, pair := range typed.Pairs {
						pairs = append(pairs, pair.Key+"="+pair.Value)
					}
				case *syntax.Label:
					for _, pair := range typed.Pairs {
						pairs = append(pairs, pair.Key+"="+pair.Value)
					}
				}
			}

			if !slices.Equal(pairs, tc.pairs) {
				t.Errorf("pairs %q, want %q", pairs, tc.pairs)
			}
		})
	}
}
