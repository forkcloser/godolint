// Package rules implements all Dockerfile linting rules ported from hadolint.
//
// A comment that says "Ported from" names the hadolint function or rule the
// code is compared with, call for call: it is the map for that comparison,
// and is kept true.
package rules

import (
	"github.com/forkcloser/godolint/internal/pragma"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL1001 checks for inline ignore pragmas.
func DL1001() rule.Rule {
	return rule.NewSimpleRule(
		DL1001Meta.Code,
		DL1001Meta.Severity,
		DL1001Meta.Message,
		checkDL1001,
	)
}

// checkDL1001 fails a line ignore pragma, and only that one: a stage or
// global pragma, a shell pragma, or an ignore list hadolint would not read
// is a plain comment here, as upstream.
func checkDL1001(instruction syntax.Instruction) bool {
	comment, ok := instruction.(*syntax.Comment)
	if !ok {
		return true // Not a comment, pass
	}

	_, isPragma := pragma.ParseIgnore(comment.Text)

	return !isPragma
}
