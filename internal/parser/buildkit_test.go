package parser_test

import (
	"errors"
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
