package rules

import (
	"slices"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3019 checks for apk add without --no-cache flag.
func DL3019() rule.Rule {
	return rule.NewSimpleRule(
		DL3019Meta.Code,
		DL3019Meta.Severity,
		DL3019Meta.Message,
		checkDL3019,
	)
}

func checkDL3019(instruction syntax.Instruction) bool {
	run, ok := instruction.(*syntax.Run)
	if !ok {
		return true
	}

	if hasCacheOrTmpfsMount(run.Flags, "/var/cache/apk") {
		return true
	}

	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	return !slices.ContainsFunc(parsed.PresentCommands, forgotApkNoCacheOption)
}

func forgotApkNoCacheOption(cmd shell.Command) bool {
	if !shell.CmdHasArgs("apk", []string{addArg}, cmd) {
		return false
	}

	return !shell.HasFlag("no-cache", cmd)
}
