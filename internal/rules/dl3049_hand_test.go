package rules_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// dl3049Lines is the lines DL3049 reports on dockerfile with the labels
// required, in order.
func dl3049Lines(dockerfile string, labels ...string) []int {
	cfg := &config.Config{LabelSchema: map[string]config.LabelType{}}
	for _, label := range labels {
		cfg.LabelSchema[label] = config.LabelTypeRawText
	}

	var lines []int

	for _, v := range testutils.LintDockerfile(dockerfile, []rule.Rule{rules.DL3049WithConfig(cfg)}) {
		if v.Code == "DL3049" {
			lines = append(lines, v.Line)
		}
	}

	slices.Sort(lines)

	return lines
}

// TestDL3049Stages holds, for hadolint's multi-stage cases, what the
// generated tests do not: the spec counts the failures, and these are the
// lines the count falls on, the FROM of every stage the label is missing
// from. The label required is foo, as in the spec.
func TestDL3049Stages(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		dockerfile string
		lines      []int
	}{
		{"warn twice: two stages, no labels", "FROM stage1\nFROM stage2", []int{1, 2}},
		{
			"warn twice: two stages, wrong labels only",
			"FROM stage1\nLABEL bar=\"baz\"\nFROM stage2\nLABEL buzz=\"fuzz\"",
			[]int{1, 3},
		},
		{
			"warn once: two stages, label present in first only",
			"FROM baseimage\nLABEL foo=\"bar\"\nFROM newimage",
			[]int{3},
		},
		{
			"warn once: two stages, label present in second only",
			"FROM baseimage\nFROM newimage\nLABEL foo=\"bar\"",
			[]int{1},
		},
		{
			"warn once: two stages, no inheritance, wrong label in one",
			"FROM baseimage\nLABEL baz=\"bar\"\nFROM newimage\nLABEL foo=\"bar\"",
			[]int{1},
		},
		{"warn twice: two stages, inheritance, no labels", "FROM baseimage as base\nFROM base", []int{1, 2}},
		{
			"don't warn: two stages, inheritance, label defined in second stage",
			"FROM baseimage as base\nFROM base\nLABEL foo=\"bar\"",
			nil,
		},
		{
			"don't warn: two stages, no inheritance, label in last stage, with copy",
			"FROM baseimage as base\nFROM anotherimage\nLABEL foo=\"bar\"\nCOPY --from=base /bar /baz",
			nil,
		},
		{
			"warn once: two stages, no inheritance, label in last stage, no connection",
			"FROM baseimage as base\nFROM anotherimage\nLABEL foo=\"bar\"",
			[]int{1},
		},
		{"don't warn: two stages, inheritance", "FROM baseimage as base\nLABEL foo=\"bar\"\nFROM base", nil},
		{
			"warn once: multiple results, no inheritance, label in one result",
			"FROM baseimage as base\nFROM newimage\nLABEL foo=\"bar\"\nCOPY --from=base /bar /baz\n\n" +
				"FROM newimage\nCOPY --from=base /bar /baz",
			[]int{6},
		},
		{
			"warn once: multiple results, inheritance, label in the first result",
			"FROM baseimage as base\nFROM base\nLABEL foo=\"bar\"\n\nFROM base",
			[]int{5},
		},
		{
			"warn once: multiple results, inheritance, label in the second result",
			"FROM baseimage as base\nFROM base\n\nFROM base\nLABEL foo=\"bar\"",
			[]int{2},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := dl3049Lines(tc.dockerfile, "foo"); !slices.Equal(got, tc.lines) {
				t.Errorf("reported at lines %v, want %v", got, tc.lines)
			}
		})
	}
}

// TestDL3049Hand holds what hadolint's cases leave open: a label is inherited
// down a chain of stages; a stage with an alias of its own that sets the
// label does not excuse the stage it is built on, since it is no longer
// called by that stage's name; COPY --from excuses a stage by its alias only;
// each required label is followed on its own; an image behind a registry
// inherits from no stage; and with no stage there is nothing to report.
func TestDL3049Hand(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		dockerfile string
		labels     []string
		lines      []int
	}{
		{
			"inherited through two stages",
			"FROM baseimage AS a\nLABEL foo=\"bar\"\nFROM a AS b\nFROM b",
			[]string{"foo"},
			nil,
		},
		{
			"the chain broken by a stage that does not inherit",
			"FROM baseimage AS a\nLABEL foo=\"bar\"\nFROM other AS b\nFROM b",
			[]string{"foo"},
			[]int{3, 4},
		},
		{
			"an aliased child that sets the label leaves its parent reported",
			"FROM baseimage AS base\nFROM base AS child\nLABEL foo=\"bar\"",
			[]string{"foo"},
			[]int{1},
		},
		{
			"COPY --from by index excuses nothing",
			"FROM baseimage\nFROM anotherimage\nLABEL foo=\"bar\"\nCOPY --from=0 /bar /baz",
			[]string{"foo"},
			[]int{1},
		},
		{
			"COPY --from of another alias excuses nothing",
			"FROM baseimage AS base\nFROM tools AS tools\nFROM anotherimage\nLABEL foo=\"bar\"\nCOPY --from=tools /a /b",
			[]string{"foo"},
			[]int{1},
		},
		{
			"two labels, each on its own stages",
			"FROM baseimage AS base\nLABEL foo=\"1\"\nFROM base\nLABEL bar=\"2\"",
			[]string{"foo", "bar"},
			nil,
		},
		{
			"a stage inherits the label its parent has, not the other one",
			"FROM baseimage AS base\nLABEL foo=\"1\"\nFROM base",
			[]string{"foo", "bar"},
			[]int{1, 3},
		},
		{
			"a child that sets one label excuses its parent for that one only",
			"FROM baseimage AS base\nFROM base\nLABEL foo=\"1\"",
			[]string{"foo", "bar"},
			[]int{1, 2},
		},
		{
			"two labels, one missing everywhere and one in the last stage only",
			"FROM baseimage\nFROM newimage\nLABEL bar=\"2\"",
			[]string{"foo", "bar"},
			[]int{1, 1, 2},
		},
		{
			"a stage copied from is excused for every label",
			"FROM baseimage AS base\nFROM anotherimage\nLABEL foo=\"1\" bar=\"2\"\nCOPY --from=base /a /b",
			[]string{"foo", "bar"},
			nil,
		},
		{
			"an image behind a registry is not the stage of that name",
			"FROM baseimage AS base\nLABEL foo=\"bar\"\nFROM myreg.io/base",
			[]string{"foo"},
			[]int{3},
		},
		{
			"a stage built by digest on an earlier one inherits its label",
			"FROM baseimage AS base\nLABEL foo=\"bar\"\nFROM base@sha256:" + strings.Repeat("a", 64),
			[]string{"foo"},
			nil,
		},
		{
			"a stage built by tag and digest on an earlier one inherits its label",
			"FROM baseimage AS base\nLABEL foo=\"bar\"\nFROM base:1.0@sha256:" + strings.Repeat("a", 64),
			[]string{"foo"},
			nil,
		},
		{"a label once is enough, with others beside it", "FROM baseimage\nLABEL a=\"1\" foo=\"2\"", []string{"foo"}, nil},
		{"no stage", "ARG VERSION=1", []string{"foo"}, nil},
		{"no label required", "FROM baseimage", nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := dl3049Lines(tc.dockerfile, tc.labels...); !slices.Equal(got, tc.lines) {
				t.Errorf("reported at lines %v, want %v", got, tc.lines)
			}
		})
	}
}
