package rules_test

import (
	"slices"
	"testing"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// dl3026Lines is the lines DL3026 reports on dockerfile with the registries
// allowed, in the order the rule reports them.
func dl3026Lines(dockerfile string, registries ...string) []int {
	cfg := config.Default()
	cfg.AllowedRegistries = registries

	var lines []int

	for _, v := range testutils.LintDockerfile(dockerfile, []rule.Rule{rules.DL3026WithConfig(cfg)}) {
		if v.Code == "DL3026" {
			lines = append(lines, v.Line)
		}
	}

	return lines
}

// TestDL3026Spec holds hadolint's cases with the lines the failures fall on,
// which the generated tests do not look at.
func TestDL3026Spec(t *testing.T) {
	t.Parallel()

	const dockerIO = "FROM ubuntu:18.04 AS builder1\n" +
		"FROM zemanlx/ubuntu:18.04 AS builder2\n" +
		"FROM docker.io/zemanlx/ubuntu:18.04 AS builder3"

	for _, tc := range []struct {
		name       string
		dockerfile string
		registries []string
		lines      []int
	}{
		{"does not warn on empty allowed registries", "FROM random.com/debian", nil, nil},
		{"warn on non-allowed registry", "FROM random.com/debian", []string{"docker.io"}, []int{1}},
		{"does not warn on allowed registries", "FROM random.com/debian", []string{"x.com", "random.com"}, nil},
		{"doesn't warn on scratch image", "FROM scratch", []string{"x.com", "random.com"}, nil},
		{"allows all forms of docker.io", dockerIO, []string{"docker.io"}, nil},
		{
			"allows using previous stages",
			"FROM random.com/foo AS builder1\nFROM builder1 AS builder2",
			[]string{"random.com"},
			nil,
		},
		{"warn on non-allowed wildcard registry", "FROM x.com/debian", []string{"*.random.com"}, []int{1}},
		{
			"does not warn on allowed wildcard registries",
			"FROM foo.random.com/debian",
			[]string{"x.com", "*.random.com"},
			nil,
		},
		{"does not warn on * registry", dockerIO, []string{"*"}, nil},
		{
			"does warn on copy from untrusted registry",
			"COPY --from=untrusted.com/repo/image:tag /foo /bar",
			[]string{"trusted.com"},
			[]int{1},
		},
		{
			"does not warn on copy from trusted registry",
			"COPY --from=trusted.com/repo/image:tag /foo /bar",
			[]string{"trusted.com"},
			nil,
		},
		{
			"warn on run with bind mount from untrusted registry",
			"RUN --mount=type=bind,from=untrusted.com/repo/image:tag,target=/foo foobar",
			[]string{"trusted.com"},
			[]int{1},
		},
		{
			"warn on run with cache mount from untrusted registry",
			"RUN --mount=type=cache,from=untrusted.com/repo/image:tag,target=/foo foobar",
			[]string{"trusted.com"},
			[]int{1},
		},
		{
			"don't warn on run with bind mount from trusted registry",
			"RUN --mount=type=bind,from=trusted.com/repo/image:tag,target=/foo foobar",
			[]string{"trusted.com"},
			nil,
		},
		{
			"don't warn on run with cache mount from trusted registry",
			"RUN --mount=type=cache,from=trusted.com/repo/image:tag,target=/foo foobar",
			[]string{"trusted.com"},
			nil,
		},
		{
			"distrust all forms of docker.io if trusted registries are given",
			dockerIO,
			[]string{"trusted.com"},
			[]int{1, 2, 3},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := dl3026Lines(tc.dockerfile, tc.registries...); !slices.Equal(got, tc.lines) {
				t.Errorf("reported at lines %v, want %v", got, tc.lines)
			}
		})
	}
}

