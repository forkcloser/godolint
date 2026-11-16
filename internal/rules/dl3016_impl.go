package rules

import (
	"slices"
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3016 checks for npm install without version pinning.
func DL3016() rule.Rule {
	return rule.NewSimpleRule(
		DL3016Meta.Code,
		DL3016Meta.Severity,
		DL3016Meta.Message,
		checkDL3016,
	)
}

func checkDL3016(instruction syntax.Instruction) bool {
	run, ok := instruction.(*syntax.Run)
	if !ok {
		return true
	}

	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	return !slices.ContainsFunc(parsed.PresentCommands, forgotToPinNpmVersion)
}

func forgotToPinNpmVersion(cmd shell.Command) bool {
	if !shell.CmdHasArgs("npm", []string{installArg}, cmd) {
		return false
	}

	packages := getNpmPackages(cmd)

	// If no packages (npm install with no args), that's OK (installs from package.json)
	if len(packages) == 0 {
		return false
	}

	for _, pkg := range packages {
		if !isNpmVersionFixed(pkg) {
			return true
		}
	}

	return false
}

func getNpmPackages(cmd shell.Command) []string {
	args := shell.GetArgs(cmd)

	filtered := filterNpmFlags(args)

	return stripNpmInstallPrefix(filtered)
}

// filterNpmFlags removes npm install flags and their values.
func filterNpmFlags(args []string) []string {
	// Flags that take a value
	flagsWithValues := map[string]bool{
		"--loglevel": true,
		"--registry": true,
	}

	var result []string

	skip := false
	for _, arg := range args {
		if skip {
			skip = false

			continue
		}

		if strings.HasPrefix(arg, "-") {
			if flagsWithValues[arg] {
				skip = true
			}

			continue
		}

		result = append(result, arg)
	}

	return result
}

func stripNpmInstallPrefix(args []string) []string {
	foundInstall := false

	var result []string

	for _, arg := range args {
		if foundInstall {
			result = append(result, arg)
		} else if arg == installArg {
			foundInstall = true
		}
	}

	return result
}

func isNpmVersionFixed(pkg string) bool {
	if hasNpmGitPrefix(pkg) {
		return isVersionedGit(pkg)
	}

	if hasNpmTarballSuffix(pkg) {
		return true
	}

	if isNpmFolder(pkg) {
		return true
	}

	return hasNpmVersionSymbol(pkg)
}

//nolint:gochecknoglobals // read-only lookup table, effectively constant
var npmGitPrefixes = []string{"git://", "git+ssh://", "git+http://", "git+https://"}

func hasNpmGitPrefix(pkg string) bool {
	for _, prefix := range npmGitPrefixes {
		if strings.HasPrefix(pkg, prefix) {
			return true
		}
	}

	return false
}

//nolint:gochecknoglobals // read-only lookup table, effectively constant
var npmTarballSuffixes = []string{tarSuffix, tarGzSuffix, tgzSuffix}

func hasNpmTarballSuffix(pkg string) bool {
	for _, suffix := range npmTarballSuffixes {
		if strings.HasSuffix(pkg, suffix) {
			return true
		}
	}

	return false
}

//nolint:gochecknoglobals // read-only lookup table, effectively constant
var npmPathPrefixes = []string{"/", "./", "../", "~/"}

func isNpmFolder(pkg string) bool {
	for _, prefix := range npmPathPrefixes {
		if strings.HasPrefix(pkg, prefix) {
			return true
		}
	}

	return false
}

func isVersionedGit(pkg string) bool {
	return strings.Contains(pkg, "#")
}

func hasNpmVersionSymbol(pkg string) bool {
	pkg = dropNpmScope(pkg)

	return strings.Contains(pkg, "@")
}

func dropNpmScope(pkg string) string {
	if strings.HasPrefix(pkg, "@") {
		idx := strings.IndexByte(pkg, '/')
		if idx != -1 && idx+1 < len(pkg) {
			return pkg[idx+1:]
		}
	}

	return pkg
}
