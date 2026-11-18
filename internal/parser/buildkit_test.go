package parser_test

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/forkcloser/godolint/internal/parser"
	"github.com/forkcloser/godolint/internal/syntax"
)

// TestParseRefusesWhatItCannotConvert: an instruction the converter refuses
// is a parse error naming its line, where it was skipped and the rest
// linted as if it were not there.
func TestParseRefusesWhatItCannotConvert(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		dockerfile string
		sentinel   error
		line       string
	}{
		{"FROM without an image", "FROM debian\nFROM\n", parser.ErrFromMissingImage, "line 2:"},
		{"an unknown instruction", "FROM debian\nFOO bar\n", parser.ErrUnknownInstruction, "line 2:"},
		{"COPY with one argument", "FROM debian\nCOPY a\n", parser.ErrCopyMissingArgs, "line 2:"},
		{"ARG without a name", "ARG\n", parser.ErrArgMissingName, "line 1:"},
		{"ONBUILD wrapping nothing the converter knows", "FROM debian\nONBUILD FOO bar\n", parser.ErrUnknownInstruction, "line 2:"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			instructions, err := parser.NewBuildkitParser().Parse([]byte(tc.dockerfile))
			if !errors.Is(err, tc.sentinel) || instructions != nil {
				t.Fatalf("Parse() = %v, %v; want %v", instructions, err, tc.sentinel)
			}

			if !strings.Contains(err.Error(), tc.line) {
				t.Errorf("Parse() error = %q; want it to name %s", err, tc.line)
			}
		})
	}
}

// TestParseKnowsEveryInstruction: a Dockerfile using each of the
// instructions the format defines parses whole.
func TestParseKnowsEveryInstruction(t *testing.T) {
	t.Parallel()

	dockerfile := `# syntax check
ARG BASE=debian
FROM ${BASE}:12 AS build
MAINTAINER someone
LABEL a="b"
ENV X=1
ARG Y
RUN echo hi
COPY a b
ADD c d
WORKDIR /app
USER nobody
EXPOSE 80
VOLUME /data
SHELL ["/bin/sh", "-c"]
STOPSIGNAL SIGTERM
HEALTHCHECK CMD true
ONBUILD RUN echo onbuild
ENTRYPOINT ["/app/run"]
CMD ["--help"]
`

	instructions, err := parser.NewBuildkitParser().Parse([]byte(dockerfile))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Every line of the Dockerfile, the comment included.
	if want := strings.Count(dockerfile, "\n"); len(instructions) != want {
		t.Errorf("Parse() = %d instructions, want %d", len(instructions), want)
	}
}

