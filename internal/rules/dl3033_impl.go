package rules

import (
	"strings"
	"unicode"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3033Rule checks that yum installs pinned versions. Ported from
// Hadolint.Rule.DL3033: a version written through a variable is pinned when
// the variable is an ENV of the stage or an ARG, which the rule tracks.
type DL3033Rule struct {
	rule.StatefulRuleBase
}

// DL3033 creates the rule for yum version pinning.
func DL3033() rule.Rule {
	return &DL3033Rule{StatefulRuleBase: rule.NewStatefulRuleBase(DL3033Meta)}
}

// InitialState returns the initial state for this rule.
func (*DL3033Rule) InitialState() rule.State {
	return rule.EmptyState(newVersionVars())
}

// Check records ENV, ARG and FROM, and fails a RUN whose packages or
// modules are not all pinned.
func (r *DL3033Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	return checkPinnedVersions(r, line, state, instruction,
		func(vars versionVars, parsed *shell.ParsedShell) bool {
			return yumPackagesPinned(vars, parsed) && yumModulesPinned(parsed)
		})
}

func yumPackagesPinned(vars versionVars, parsed *shell.ParsedShell) bool {
	for _, pkg := range getYumPackages(parsed) {
		if !isYumPackageVersionFixed(vars, pkg) {
			return false
		}
	}

	return true
}

func yumModulesPinned(parsed *shell.ParsedShell) bool {
	for _, mod := range getYumModules(parsed) {
		if !isYumModuleVersionFixed(mod) {
			return false
		}
	}

	return true
}

func getYumPackages(parsed *shell.ParsedShell) []string {
	var packages []string

	for _, cmd := range parsed.PresentCommands {
		if shell.CmdHasArgs(yumCommand, []string{moduleArg}, cmd) {
			continue
		}

		packages = append(packages, yumInstallFilter(cmd)...)
	}

	return packages
}

func getYumModules(parsed *shell.ParsedShell) []string {
	var modules []string

	for _, cmd := range parsed.PresentCommands {
		if !shell.CmdHasArgs(yumCommand, []string{moduleArg}, cmd) {
			continue
		}

		modules = append(modules, yumInstallFilter(cmd)...)
	}

	return modules
}

func yumInstallFilter(cmd shell.Command) []string {
	if !shell.CmdHasArgs(yumCommand, []string{installArg}, cmd) {
		return nil
	}

	args := shell.GetArgsNoFlags(cmd)

	var packages []string

	for _, arg := range args {
		if arg != installArg && arg != moduleArg {
			packages = append(packages, arg)
		}
	}

	return packages
}

// isYumPackageVersionFixed is hadolint's packageVersionFixed: a name with no
// dash has no version; an .rpm file has one; a name with a variable is pinned
// when the variable is defined; otherwise what follows the first dash must
// look like a version.
func isYumPackageVersionFixed(vars versionVars, pkg string) bool {
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

	return isVersionLike(parts[1:])
}

func isVersionLike(parts []string) bool {
	if len(parts) == 0 {
		return false
	}

	allValid := true
	hasDigitStart := false

	for _, part := range parts {
		if !isValidVersionPart(part) {
			allValid = false

			break
		}

		if part != "" && unicode.IsDigit(rune(part[0])) {
			hasDigitStart = true
		}
	}

	return allValid && hasDigitStart
}

// isValidVersionPart reports whether every character of part may stand in a
// version. A part with no character is valid, as in hadolint: two dashes
// side by side, which a brace expansion between them leaves
// (httpd-{devel,tools}-2.4.6), do not unmake a version.
func isValidVersionPart(part string) bool {
	for _, ch := range part {
		if !isVersionChar(ch) {
			return false
		}
	}

	return true
}

func isVersionChar(char rune) bool {
	return unicode.IsDigit(char) ||
		unicode.IsUpper(char) ||
		unicode.IsLower(char) ||
		char == '.' ||
		char == '~' ||
		char == '^' ||
		char == '_' ||
		char == ':' ||
		char == '+'
}

func isYumModuleVersionFixed(mod string) bool {
	return strings.Contains(mod, ":")
}
