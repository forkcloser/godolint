package rules

import "strings"

// stageNamed is the name a build looks a stage up by for the source of a
// COPY --from or of a mount: the lower case of what is written. Not for the
// image of a FROM, which a build looks up as written.
func stageNamed(from string) string {
	return strings.ToLower(from)
}