// TestParseFromSplitsTheImageReference: the image of a FROM is its name
// (registry and port included), its tag and its digest, apart. The tag
// follows the last colon after the last slash, so a registry's port is not
// one; the digest follows the @ and is not part of the name; and a ${...}
// expansion is opaque, its colons, slashes and @ splitting nothing.
func TestParseFromSplitsTheImageReference(t *testing.T) {
	t.Parallel()

	const digest = "sha256:" +
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	for _, tc := range []struct {
		reference          string
		image, tag, digest string
	}{
		{"debian", "debian", "", ""},
		{"debian:12", "debian", "12", ""},
		{"debian@" + digest, "debian", "", digest},
		{"debian:12@" + digest, "debian", "12", digest},
		{"library/debian:12", "library/debian", "12", ""},
		{"myreg.io/team/tools", "myreg.io/team/tools", "", ""},
		{"myreg.io/team/tools:1.0", "myreg.io/team/tools", "1.0", ""},
		{"myreg.io:5000/team/tools", "myreg.io:5000/team/tools", "", ""},
		{"myreg.io:5000/team/tools:1.0", "myreg.io:5000/team/tools", "1.0", ""},
		{"myreg.io:5000/team/tools:1.0@" + digest, "myreg.io:5000/team/tools", "1.0", digest},
		{"myreg.io:5000/team/tools@" + digest, "myreg.io:5000/team/tools", "", digest},
		{"localhost:5000/tools", "localhost:5000/tools", "", ""},
		{"localhost:5000/tools:1.0", "localhost:5000/tools", "1.0", ""},
		{"${BASE}", "${BASE}", "", ""},
		{"${BASE}:12", "${BASE}", "12", ""},
		{"$BASE:$TAG", "$BASE", "$TAG", ""},
		{"${BASE:-debian}", "${BASE:-debian}", "", ""},
		{"${BASE:-debian:12}", "${BASE:-debian:12}", "", ""},
		{"${BASE:-debian}:12", "${BASE:-debian}", "12", ""},
		{"${REGISTRY:-docker.io}/library/debian", "${REGISTRY:-docker.io}/library/debian", "", ""},
		{"${REGISTRY:-myreg.io:5000}/tools:1.0", "${REGISTRY:-myreg.io:5000}/tools", "1.0", ""},
		{"debian:${TAG:-12}", "debian", "${TAG:-12}", ""},
		{"debian:${TAG:-a/b}", "debian", "${TAG:-a/b}", ""},
		{"debian:${TAG:-${OTHER:-12}}", "debian", "${TAG:-${OTHER:-12}}", ""},
		{"${BASE:-debian@" + digest + "}", "${BASE:-debian@" + digest + "}", "", ""},
		{"debian${DIGEST:+@$DIGEST}", "debian${DIGEST:+@$DIGEST}", "", ""},
		{"${BASE:-${OTHER:-debian}:12}", "${BASE:-${OTHER:-debian}:12}", "", ""},
		{"debian:${TAG", "debian", "${TAG", ""},
		{"de{bian:12", "de{bian", "12", ""},
	} {
		t.Run(tc.reference, func(t *testing.T) {
			t.Parallel()

			instructions, err := parser.NewBuildkitParser().Parse([]byte("FROM " + tc.reference + " AS stage\n"))
			if err != nil || len(instructions) != 1 {
				t.Fatalf("Parse() = %v, %v; want one instruction", instructions, err)
			}

			from, ok := instructions[0].Instruction.(*syntax.From)
			if !ok {
				t.Fatalf("parsed as %T; want a FROM", instructions[0].Instruction)
			}

			deref := func(s *string) string {
				if s == nil {
					return ""
				}

				return *s
			}

			got := [3]string{from.Image.Image, deref(from.Image.Tag), deref(from.Image.Digest)}
			if want := [3]string{tc.image, tc.tag, tc.digest}; got != want {
				t.Errorf("image, tag, digest = %q; want %q", got, want)
			}

			if deref(from.Image.Alias) != "stage" {
				t.Errorf("alias %q; want stage", deref(from.Image.Alias))
			}
		})
	}
}

