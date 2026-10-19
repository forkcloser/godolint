package rules

import (
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3040 checks for dnf clean all after a dnf install.
func DL3040() rule.Rule {
	return rule.NewSimpleRule(
		DL3040Meta.Code,
		DL3040Meta.Severity,
		DL3040Meta.Message,
		checkDL3040,
	)
}

// checkDL3040 passes a RUN whose dnf cache is a mount, and otherwise one
// where, for dnf and for microdnf alike, there is no install or the first
// install comes before the first clean.
func checkDL3040(instruction syntax.Instruction) bool {
	run, ok := instruction.(*syntax.Run)
	if !ok {
		return true
	}

	if hasCacheOrTmpfsMount(run.Flags, "/var/cache/libdnf5") ||
		hasCacheOrTmpfsMount(run.Flags, ".cache/libdnf5") {
		return true
	}

	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	for _, name := range []string{dnfCommand, microdnfCommand} {
		install := func(cmd shell.Command) bool { return isDnfInstallCmd(name, cmd) }
		clean := func(cmd shell.Command) bool { return isDnfCleanCmd(name, cmd) }

		if firstCommand(parsed.PresentCommands, install) < 0 {
			continue
		}

		if !cleanFollowsInstall(parsed.PresentCommands, install, clean) {
			return false
		}
	}

	return true
}

// isDnfInstallCmd is a dnf or microdnf command that installs or upgrades,
// in any of the spellings dnf accepts.
func isDnfInstallCmd(name string, cmd shell.Command) bool {
	return shell.CmdHasArgs(name, []string{
		installArg, "in", "upgrade", "up", "upgrade-minimal", "up-min", "reinstall", "rei",
	}, cmd)
}

// isDnfCleanCmd is `<name> clean all`, or an rm of dnf's cache.
func isDnfCleanCmd(name string, cmd shell.Command) bool {
	if shell.CmdHasAllArgs(name, []string{cleanArg, allArg}, cmd) {
		return true
	}

	return shell.CmdHasAllArgs("rm", []string{recursiveForceFlag, "/var/cache/libdnf5"}, cmd)
}
