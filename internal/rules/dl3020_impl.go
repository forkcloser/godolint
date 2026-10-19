package rules

import (
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

	// ADD is OK if all sources are archives or URLs
	for _, src := range add.Source {
		if !isArchive(src) && !isURL(src) {
			return false
		}
	}

	return true
}

// isArchive checks if path is an archive file.
// Ported from archiveFileFormatExtensions in Hadolint/Rule.hs.
func isArchive(path string) bool {
	path = dropQuotes(path)
	// Archive extensions from hadolint
	archiveExts := []string{
		tarSuffix, ".Z", ".bz2", ".gz", ".lz", ".lzma",
		".tZ", ".tb2", ".tbz", ".tbz2", tgzSuffix, ".tlz",
		".txz", ".tzo", ".t7z", ".tz", ".taz",
		".xz", ".zst", ".tzst",
	}

	for _, ext := range archiveExts {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return true
		}
	}

	return false
}

// isURL checks if path is a URL.
func isURL(path string) bool {
	path = dropQuotes(path)

	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}
