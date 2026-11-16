package rules

import (
	"slices"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3047 checks for wget without progress bar option.
func DL3047() rule.Rule {
	return rule.NewSimpleRule(
		DL3047Meta.Code,
		DL3047Meta.Severity,
		DL3047Meta.Message,
		checkDL3047,
	)
}

func checkDL3047(instruction syntax.Instruction) bool {
	run, ok := instruction.(*syntax.Run)
	if !ok {
		return true
	}

	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	return !slices.ContainsFunc(parsed.PresentCommands, forgotWgetProgress)
}

func forgotWgetProgress(cmd shell.Command) bool {
	if cmd.Name != "wget" {
		return false
	}

	if shell.HasFlag("progress", cmd) {
		return false
	}

	if hasWgetSpecialFlags(cmd) {
		return false
	}

	return true
}

func hasWgetSpecialFlags(cmd shell.Command) bool {
	if shell.HasAnyFlag([]string{"q", "quiet"}, cmd) {
		return true
	}

	if shell.HasAnyFlag([]string{"o", "output-file"}, cmd) {
		return true
	}

	if shell.HasAnyFlag([]string{"a", "append-output"}, cmd) {
		return true
	}

	if shell.HasFlag("no-verbose", cmd) {
		return true
	}

	// Special case: -nv (combined flag)
	if shell.CmdHasArgs("wget", []string{"-nv"}, cmd) {
		return true
	}

	return false
}
