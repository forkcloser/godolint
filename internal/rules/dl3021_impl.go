package rules

import (
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3021 creates a rule for checking COPY with multiple sources ends with /.
func DL3021() rule.Rule {
	return rule.NewSimpleRule(
		DL3021Meta.Code,
		DL3021Meta.Severity,
		DL3021Meta.Message,
		checkDL3021,
	)
}

func checkDL3021(instruction syntax.Instruction) bool {
	copyInstr, ok := instruction.(*syntax.Copy)
	if !ok {
		return true
	}

	if len(copyInstr.Source) <= 1 {
		return true
	}

	dest := dropQuotes(copyInstr.Destination)
	if dest == "" {
		return false
	}

	return dest[len(dest)-1] == '/'
}
