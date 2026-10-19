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

	// Check all commands for npm install without version pinning
	return !slices.ContainsFunc(parsed.PresentCommands, forgotToPinNpmVersion)
}

func forgotToPinNpmVersion(cmd shell.Command) bool {
	// Must be npm install
	if !shell.CmdHasArgs("npm", []string{installArg}, cmd) {
		return false
	}

	// Get packages
	packages := getNpmPackages(cmd)

	// If no packages (npm install with no args), that's OK (installs from package.json)
	if len(packages) == 0 {
		return false
	}

	// Check all packages have version pins
	for _, pkg := range packages {
		if !isNpmVersionFixed(pkg) {
			return true
		}
	}

	return false
}

func getNpmPackages(cmd shell.Command) []string {
	// Get all arguments
	args := shell.GetArgs(cmd)

	// Remove flags that take values
	filtered := filterNpmFlags(args)

	// Strip "install" prefix
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

		// Skip flags that start with -
		if strings.HasPrefix(arg, "-") {
			// Check if this flag takes a value
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
	// Find "install" and skip to packages after it
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
	// Git URL with commit/tag
	if hasNpmGitPrefix(pkg) {
		return isVersionedGit(pkg)
	}

	// Tarball file
	if hasNpmTarballSuffix(pkg) {
		return true
	}

	// Folder path
	if isNpmFolder(pkg) {
		return true
	}

	// Package with version symbol (@)
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
	// Drop scope prefix if present
	pkg = dropNpmScope(pkg)

	return strings.Contains(pkg, "@")
}

func dropNpmScope(pkg string) string {
	// If starts with @, drop until first /
	if strings.HasPrefix(pkg, "@") {
		idx := strings.IndexByte(pkg, '/')
		if idx != -1 && idx+1 < len(pkg) {
			return pkg[idx+1:]
		}
	}

	return pkg
}
