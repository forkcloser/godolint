package rules

import (
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3027 checks for use of `apt` instead of `apt-get` or `apt-cache`.
func DL3027() rule.Rule {
	return rule.NewSimpleRule(
		DL3027Meta.Code,
		DL3027Meta.Severity,
		DL3027Meta.Message,
		checkDL3027,
	)
}

func checkDL3027(instruction syntax.Instruction) bool {
	run, ok := instruction.(*syntax.Run)
	if !ok {
		return true
	}

	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	return !shell.UsingProgram("apt", parsed)
}