// TestParsePairs: the keys and values of ENV and LABEL, as they are meant:
// quotes of either kind gone wherever they stand, a pair after a value that
// is quoted in part not lost, an escaped quote or blank itself, and the old
// `key value` form taking the rest of the line.
func TestParsePairs(t *testing.T) {
	t.Parallel()

	type pair = [2]string

	for _, tc := range []struct {
		name string
		line string
		want []pair
	}{
		{"plain", `A=1 B=two`, []pair{{"A", "1"}, {"B", "two"}}},
		{"double quotes", `A="x y" B=2`, []pair{{"A", "x y"}, {"B", "2"}}},
		{"single quotes", `A='x y' B='2'`, []pair{{"A", "x y"}, {"B", "2"}}},
		{"empty values", `A="" B='' C=3`, []pair{{"A", ""}, {"B", ""}, {"C", "3"}}},
		{"quoted in part, then another pair", `PATH="/opt/bin":$PATH NEXT=1`, []pair{{"PATH", "/opt/bin:$PATH"}, {"NEXT", "1"}}},
		{"a quoted start and a bare end", `A="x"y B=$A`, []pair{{"A", "xy"}, {"B", "$A"}}},
		{"a bare start and a quoted end", `A=x"y z" B=1`, []pair{{"A", "xy z"}, {"B", "1"}}},
		{"both kinds in one value", `A="x "'y z' B=1`, []pair{{"A", "x y z"}, {"B", "1"}}},
		{"a quoted key", `"com.example.vendor"="ACME Incorporated" 'k'=v`, []pair{{"com.example.vendor", "ACME Incorporated"}, {"k", "v"}}},
		{"an escaped double quote", `A="say \"hi\"" B=1`, []pair{{"A", `say "hi"`}, {"B", "1"}}},
		{"a double quote inside single quotes", `A='say "hi"'`, []pair{{"A", `say "hi"`}}},
		{"an escaped blank", `A=x\ y B=1`, []pair{{"A", "x y"}, {"B", "1"}}},
		{"an escaped tab", "A=b\\\tc B=1", []pair{{"A", "b\tc"}, {"B", "1"}}},
		{"a double quote escaped inside single quotes is not", `A='x\"y'`, []pair{{"A", `x\"y`}}},
		{"a single quote escaped inside double quotes is not", `A="x\'y"`, []pair{{"A", `x\'y`}}},
		{"a backslash that escapes nothing", `A=C:\dir\file B="C:\dir"`, []pair{{"A", `C:\dir\file`}, {"B", `C:\dir`}}},
		{"the old form", `key the rest of the line`, []pair{{"key", "the rest of the line"}}},
		{"the old form, quoted", `version "1.0.0"`, []pair{{"version", "1.0.0"}}},
		{"the old form, quoted in part", `maintainer "Jane Doe" <jane@example.com>`, []pair{{"maintainer", "Jane Doe <jane@example.com>"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			instructions, err := parser.NewBuildkitParser().
				Parse([]byte("FROM a\nENV " + tc.line + "\nLABEL " + tc.line + "\n"))
			if err != nil || len(instructions) != 3 {
				t.Fatalf("Parse() = %v, %v; want three instructions", instructions, err)
			}

			var env, label []pair

			for _, kv := range instructions[1].Instruction.(*syntax.Env).Pairs {
				env = append(env, pair{kv.Key, kv.Value})
			}

			for _, kv := range instructions[2].Instruction.(*syntax.Label).Pairs {
				label = append(label, pair{kv.Key, kv.Value})
			}

			if !slices.Equal(env, tc.want) {
				t.Errorf("ENV pairs %q, want %q", env, tc.want)
			}

			if !slices.Equal(label, tc.want) {
				t.Errorf("LABEL pairs %q, want %q", label, tc.want)
			}
		})
	}
}

// TestParseOnBuildBlank: the blank between ONBUILD and the instruction it
// wraps is any of those buildkit splits an instruction from its arguments
// by, a tab among them.
func TestParseOnBuildBlank(t *testing.T) {
	t.Parallel()

	for name, line := range map[string]string{
		"a space":     "ONBUILD RUN true",
		"a tab":       "ONBUILD\tRUN true",
		"a form feed": "ONBUILD\fRUN true",
		"several":     "ONBUILD \t  RUN true",
		"lower case":  "onbuild\tRUN true",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			instructions, err := parser.NewBuildkitParser().Parse([]byte("FROM a\n" + line + "\n"))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(instructions) != 2 {
				t.Fatalf("%d instructions, want the FROM and the ONBUILD", len(instructions))
			}

			onBuild, wrapped := instructions[1].Instruction.(*syntax.OnBuild)
			if !wrapped {
				t.Fatalf("the second instruction is a %T, want an ONBUILD", instructions[1].Instruction)
			}

			if run, isRun := onBuild.Inner.(*syntax.Run); !isRun || run.Command != "true" {
				t.Errorf("the ONBUILD wraps %#v, want RUN true", onBuild.Inner)
			}
		})
	}
}

