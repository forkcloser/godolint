package rules_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3060Hand holds what hadolint's cases leave open about the stages: a
// RUN before any FROM, a stage with no alias that nothing can be built on,
// two stages written alike, which are one to hadolint, and the order of the
// failures.
func TestDL3060Hand(t *testing.T) {
	t.Parallel()

	allRules := []rule.Rule{rules.DL3060()}

	for _, tc := range []struct {
		name       string
		dockerfile string
		lines      []int
	}{
		{"no FROM at all", "RUN yarn install", []int{1}},
		{"a RUN before the first FROM", "RUN yarn install\nFROM node\nRUN true", nil},
		{"a RUN before the first FROM, the last stage FROM scratch", "RUN yarn install\nFROM scratch\nRUN true", []int{1}},
		{"a RUN before the first FROM, the last stage scratch under an alias", "RUN yarn install\nFROM scratch AS base", nil},
		{"an image behind a registry is not the stage of that name", "FROM node AS deps\nRUN yarn install\nFROM reg.io/deps\nRUN true", nil},
		{"a stage with no alias, then another image", "FROM node\nRUN yarn install\nFROM scratch", nil},
		{"two stages written alike", "FROM node\nRUN yarn install\nFROM node\nRUN true", []int{2}},
		{"the same image under another tag", "FROM node:20\nRUN yarn install\nFROM node:22\nRUN true", nil},
		{"the same image under another alias", "FROM node AS one\nRUN yarn install\nFROM node AS two\nRUN true", nil},
		{
			"the same image for another platform",
			"FROM --platform=linux/amd64 node\nRUN yarn install\nFROM --platform=linux/arm64 node\nRUN true",
			nil,
		},
		{
			"the same image by another digest",
			"FROM node@sha256:" + strings.Repeat("a", 64) + "\nRUN yarn install\nFROM node@sha256:" +
				strings.Repeat("b", 64) + "\nRUN true",
			nil,
		},
		{
			"two stages written alike, by digest",
			"FROM node@sha256:" + strings.Repeat("a", 64) + "\nRUN yarn install\nFROM node@sha256:" +
				strings.Repeat("a", 64) + "\nRUN true",
			[]int{2},
		},
		{"a stage built on by its alias", "FROM node AS deps\nRUN yarn install\nFROM deps\nRUN true", []int{2}},
		{"a stage with an alias nothing is built on", "FROM node AS deps\nRUN yarn install\nFROM other\nRUN true", nil},
		{"the last stage", "FROM other AS first\nRUN true\nFROM node\nRUN yarn install", []int{4}},
		{"two RUNs of the last stage, in their order", "FROM node\nRUN yarn install a\nRUN true\nRUN yarn install b", []int{2, 4}},
		{
			"an earlier stage built on and the last one",
			"FROM node AS deps\nRUN yarn install\nFROM deps\nRUN yarn install more",
			[]int{2, 4},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var lines []int
			for _, v := range testutils.LintDockerfile(tc.dockerfile, allRules) {
				lines = append(lines, v.Line)
			}

			if !slices.Equal(lines, tc.lines) {
				t.Errorf("reported at lines %v, want %v", lines, tc.lines)
			}
		})
	}
}
