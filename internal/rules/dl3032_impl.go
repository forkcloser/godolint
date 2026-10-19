package rules

import (
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3032 checks for yum clean all after yum install.
func DL3032() rule.Rule {
	return rule.NewSimpleRule(
		DL3032Meta.Code,
		DL3032Meta.Severity,
		DL3032Meta.Message,
		checkDL3032,
	)
}

// checkDL3032 passes a RUN with no yum install, or one whose first yum
// install comes before its first clean: a clean before the install leaves
// the cache the install filled.
func checkDL3032(instruction syntax.Instruction) bool {
	run, ok := instruction.(*syntax.Run)
	if !ok {
		return true
	}

	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	if firstCommand(parsed.PresentCommands, isYumInstall) < 0 {
		return true
	}

	return cleanFollowsInstall(parsed.PresentCommands, isYumInstall, isYumClean)
}

func isYumClean(cmd shell.Command) bool {
	// yum clean all
	if shell.CmdHasAllArgs(yumCommand, []string{cleanArg, allArg}, cmd) {
		return true
	}

	// rm -rf /var/cache/yum/*
	return shell.CmdHasAllArgs("rm", []string{recursiveForceFlag, "/var/cache/yum/*"}, cmd)
}
