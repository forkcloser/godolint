package rules_test

import (
	"slices"
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestDL3042Hand holds what hadolint's cases leave open: what counts as a
// pip install, a stage starting with what the image it is built on has, and
// PIP_NO_CACHE_DIR set in the RUN's own text.
func TestDL3042Hand(t *testing.T) {
	t.Parallel()

	const (
		install = "RUN pip install foo==1.0\n"
		off     = "ENV PIP_NO_CACHE_DIR=1\n"
	)

	for _, tc := range []struct {
		name       string
		dockerfile string
		lines      []int
	}{
		{"pip as a python module", "FROM a\nRUN python -m pip install foo==1.0\n", []int{2}},
		{"pip as a python3 module", "FROM a\nRUN python3 -m pip install foo==1.0\n", []int{2}},
		{"pip as a module, the cache off", "FROM a\nRUN python -m pip install --no-cache-dir foo==1.0\n", nil},
		{"pip as a module, an option before install", "FROM a\nRUN python -m pip --quiet install foo==1.0\n", []int{2}},
		{"a pip of one python version", "FROM a\nRUN pip3.11 install foo==1.0\n", []int{2}},
		{"pip3 is no python module", "FROM a\nRUN python -m pip3 install foo==1.0\n", nil},
		{"pipenv is not pip", "FROM a\nRUN pipenv install foo\n", nil},
		{"pipenv as a module is not pip", "FROM a\nRUN python -m pipenv install foo\n", nil},
		{"a pip that installs nothing", "FROM a\nRUN pip list\n", nil},

		{"a stage under an alias, built on one with the cache off", "FROM a AS base\n" + off + "FROM base AS builder\n" + install, nil},
		{"a stage under no alias, built on one with the cache off", "FROM a AS base\n" + off + "FROM base\n" + install, nil},
		{"a second stage of an image whose first has the cache off", "FROM a\n" + off + "FROM a AS build\n" + install, nil},
		{"through two stages", "FROM a AS base\n" + off + "FROM base AS mid\nFROM mid AS final\n" + install, nil},
		{"a stage built on another image", "FROM a AS base\n" + off + "FROM b AS other\n" + install, []int{4}},
		{"a stage built on one with the cache on", "FROM a AS base\nFROM base AS builder\n" + install, []int{3}},
		{
			"an alias taken again starts again",
			"FROM a AS base\n" + off + "FROM b AS base\n" + install,
			[]int{4},
		},

		{"set in the RUN, before a semicolon", "FROM a\nRUN export PIP_NO_CACHE_DIR=1; pip install foo==1.0\n", nil},
		{"set in the RUN, for the one command", "FROM a\nRUN PIP_NO_CACHE_DIR=true pip install foo==1.0\n", nil},
		{"set to off in the RUN", "FROM a\nRUN PIP_NO_CACHE_DIR=off pip install foo==1.0\n", []int{2}},
		{"set to 0 in the RUN", "FROM a\nRUN PIP_NO_CACHE_DIR=0 pip install foo==1.0\n", []int{2}},
		{"set to nothing in the RUN", "FROM a\nRUN PIP_NO_CACHE_DIR= pip install foo==1.0\n", []int{2}},
		{
			"set twice in the RUN, off first",
			"FROM a\nRUN PIP_NO_CACHE_DIR=0 pip install a==1 && PIP_NO_CACHE_DIR=1 pip install b==1\n",
			[]int{2},
		},
		{
			"set twice in the RUN, on first",
			"FROM a\nRUN PIP_NO_CACHE_DIR=1 pip install a==1 && PIP_NO_CACHE_DIR=0 pip install b==1\n",
			nil,
		},

		{"given twice in one ENV, on last", "FROM a\nENV PIP_NO_CACHE_DIR=0 PIP_NO_CACHE_DIR=1\n" + install, nil},
		{"given twice in one ENV, off last", "FROM a\nENV PIP_NO_CACHE_DIR=1 PIP_NO_CACHE_DIR=0\n" + install, []int{3}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var lines []int
			for _, v := range testutils.LintDockerfile(tc.dockerfile, []rule.Rule{rules.DL3042()}) {
				lines = append(lines, v.Line)
			}

			if !slices.Equal(lines, tc.lines) {
				t.Errorf("reported at lines %v, want %v", lines, tc.lines)
			}
		})
	}
}
