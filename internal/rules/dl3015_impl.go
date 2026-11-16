package rules

import (
	"slices"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3015 checks for apt-get install without --no-install-recommends.
func DL3015() rule.Rule {
	return rule.NewSimpleRule(
		DL3015Meta.Code,
		DL3015Meta.Severity,
		DL3015Meta.Message,
		checkDL3015,
	)
}

func checkDL3015(instruction syntax.Instruction) bool {
	run, ok := instruction.(*syntax.Run)
	if !ok {
		return true
	}

	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	return !slices.ContainsFunc(parsed.PresentCommands, forgotNoInstallRecommends)
}

func forgotNoInstallRecommends(cmd shell.Command) bool {
	if !shell.CmdHasArgs("apt-get", []string{installArg}, cmd) {
		return false
	}

	return !disablesRecommendOption(cmd)
}

func disablesRecommendOption(cmd shell.Command) bool {
	if shell.HasFlag("no-install-recommends", cmd) {
		return true
	}

	if shell.HasArg("APT::Install-Recommends=false", cmd) {
		return true
	}

	return false
}
