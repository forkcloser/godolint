package rules

import (
	"slices"
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3028 checks for gem install without version pinning.
func DL3028() rule.Rule {
	return rule.NewSimpleRule(
		DL3028Meta.Code,
		DL3028Meta.Severity,
		DL3028Meta.Message,
		checkDL3028,
	)
}

func checkDL3028(instruction syntax.Instruction) bool {
	run, ok := instruction.(*syntax.Run)
	if !ok {
		return true
	}

	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	for _, cmd := range parsed.PresentCommands {
		gems := getGemPackages(cmd)
		for _, gem := range gems {
			if !isGemVersionFixed(gem) {
				return false
			}
		}
	}

	return true
}

func getGemPackages(cmd shell.Command) []string {
	if !shell.CmdHasArgs("gem", []string{installArg}, cmd) && !shell.CmdHasArgs("gem", []string{"i"}, cmd) {
		return nil
	}

	// A version flag pins every package the command installs, which is what
	// the rule asks for; nothing to report.
	args := shell.GetArgs(cmd)
	if hasGemVersionFlag(args) {
		return nil
	}

	// Everything after "--" belongs to the built extension, not to gem.
	if end := slices.Index(args, "--"); end >= 0 {
		args = args[:end]
	}

	return gemOperands(args)
}

// hasGemVersionFlag reports whether the command carries gem's version flag, in
// either of its three spellings.
func hasGemVersionFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-v" || arg == "--version" || strings.HasPrefix(arg, "--version=") {
			return true
		}
	}

	return false
}

// gemOperands drops the subcommand and the flags, leaving the package names.
// A flag without "=" takes the next argument as its value.
func gemOperands(args []string) []string {
	var packages []string

	skipNext := false

	for _, arg := range args {
		switch {
		case skipNext:
			skipNext = false
		case arg == installArg || arg == "i":
		case strings.HasPrefix(arg, "-"):
			skipNext = !strings.Contains(arg, "=")
		default:
			packages = append(packages, arg)
		}
	}

	return packages
}

func isGemVersionFixed(pkg string) bool {
	return strings.Contains(pkg, ":")
}
