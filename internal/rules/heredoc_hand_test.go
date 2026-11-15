package rules_test

import (
	"slices"
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestRunHeredocHand: the rules read a RUN written with a heredoc as they
// read any RUN, which hadolint's cases do not cover. Each row is hadolint's
// verdict, read from its parser (language-docker's Parser/Arguments.hs) and
// the rule.
func TestRunHeredocHand(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		rule       func() rule.Rule
		dockerfile string
		lines      []int
	}{
		{"sudo in the body", rules.DL3004, "FROM a\nRUN <<EOF\nsudo true\nEOF\n", []int{2}},
		{"sudo in the body, under ONBUILD", rules.DL3004, "FROM a\nONBUILD RUN <<EOF\nsudo true\nEOF\n", []int{2}},
		{"an unpinned package in the body", rules.DL3008, "FROM a\nRUN <<EOF\napt-get update\napt-get install curl\nEOF\n", []int{2}},
		{"a pipe in the body", rules.DL4006, "FROM a\nRUN <<EOF\nwget -O - https://example.com/x | sh\nEOF\n", []int{2}},
		{"two RUNs of one command each", rules.DL3059, "FROM a\nRUN <<EOF\necho a\nEOF\n\nRUN <<EOF\necho b\nEOF\n", []int{6}},
		{"sudo before a heredoc of data", rules.DL3004, "FROM a\nRUN sudo true && cat > /etc/motd <<EOF\nhello\nEOF\n", []int{2}},
		{"sudo in a heredoc of data", rules.DL3004, "FROM a\nRUN cat > /etc/sudoers.d/x <<EOF\nsudo ALL\nEOF\n", nil},
		{"sudo in what another program reads", rules.DL3004, "FROM a\nRUN python3 <<EOF\nsudo = 1\nEOF\n", nil},
		{"sudo in a script of another interpreter", rules.DL3004, "FROM a\nRUN <<EOF\n#!/usr/bin/env python3\nsudo = 1\nEOF\n", nil},
		{"sudo after the heredoc, in the next RUN", rules.DL3004, "FROM a\nRUN <<EOF\ntrue\nEOF\nRUN sudo true\n", []int{5}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var lines []int
			for _, v := range testutils.LintDockerfile(tc.dockerfile, []rule.Rule{tc.rule()}) {
				lines = append(lines, v.Line)
			}

			if !slices.Equal(lines, tc.lines) {
				t.Errorf("reported at lines %v, want %v", lines, tc.lines)
			}
		})
	}
}
