package rules

import (
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3035 checks for zypper dist-upgrade usage.
func DL3035() rule.Rule {
	return rule.NewSimpleRule(
		DL3035Meta.Code,
		DL3035Meta.Severity,
		DL3035Meta.Message,
		checkDL3035,
	)
}

func checkDL3035(instruction syntax.Instruction) bool {
	run, ok := instruction.(*syntax.Run)
	if !ok {
		return true
	}

	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	for _, cmd := range parsed.PresentCommands {
		if shell.CmdHasArgs(zypperCommand, []string{"dist-upgrade"}, cmd) ||
			shell.CmdHasArgs(zypperCommand, []string{"dup"}, cmd) {
			return false
		}
	}

	return true
}
