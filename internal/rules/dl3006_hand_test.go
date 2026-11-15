package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3006Hand holds what hadolint's own cases leave open: a stage's own
// alias is not a previous stage, the shapes a variable can take inside an
// image name, where each literal around the variables must stand in the
// stage's name, and that an image behind a registry is not a stage.
func TestDL3006Hand(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{rules.DL3006()}

	for _, tc := range []struct {
		name       string
		dockerfile string
		violation  bool
	}{
		{"an image named as its own alias", "FROM debian AS debian", true},
		{"a variable matching only its own alias", "FROM app-${X} AS app-final", true},
		{
			"a variable in the middle of a stage name",
			"FROM debian:12 AS base-amd64-slim\nFROM base-${ARCH}-slim",
			false,
		},
		{
			"a suffix that no stage has",
			"FROM debian:12 AS base-amd64-slim\nFROM base-${ARCH}-full",
			true,
		},
		{
			"two variables side by side",
			"FROM debian:12 AS base-amd64-slim\nFROM base-${ARCH}${LIBC}-slim",
			false,
		},
		{
			"a prefix alone is not a match",
			"FROM debian:12 AS base\nFROM base-${ARCH}",
			true,
		},
		{
			"an unbraced variable ends at the first character a name cannot hold",
			"FROM debian:12 AS img1-slim\nFROM img$n-slim",
			false,
		},
		{
			"an unbraced variable's name holds an underscore",
			"FROM debian:12 AS img1-slim\nFROM img$my_var-slim",
			false,
		},
		{
			"an unbraced variable's name holds a digit",
			"FROM debian:12 AS img1-slim\nFROM img$v2-slim",
			false,
		},
		{
			"the first literal starts the stage's name",
			"FROM debian:12 AS xbase-amd64\nFROM base-${ARCH}",
			true,
		},
		{
			"the last literal ends the stage's name",
			"FROM debian:12 AS base-amd64-slim-x\nFROM base-${ARCH}-slim",
			true,
		},
		{
			"a literal between two variables is in the stage's name",
			"FROM debian:12 AS base-amd64-musl-1-slim\nFROM base-${ARCH}-musl-${REV}-slim",
			false,
		},
		{
			"a literal between two variables that the stage's name lacks",
			"FROM debian:12 AS base-amd64-glibc-1-slim\nFROM base-${ARCH}-musl-${REV}-slim",
			true,
		},
		{
			"a literal between two variables is not matched twice",
			"FROM debian:12 AS a-xmid\nFROM a-${X}mid${Y}mid",
			true,
		},
		{
			"a literal between two variables, then the same one at the end",
			"FROM debian:12 AS a-xmidymid\nFROM a-${X}mid${Y}mid",
			false,
		},
		{
			"a registry is part of the name, so no stage matches",
			"FROM debian:12 AS base-amd64\nFROM myreg.io/base-${ARCH}",
			true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			violations := testutils.LintDockerfile(tc.dockerfile, allRules)
			if tc.violation {
				testutils.AssertContainsViolation(t, violations, "DL3006")
			} else {
				testutils.AssertNoViolation(t, violations, "DL3006")
			}
		})
	}
}