// TestParseRunHeredoc: the script of a RUN written with a heredoc. The body
// is the script where the heredoc is the whole command, and never an
// instruction; where a command stands before the heredoc, the script is that
// command and the body is its input; a body that names an interpreter that
// is no shell is no script.
func TestParseRunHeredoc(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		dockerfile string
		// commands is the Command of every RUN, ONBUILD's included, in order.
		commands []string
		// instructions is how many instructions the file holds.
		instructions int
	}{
		{"the body is the script", "FROM a\nRUN <<EOF\napt-get update\nsudo true\nEOF\n", []string{"apt-get update\nsudo true"}, 2},
		{
			"body lines that look like instructions are the script's",
			"FROM a\nRUN <<EOF\nFROM b\nMAINTAINER x\nRUN sudo true\nEOF\nRUN true\n",
			[]string{"FROM b\nMAINTAINER x\nRUN sudo true", "true"},
			3,
		},
		{"a dash and tabs", "FROM a\nRUN <<-EOF\n\tsudo true\n\t\tsudo false\n\tEOF\n", []string{"sudo true\nsudo false"}, 2},
		{"a delimiter in double quotes", "FROM a\nRUN <<\"EOF\"\nsudo $x\nEOF\n", []string{"sudo $x"}, 2},
		{"a delimiter in single quotes", "FROM a\nRUN <<'EOT'\nsudo $x\nEOT\n", []string{"sudo $x"}, 2},
		{"a flag before the heredoc", "FROM a\nRUN --network=none <<EOF\nsudo true\nEOF\n", []string{"sudo true"}, 2},
		{"a shell's shebang", "FROM a\nRUN <<EOF\n#!/bin/bash -e\nsudo true\nEOF\n", []string{"#!/bin/bash -e\nsudo true"}, 2},
		{"a shell's shebang through env", "FROM a\nRUN <<EOF\n#!/usr/bin/env -S bash -e\nsudo true\nEOF\n", []string{"#!/usr/bin/env -S bash -e\nsudo true"}, 2},
		{"another interpreter's shebang", "FROM a\nRUN <<EOF\n#!/usr/bin/env python3\nsudo = 1\nEOF\n", []string{""}, 2},
		{"another interpreter's shebang, by path", "FROM a\nRUN <<EOF\n#!/usr/bin/perl\nsudo();\nEOF\n", []string{""}, 2},
		{"a command fed the heredoc", "FROM a\nRUN python3 <<EOF\nsudo = 1\nEOF\n", []string{"python3"}, 2},
		{
			"a command before the heredoc, more after it",
			"FROM a\nRUN sudo true && cat <<EOF > /etc/motd\nsudo false\nEOF\n",
			[]string{"sudo true && cat"},
			2,
		},
		{"a heredoc on a descriptor", "FROM a\nRUN cat 3<<EOF\nsudo false\nEOF\n", []string{"cat"}, 2},
		{"two heredocs and no command", "FROM a\nRUN <<ONE <<TWO\nsudo one\nONE\nsudo two\nTWO\n", []string{""}, 2},
		{"a heredoc sent to a file", "FROM a\nRUN <<EOF > /out\nsudo true\nEOF\n", []string{""}, 2},
		{"a heredoc piped on", "FROM a\nRUN <<EOF | sh\nsudo true\nEOF\n", []string{""}, 2},
		{"a here-string is no heredoc", "FROM a\nRUN cat <<<hello\n", []string{"cat <<<hello"}, 2},
		{"under ONBUILD", "FROM a\nONBUILD RUN <<EOF\nsudo true\nEOF\nRUN true\n", []string{"sudo true", "true"}, 3},
		{"under ONBUILD, a command before it", "FROM a\nONBUILD RUN sudo true && cat <<EOF\nx\nEOF\n", []string{"sudo true && cat"}, 2},
		{"a COPY and an ADD of a heredoc", "FROM a\nCOPY <<EOF /etc/motd\nFROM b\nRUN sudo true\nEOF\nADD <<EOF /x\nhello\nEOF\n", nil, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			instructions, err := parser.NewBuildkitParser().Parse([]byte(tc.dockerfile))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			var commands []string

			for _, instruction := range instructions {
				inner := instruction.Instruction
				if onBuild, wrapped := inner.(*syntax.OnBuild); wrapped {
					inner = onBuild.Inner
				}

				if run, isRun := inner.(*syntax.Run); isRun {
					commands = append(commands, run.Command)
				}
			}

			if !slices.Equal(commands, tc.commands) {
				t.Errorf("the RUNs' commands are %q, want %q", commands, tc.commands)
			}

			if len(instructions) != tc.instructions {
				t.Errorf("%d instructions, want %d: %v", len(instructions), tc.instructions, instructions)
			}
		})
	}
}

