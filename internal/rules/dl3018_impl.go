package rules

import (
	"slices"
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3018 checks for apk add without version pinning.
func DL3018() rule.Rule {
	return rule.NewSimpleRule(
		DL3018Meta.Code,
		DL3018Meta.Severity,
		DL3018Meta.Message,
		checkDL3018,
	)
}

func checkDL3018(instruction syntax.Instruction) bool {
	run, ok := instruction.(*syntax.Run)
	if !ok {
		return true
	}

	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	for _, cmd := range parsed.PresentCommands {
		packages := apkAddPackages(cmd)
		for _, pkg := range packages {
			if !isApkVersionFixed(pkg) {
				return false
			}
		}
	}

	return true
}

// apkAddPackages is the packages of an apk add: its arguments that are no
// flag, without `add` and without the values of the flags that take one
// after them (-t, --virtual, --repository, -X). A flag that carries its
// value after `=` takes nothing after it: in `--repository=URL hugo`, hugo is
// a package.
func apkAddPackages(cmd shell.Command) []string {
	if !shell.CmdHasArgs("apk", []string{addArg}, cmd) {
		return nil
	}

	valueFlags := []string{"t", "virtual", "repository", "X"}

	return slices.DeleteFunc(
		shell.GetArgsNoFlags(shell.DropFlagArg(valueFlags, cmd)),
		func(arg string) bool { return arg == addArg },
	)
}

func isApkVersionFixed(pkg string) bool {
	// Package is version-fixed if it has:
	// 1. a version constraint: = pins it, ~ > < bound it, as apk reads them
	// 2. .apk suffix (package file)
	return strings.ContainsAny(pkg, "=~><") || strings.HasSuffix(pkg, ".apk")
}
