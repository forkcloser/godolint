package parser_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/forkcloser/godolint/internal/parser"
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
