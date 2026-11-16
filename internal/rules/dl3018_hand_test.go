package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3018Hand holds what hadolint's cases leave open about the flags of
// apk add that take a value: written apart, the value is no package;
// written with `=`, the flag takes nothing after it and the next argument
// is a package like any other.
func TestDL3018Hand(t *testing.T) {
	t.Parallel()

	const repo = "https://dl-cdn.alpinelinux.org/alpine/edge/testing"

	for _, tc := range []struct {
		name     string
		run      string
		reported bool
	}{
		{"a repository given with =, an unpinned package", "apk add --no-cache --repository=" + repo + " hugo", true},
		{"a repository given with =, a pinned package", "apk add --no-cache --repository=" + repo + " hugo=0.1", false},
		{"a repository given apart, an unpinned package", "apk add --no-cache --repository " + repo + " hugo", true},
		{"a repository given apart, a pinned package", "apk add --no-cache --repository " + repo + " hugo=0.1", false},
		{"a repository given apart, and nothing else", "apk add --no-cache --repository " + repo, false},
		{"a virtual name given with =", "apk add --virtual=.build-deps gcc", true},
		{"a virtual name given apart", "apk add --virtual .build-deps gcc=1", false},
		{"a virtual name given apart is no package", "apk add -t .build-deps gcc=1", false},
		{"-X given apart", "apk add -X " + repo + " hugo", true},
		{"-X given apart, a pinned package", "apk add -X " + repo + " hugo=0.1", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			violations := testutils.LintDockerfile("FROM alpine:3.20\nRUN "+tc.run+"\n", []rule.Rule{rules.DL3018()})
			if reported := len(violations) != 0; reported != tc.reported {
				t.Errorf("reported %v, want %v: %v", reported, tc.reported, violations)
			}
		})
	}
}
