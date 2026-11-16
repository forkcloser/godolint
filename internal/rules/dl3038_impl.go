package rules

import (
	"slices"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3038 checks for dnf/microdnf install without -y flag.
func DL3038() rule.Rule {
	return rule.NewSimpleRule(
		DL3038Meta.Code,
		DL3038Meta.Severity,
		DL3038Meta.Message,
		checkDL3038,
	)
}

func checkDL3038(instruction syntax.Instruction) bool {
	run, ok := instruction.(*syntax.Run)
	if !ok {
		return true
	}

	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	return !slices.ContainsFunc(parsed.PresentCommands, forgotDnfYesOption)
}

func forgotDnfYesOption(cmd shell.Command) bool {
	if !isDnfInstall(cmd) {
		return false
	}

	return !hasDnfYesOption(cmd)
}

func isDnfInstall(cmd shell.Command) bool {
	if cmd.Name != dnfCommand && cmd.Name != microdnfCommand {
		return false
	}

	return shell.CmdHasArgs(cmd.Name, []string{installArg}, cmd) ||
		shell.CmdHasArgs(cmd.Name, []string{"groupinstall"}, cmd) ||
		shell.CmdHasArgs(cmd.Name, []string{"localinstall"}, cmd)
}

func hasDnfYesOption(cmd shell.Command) bool {
	return shell.HasAnyFlag([]string{"y", "assumeyes"}, cmd)
}
