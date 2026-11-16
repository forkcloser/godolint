package rules

import (
	"slices"
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3020 creates a rule for checking ADD vs COPY usage.
func DL3020() rule.Rule {
	return rule.NewSimpleRule(
		DL3020Meta.Code,
		DL3020Meta.Severity,
		DL3020Meta.Message,
		checkDL3020,
	)
}

func checkDL3020(instruction syntax.Instruction) bool {
	add, ok := instruction.(*syntax.Add)
	if !ok {
		return true
	}

	for _, src := range add.Source {
		if !isArchive(src) && !isURL(src) {
			return false
		}
	}

	return true
}

// archiveSuffixes is the suffixes of what a build's ADD unpacks: a tar, as it
// is or packed with one of the four compressions the build tells by content,
// gzip, bzip2, xz and zstd. A file packed with any other (compress, lzip,
// lzma, lzop, 7z) is added as it is, and is no archive to ADD.
//
//nolint:gochecknoglobals // a constant list
var archiveSuffixes = []string{
	tarSuffix,
	".gz", tgzSuffix, ".taz",
	".bz2", ".tz2", ".tbz", ".tbz2", ".tb2",
	".xz", ".txz",
	".zst", ".tzst",
}

// isArchive reports whether path is named as an archive ADD unpacks. The
// case of the name does not matter: the build looks at the content.
func isArchive(path string) bool {
	path = strings.ToLower(dropQuotes(path))

	return slices.ContainsFunc(archiveSuffixes, func(suffix string) bool { return strings.HasSuffix(path, suffix) })
}

// isURL checks if path is a URL.
func isURL(path string) bool {
	path = dropQuotes(path)

	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}