// TestParseRunHeredocShells: a body that names one of the shells as its
// interpreter is the script, shebang line and all, under any of their names
// and at any path.
func TestParseRunHeredocShells(t *testing.T) {
	t.Parallel()

	for _, shell := range []string{"sh", "bash", "ash", "dash", "ksh", "zsh"} {
		for _, shebang := range []string{
			"#!/bin/" + shell, "#!/usr/local/bin/" + shell + " -eu", "#!/usr/bin/env " + shell, "#!/bin/busybox " + shell,
		} {
			want := shebang + "\nsudo true"

			instructions, err := parser.NewBuildkitParser().Parse([]byte("FROM a\nRUN <<EOF\n" + want + "\nEOF\n"))
			if err != nil || len(instructions) != 2 {
				t.Fatalf("%s: Parse() = %v, %v; want two instructions", shebang, instructions, err)
			}

			run, isRun := instructions[1].Instruction.(*syntax.Run)
			if !isRun || run.Command != want {
				t.Errorf("%s: the RUN's command is %q, want %q", shebang, run.Command, want)
			}
		}
	}
}

// TestParseCommentsStandOnTheirLines: every comment of the file is an
// instruction on the line it is written on, whatever stands between it and
// the next instruction, before the first one and after the last; a `#` line
// among the lines of an instruction, in a continuation or a heredoc, is not
// one.
func TestParseCommentsStandOnTheirLines(t *testing.T) {
	t.Parallel()

	dockerfile := strings.Join([]string{
		"# syntax=docker/dockerfile:1", // 1
		"# top",                        // 2
		"#",                            // 3
		"# above from",                 // 4
		"FROM a",                       // 5
		"",                             // 6
		"## two hashes",                // 7
		"",                             // 8
		"RUN true \\",                  // 9
		"  # inside a continuation",    // 10
		"  && true",                    // 11
		"RUN <<EOF",                    // 12
		"# in a heredoc",               // 13
		"EOF",                          // 14
		"  # indented",                 // 15
		"COPY <<EOF /x",                // 16
		"# in a copy's heredoc",        // 17
		"EOF",                          // 18
		"# after the last",             // 19
		"",
	}, "\n")

	instructions, err := parser.NewBuildkitParser().Parse([]byte(dockerfile))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	var got []string

	for _, instruction := range instructions {
		text := instruction.Instruction.Name()
		if comment, isComment := instruction.Instruction.(*syntax.Comment); isComment {
			text = "#" + comment.Text
		}

		got = append(got, strconv.Itoa(instruction.LineNumber)+" "+text)
	}

	want := []string{
		"1 #syntax=docker/dockerfile:1", "2 #top", "3 #", "4 #above from", "5 FROM",
		"7 ## two hashes", "9 RUN", "12 RUN", "15 #indented", "16 COPY", "19 #after the last",
	}

	if !slices.Equal(got, want) {
		t.Errorf("the instructions are\n%q\nwant\n%q", got, want)
	}
}
