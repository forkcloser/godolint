package rules

import (
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3003 checks for use of `cd` instead of WORKDIR.
func DL3003() rule.Rule {
	return rule.NewSimpleRule(
		DL3003Meta.Code,
		DL3003Meta.Severity,
		DL3003Meta.Message,
		checkDL3003,
	)
}

func checkDL3003(instruction syntax.Instruction) bool {
	run, ok := instruction.(*syntax.Run)
	if !ok {
		return true
	}

	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	return !shell.UsingProgram("cd", parsed)
}
