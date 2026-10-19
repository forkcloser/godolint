package rules

import (
	"strings"
	"unicode"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3041Rule checks that dnf and microdnf install pinned versions. Ported
// from Hadolint.Rule.DL3041: a version written through a variable is pinned
// when the variable is an ENV of the stage or an ARG, which the rule tracks.
type DL3041Rule struct{}

// DL3041 creates the rule for dnf version pinning.
func DL3041() rule.Rule {
	return &DL3041Rule{}
}

// Code returns the rule code.
func (*DL3041Rule) Code() rule.Code {
	return DL3041Meta.Code
}

// Severity returns the rule severity.
func (*DL3041Rule) Severity() rule.Severity {
	return DL3041Meta.Severity
}

// Message returns the rule message.
func (*DL3041Rule) Message() string {
	return DL3041Meta.Message
}

// InitialState returns the initial state for this rule.
func (*DL3041Rule) InitialState() rule.State {
	return rule.EmptyState(newVersionVars())
}

// Check records ENV, ARG and FROM, and fails a RUN whose packages or
// modules are not all pinned.
func (*DL3041Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	return checkPinnedVersions(DL3041Meta, line, state, instruction,
		func(vars versionVars, parsed *shell.ParsedShell) bool {
			return dnfPackagesPinned(vars, parsed) && dnfModulesPinned(parsed)
		})
}

// Finalize performs no final check.
func (*DL3041Rule) Finalize(state rule.State) rule.State {
	return state
}

func dnfPackagesPinned(vars versionVars, parsed *shell.ParsedShell) bool {
	for _, pkg := range getDnfPackages(parsed) {
		if !isDnfPackageVersionFixed(vars, pkg) {
			return false
		}
	}

	return true
}

func dnfModulesPinned(parsed *shell.ParsedShell) bool {
	for _, mod := range getDnfModules(parsed) {
		if !isDnfModuleVersionFixed(mod) {
			return false
		}
	}

	return true
}

// getDnfPackages is every package a dnf install names; module and group
// commands are not package installs.
func getDnfPackages(parsed *shell.ParsedShell) []string {
	var packages []string

	for _, cmd := range parsed.PresentCommands {
		if isDnfModuleCmd(cmd) || isDnfGroupCmd(cmd) {
			continue
		}

		packages = append(packages, dnfInstallFilter(cmd)...)
	}

	return packages
}

func getDnfModules(parsed *shell.ParsedShell) []string {
	var modules []string

	for _, cmd := range parsed.PresentCommands {
		if !isDnfModuleCmd(cmd) {
			continue
		}

		modules = append(modules, dnfInstallFilter(cmd)...)
	}

	return modules
}

func isDnfCmd(cmd shell.Command) bool {
	return cmd.Name == dnfCommand || cmd.Name == microdnfCommand
}

func isDnfModuleCmd(cmd shell.Command) bool {
	return isDnfCmd(cmd) && shell.CmdHasArgs(cmd.Name, []string{moduleArg}, cmd)
}

func isDnfGroupCmd(cmd shell.Command) bool {
	return isDnfCmd(cmd) && shell.CmdHasArgs(cmd.Name, []string{"group"}, cmd)
}

func dnfInstallFilter(cmd shell.Command) []string {
	if !isDnfCmd(cmd) || !shell.CmdHasArgs(cmd.Name, []string{installArg}, cmd) {
		return nil
	}

	var packages []string

	for _, arg := range shell.GetArgsNoFlags(shell.DropFlagArg(dnfFlagsWithValue(), cmd)) {
		if arg != installArg && arg != moduleArg {
			packages = append(packages, arg)
		}
	}

	return packages
}

// dnfFlagsWithValue are the dnf and microdnf options that take their value as
// a separate argument, which is not a package (--releasever 9).
func dnfFlagsWithValue() []string {
	return []string{
		"c", "config", "comment", "d", "debuglevel", "destdir", "disableplugin", "disablerepo",
		"downloaddir", "e", "enableplugin", "enablerepo", "errorlevel", "exclude", "forcearch",
		"installroot", "R", "randomwait", "releasever", "repo", "repofrompath", "repoid", "setopt", "x",
	}
}

// isDnfPackageVersionFixed is hadolint's packageVersionFixed: a name with no
// dash has no version; an .rpm file has one; a name with a variable is pinned
// when the variable is defined; otherwise what follows the first dash must
// look like a version.
func isDnfPackageVersionFixed(vars versionVars, pkg string) bool {
	parts := strings.Split(pkg, "-")
	if len(parts) <= 1 {
		return false
	}

	if strings.HasSuffix(pkg, ".rpm") {
		return true
	}

	if strings.Contains(pkg, "$") {
		return vars.defines(pkg)
	}

	return isDnfVersionLike(parts[1:])
}

func isDnfVersionLike(parts []string) bool {
	if len(parts) == 0 {
		return false
	}

	hasDigitStart := false

	for _, part := range parts {
		if !isDnfValidVersionPart(part) {
			return false
		}

		if unicode.IsDigit(rune(part[0])) {
			hasDigitStart = true
		}
	}

	return hasDigitStart
}

func isDnfValidVersionPart(part string) bool {
	if part == "" {
		return false
	}

	for _, ch := range part {
		if !isDnfVersionChar(ch) {
			return false
		}
	}

	return true
}

func isDnfVersionChar(char rune) bool {
	return unicode.IsDigit(char) ||
		unicode.IsUpper(char) ||
		unicode.IsLower(char) ||
		char == '.' ||
		char == '~' ||
		char == '^' ||
		char == '_' ||
		char == ':' ||
		char == '+' ||
		char == '*'
}

func isDnfModuleVersionFixed(mod string) bool {
	return strings.Contains(mod, ":")
}