// TestDL3026Hand holds what hadolint's cases leave open: the other name of
// Docker Hub, the shapes of a wildcard, what makes the start of a name a
// registry (a dot, a port, localhost), a stage named by its alias or its
// number in COPY and in a mount, the alias of a stage whose image is
// refused, the mounts that take nothing from an image, and that one RUN is
// reported once.
func TestDL3026Hand(t *testing.T) {
	t.Parallel()

	const trusted = "trusted.com"

	for _, tc := range []struct {
		name       string
		dockerfile string
		registries []string
		lines      []int
	}{
		{"an image of no registry, hub.docker.com allowed", "FROM debian:12", []string{"hub.docker.com"}, nil},
		{"an image of no registry, another allowed", "FROM debian:12", []string{trusted}, []int{1}},
		{"a wildcard at the start with no dot", "FROM my-trusted.com/a", []string{"*trusted.com"}, nil},
		{"a wildcard at the start, another ending", "FROM trusted.com.evil.io/a", []string{"*trusted.com"}, []int{1}},
		{"a wildcard at the end", "FROM trusted.com/a\nFROM trusted.io/b", []string{"trusted.*"}, nil},
		{"a wildcard at the end, another start", "FROM untrusted.com/a", []string{"trusted.*"}, []int{1}},
		{"a wildcard in the middle is no wildcard", "FROM a.b.com/x", []string{"a.*.com"}, []int{1}},
		{"an exact name is not a prefix", "FROM trusted.com.evil.io/a", []string{trusted}, []int{1}},
		{"an exact name is not a suffix", "FROM untrusted.com/a", []string{"trusted.com"}, []int{1}},
		{"localhost is a registry", "FROM localhost/tools", []string{"docker.io"}, []int{1}},
		{"a host with a port is a registry", "FROM scratch\nCOPY --from=reg:5000/img /a /b", []string{"docker.io"}, []int{2}},
		{"FROM a host with a port", "FROM reg:5000/img:1", []string{"docker.io"}, []int{1}},
		{"FROM a host with a port, allowed", "FROM reg:5000/img:1", []string{"reg:5000"}, nil},
		{"FROM a host with a dot and a port, allowed without the port", "FROM myreg.io:5000/img", []string{"myreg.io"}, []int{1}},
		{"FROM a host with a dot and a port, allowed by a wildcard", "FROM myreg.io:5000/img", []string{"myreg.io*"}, nil},
		{
			"an earlier stage by its alias and a digest",
			"FROM trusted.com/a AS base\nFROM base@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			[]string{"trusted.com"},
			nil,
		},
		{"a host with a port, allowed", "FROM scratch\nCOPY --from=reg:5000/img /a /b", []string{"reg:5000"}, nil},
		{
			"a mount from a host with a port",
			"FROM scratch\nRUN --mount=type=bind,from=reg:5000/img,target=/x true",
			[]string{"docker.io"},
			[]int{2},
		},
		{"localhost allowed", "FROM localhost/tools", []string{"localhost"}, nil},
		{"a namespace is not a registry", "FROM library/debian", []string{"docker.io"}, nil},
		{"a name with a dot and no slash is an image", "FROM my.image", []string{"docker.io"}, nil},
		{
			"an image named as its own alias is not an earlier stage",
			"FROM untrusted.com/base AS untrusted.com/base",
			[]string{trusted},
			[]int{1},
		},
		{
			"a stage whose image is refused still gives its alias",
			"FROM untrusted.com/a AS base\nFROM base",
			[]string{trusted},
			[]int{1},
		},
		{
			"an image behind a registry is not the stage of that name",
			"FROM trusted.com/a AS base\nFROM untrusted.com/base",
			[]string{trusted},
			[]int{2},
		},
		{
			"copy from an earlier stage by its alias",
			"FROM trusted.com/a AS build\nFROM scratch\nCOPY --from=build /a /b",
			[]string{trusted},
			nil,
		},
		{
			"copy from an earlier stage by its number",
			"FROM trusted.com/a\nFROM scratch\nCOPY --from=0 /a /b",
			[]string{trusted},
			nil,
		},
		{"copy from an image of no registry", "FROM scratch\nCOPY --from=debian:12 /a /b", []string{trusted}, []int{2}},
		{"copy with no from", "FROM scratch\nCOPY a b", []string{trusted}, nil},
		{"copy from nothing names no stage", "FROM scratch\nCOPY --from= a b", []string{trusted}, []int{2}},
		{
			"a mount from an earlier stage by its alias",
			"FROM trusted.com/a AS build\nFROM scratch\nRUN --mount=type=bind,from=build,target=/x true",
			[]string{trusted},
			nil,
		},
		{
			"a mount from an earlier stage by its number",
			"FROM trusted.com/a\nFROM scratch\nRUN --mount=type=cache,from=0,target=/x true",
			[]string{trusted},
			nil,
		},
		{
			"a mount with no type is a bind mount",
			"FROM scratch\nRUN --mount=from=untrusted.com/a,target=/x true",
			[]string{trusted},
			[]int{2},
		},
		{"a mount with no from", "FROM scratch\nRUN --mount=type=cache,target=/x true", []string{trusted}, nil},
		{
			"a secret mount takes nothing from an image",
			"FROM scratch\nRUN --mount=type=secret,id=from=untrusted.com/a true",
			[]string{trusted},
			nil,
		},
		{
			"a tmpfs mount takes nothing from an image",
			"FROM scratch\nRUN --mount=type=tmpfs,from=untrusted.com/a,target=/x true",
			[]string{trusted},
			nil,
		},
		{
			"the second of two mounts",
			"FROM scratch\nRUN --mount=type=bind,from=trusted.com/a,target=/x --mount=type=cache,from=untrusted.com/b,target=/y true",
			[]string{trusted},
			[]int{2},
		},
		{
			"two mounts from untrusted images are one failure",
			"FROM scratch\nRUN --mount=type=bind,from=untrusted.com/a,target=/x --mount=type=cache,from=untrusted.com/b,target=/y true",
			[]string{trusted},
			[]int{2},
		},
		{"a run with no mount", "FROM scratch\nRUN --network=none true", []string{trusted}, nil},
		{"another instruction", "FROM scratch\nADD untrusted.com/a /b", []string{trusted}, nil},
		{"nothing allowed means everything is", "FROM untrusted.com/a\nCOPY --from=other.io/b /a /b", nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := dl3026Lines(tc.dockerfile, tc.registries...); !slices.Equal(got, tc.lines) {
				t.Errorf("reported at lines %v, want %v", got, tc.lines)
			}
		})
	}
}
