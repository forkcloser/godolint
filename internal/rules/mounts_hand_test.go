package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestCacheMountHand holds how the six rules that excuse a missing cleanup
// read a RUN's mounts, which hadolint's cases leave open past `target=` and
// the exact path: a cache or tmpfs mount whose target holds the path
// anywhere in it, written target=, dst= or destination=.
func TestCacheMountHand(t *testing.T) {
	t.Parallel()

	for _, leaves := range []struct {
		name string
		rule func() rule.Rule
		// run is a command the rule reports when nothing excuses it.
		run string
		// path is where the cache is, as an image running as root has it.
		path string
	}{
		{"DL3009", rules.DL3009, "apt-get update", "/var/lib/apt/lists"},
		{"DL3019", rules.DL3019, "apk add foo", "/var/cache/apk"},
		{"DL3036", rules.DL3036, "zypper install -y foo", "/var/cache/zypp"},
		{"DL3040, the system cache", rules.DL3040, "dnf install -y foo", "/var/cache/libdnf5"},
		{"DL3040, the user cache", rules.DL3040, "dnf install -y foo", "/root/.cache/libdnf5"},
		{"DL3042", rules.DL3042, "pip install foo", "/root/.cache/pip"},
		{"DL3060", rules.DL3060, "yarn install", "/root/.cache/yarn"},
	} {
		path := leaves.path

		for _, tc := range []struct {
			name   string
			flags  string
			passes bool
		}{
			{"no mount", "", false},
			{"a cache mount", "--mount=type=cache,target=" + path, true},
			{"a tmpfs mount", "--mount=type=tmpfs,target=" + path, true},
			{"the target written dst", "--mount=type=cache,dst=" + path, true},
			{"the target written destination", "--mount=type=cache,destination=" + path, true},
			{"the target before the type", "--mount=target=" + path + ",type=cache", true},
			{"the target among other options", "--mount=type=cache,id=one,target=" + path + ",sharing=locked", true},
			{"the target in double quotes", `--mount=type=cache,target="` + path + `"`, true},
			{"a directory under the path", "--mount=type=cache,target=" + path + "/sub", true},
			{"the path under another directory", "--mount=type=cache,target=/mnt" + path, true},
			{"the path and more in the same name", "--mount=type=cache,target=" + path + "-old", true},
			{"the second of two mounts", "--mount=type=cache,target=/elsewhere --mount=type=cache,target=" + path, true},
			{"a cache mount elsewhere", "--mount=type=cache,target=/elsewhere", false},
			{"the path as the source", "--mount=type=cache,target=/elsewhere,source=" + path, false},
			{"a bind mount", "--mount=type=bind,target=" + path, false},
			{"a mount of no type, which is a bind", "--mount=target=" + path, false},
			{"a secret", "--mount=type=secret,target=" + path, false},
		} {
			t.Run(leaves.name+"/"+tc.name, func(t *testing.T) {
				t.Parallel()

				dockerfile := "FROM image:1\nRUN " + tc.flags + " " + leaves.run
				violations := testutils.LintDockerfile(dockerfile, []rule.Rule{leaves.rule()})

				if passed := len(violations) == 0; passed != tc.passes {
					t.Errorf("%q: passed %v, want %v: %v", dockerfile, passed, tc.passes, violations)
				}
			})
		}
	}
}

// TestCacheMountHandApt holds DL3009's second way out: apt's two
// directories mounted, both of them.
func TestCacheMountHandApt(t *testing.T) {
	t.Parallel()

	const (
		lib   = "--mount=type=cache,dst=/var/lib/apt"
		cache = "--mount=type=tmpfs,destination=/var/cache/apt"
	)

	for _, tc := range []struct {
		name   string
		flags  string
		passes bool
	}{
		{"both directories", lib + " " + cache, true},
		{"the state directory alone", lib, false},
		{"the cache directory alone", cache, false},
		{"both under another root", lib + "-one " + "--mount=type=cache,target=/mnt/var/cache/apt/archives", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dockerfile := "FROM image:1\nRUN " + tc.flags + " apt-get update"
			violations := testutils.LintDockerfile(dockerfile, []rule.Rule{rules.DL3009()})

			if passed := len(violations) == 0; passed != tc.passes {
				t.Errorf("%q: passed %v, want %v: %v", dockerfile, passed, tc.passes, violations)
			}
		})
	}
}
