package rules

import (
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3000 creates a rule for checking WORKDIR paths are absolute.
func DL3000() rule.Rule {
	return rule.NewSimpleRule(
		DL3000Meta.Code,
		DL3000Meta.Severity,
		DL3000Meta.Message,
		checkDL3000,
	)
}

func checkDL3000(instruction syntax.Instruction) bool {
	workdir, ok := instruction.(*syntax.Workdir)
	if !ok {
		return true
	}

	path := dropQuotes(workdir.Directory)

	if strings.HasPrefix(path, "$") {
		return true
	}

	if strings.HasPrefix(path, "/") {
		return true
	}

	// Windows absolute path (drive letter + colon)
	if isWindowsAbsolute(path) {
		return true
	}

	return false
}
