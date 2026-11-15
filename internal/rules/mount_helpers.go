package rules

import "strings"

// hasCacheOrTmpfsMount reports whether one of a RUN's --mount flags is a
// cache or tmpfs mount whose target holds fragment anywhere in it. hadolint
// asks for the fragment, not for the path: a cache in another user's home
// (/home/node/.cache/yarn for .cache/yarn) is the cache all the same. The
// target is written target=, dst= or destination=, as BuildKit takes it.
func hasCacheOrTmpfsMount(flags []string, fragment string) bool {
	for _, flag := range flags {
		spec, isMount := strings.CutPrefix(flag, "--mount=")
		if !isMount {
			continue
		}

		mountType, target := "", ""

		for part := range strings.SplitSeq(spec, ",") {
			key, value, hasValue := strings.Cut(part, "=")
			if !hasValue {
				continue
			}

			switch strings.TrimSpace(key) {
			case "type":
				mountType = strings.TrimSpace(value)
			case "target", "dst", "destination":
				target = strings.TrimSpace(value)
			default:
				// Other mount options are irrelevant here.
			}
		}

		if (mountType == "cache" || mountType == "tmpfs") && strings.Contains(target, fragment) {
			return true
		}
	}

	return false
}
