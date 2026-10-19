package rules

import (
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3060 checks for yarn cache clean after yarn install.
func DL3060() rule.Rule {
	return rule.NewSimpleRule(
		DL3060Meta.Code,
		DL3060Meta.Severity,
		DL3060Meta.Message,
		checkDL3060,
	)
}

// checkDL3060 passes a RUN with no yarn install, one whose yarn cache is a
// mount, or one whose first yarn install comes before its first cache clean.
func checkDL3060(instruction syntax.Instruction) bool {
	run, ok := instruction.(*syntax.Run)
	if !ok {
		return true
	}

	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	if firstCommand(parsed.PresentCommands, isYarnInstall) < 0 {
		return true
	}

	if hasCacheOrTmpfsMount(run.Flags, ".cache/yarn") ||
		hasCacheOrTmpfsMount(run.Flags, "/root/.cache/yarn") {
		return true
	}

	return cleanFollowsInstall(parsed.PresentCommands, isYarnInstall, isYarnCacheClean)
}

func isYarnInstall(cmd shell.Command) bool {
	return shell.CmdHasArgs("yarn", []string{installArg}, cmd)
}

func isYarnCacheClean(cmd shell.Command) bool {
	return shell.CmdHasAllArgs("yarn", []string{"cache", cleanArg}, cmd)
}
