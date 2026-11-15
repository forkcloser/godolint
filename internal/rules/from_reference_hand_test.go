package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestFromReferenceHand holds what the rules make of an image reference
// hadolint's cases do not write: a registry with a port, whose colon is not
// the tag's; a ${...} expansion with a default, whose colon is not either;
// and a stage built on an earlier one by digest, which is still built on it.
func TestFromReferenceHand(t *testing.T) {
	t.Parallel()

	const (
		digest = "@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		apt    = "RUN apt-get update && apt-get install -y curl=1\n"
	)

	for _, tc := range []struct {
		name       string
		rule       rule.Rule
		dockerfile string
		violation  bool
	}{
		{"untagged behind a registry with a port", rules.DL3006(), "FROM myreg.io:5000/team/tools", true},
		{"untagged behind localhost with a port", rules.DL3006(), "FROM localhost:5000/tools", true},
		{"tagged behind a registry with a port", rules.DL3006(), "FROM myreg.io:5000/team/tools:1.0", false},
		{"by digest behind a registry with a port", rules.DL3006(), "FROM myreg.io:5000/team/tools" + digest, false},

		{"latest behind a registry with a port", rules.DL3007(), "FROM myreg.io:5000/team/tools:latest", true},
		{"latest behind localhost with a port", rules.DL3007(), "FROM localhost:5000/tools:latest", true},
		{"a port is not a tag", rules.DL3007(), "FROM myreg.io:5000/latest", false},
		{"latest and a digest behind a registry with a port", rules.DL3007(), "FROM myreg.io:5000/tools:latest" + digest, false},
		{"latest on a variable with a default", rules.DL3007(), "FROM ${BASE:-debian}:latest", true},
		{"latest behind a registry variable with a default", rules.DL3007(), "FROM ${REGISTRY:-docker.io}/debian:latest", true},
		{"latest as a variable's default only", rules.DL3007(), "FROM ${BASE:-debian:latest}", false},

		{
			"a relative COPY in a stage built by digest on one with a WORKDIR", rules.DL3045(),
			"FROM debian:12 AS build\nWORKDIR /x\nFROM build" + digest + "\nCOPY a b",
			false,
		},
		{
			"a relative COPY in a stage built by digest on one without a WORKDIR", rules.DL3045(),
			"FROM debian:12 AS build\nFROM build" + digest + "\nCOPY a b",
			true,
		},
		{
			"apt lists left in a stage another is built on by digest", rules.DL3009(),
			"FROM debian:12 AS build\n" + apt + "FROM build" + digest + "\nRUN true",
			true,
		},
		{
			"apt lists left in a stage nothing is built on", rules.DL3009(),
			"FROM debian:12 AS build\n" + apt + "FROM other" + digest + "\nRUN true",
			false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			violations := testutils.LintDockerfile(tc.dockerfile, []rule.Rule{tc.rule})
			code := string(tc.rule.Code())

			if tc.violation {
				testutils.AssertContainsViolation(t, violations, code)
			} else {
				testutils.AssertNoViolation(t, violations, code)
			}
		})
	}
}
