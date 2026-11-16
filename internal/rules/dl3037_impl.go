package rules

import (
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3037 checks for zypper install without version pinning.
func DL3037() rule.Rule {
	return rule.NewSimpleRule(
		DL3037Meta.Code,
		DL3037Meta.Severity,
		DL3037Meta.Message,
		checkDL3037,
	)
}

func checkDL3037(instruction syntax.Instruction) bool {
	run, ok := instruction.(*syntax.Run)
	if !ok {
		return true
	}

	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	packages := getZypperPackages(parsed)
	for _, pkg := range packages {
		if !isZypperVersionFixed(pkg) {
			return false
		}
	}

	return true
}

func getZypperPackages(parsed *shell.ParsedShell) []string {
	var packages []string

	for _, cmd := range parsed.PresentCommands {
		if shell.CmdHasArgs(zypperCommand, []string{installArg}, cmd) ||
			shell.CmdHasArgs(zypperCommand, []string{"in"}, cmd) {
			args := shell.GetArgsNoFlags(cmd)
			for _, arg := range args {
				if arg != installArg && arg != "in" {
					packages = append(packages, arg)
				}
			}
		}
	}

	return packages
}

func isZypperVersionFixed(pkg string) bool {
	if strings.Contains(pkg, "=") ||
		strings.Contains(pkg, ">=") ||
		strings.Contains(pkg, ">") ||
		strings.Contains(pkg, "<=") ||
		strings.Contains(pkg, "<") {
		return true
	}

	// .rpm files are considered version-fixed
	if strings.HasSuffix(pkg, ".rpm") {
		return true
	}

	return false
}
