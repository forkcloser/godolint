package rules

import (
	"slices"
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3013 checks for pip install without version pinning.
func DL3013() rule.Rule {
	return rule.NewSimpleRule(
		DL3013Meta.Code,
		DL3013Meta.Severity,
		DL3013Meta.Message,
		checkDL3013,
	)
}

func checkDL3013(instruction syntax.Instruction) bool {
	run, ok := instruction.(*syntax.Run)
	if !ok {
		return true
	}

	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	// Check all commands for pip install without version pinning
	return !slices.ContainsFunc(parsed.PresentCommands, forgotToPinPipVersion)
}

func forgotToPinPipVersion(cmd shell.Command) bool {
	// Must be a pip install command
	if !shell.IsPipInstall(cmd) {
		return false
	}

	// Installing from requirements file is OK
	if isRequirementInstall(cmd) {
		return false
	}

	// Using constraint files is OK
	if hasBuildConstraint(cmd) {
		return false
	}

	// Check all packages have version pins
	packages := getPipPackages(cmd)
	for _, pkg := range packages {
		if !isPipVersionFixed(pkg) {
			return true
		}
	}

	return false
}

func isRequirementInstall(cmd shell.Command) bool {
	args := shell.GetArgs(cmd)
	for _, arg := range args {
		if arg == "--requirement" || arg == "-r" || arg == "." {
			return true
		}
	}

	return false
}

func hasBuildConstraint(cmd shell.Command) bool {
	return shell.HasFlag("constraint", cmd) || shell.HasFlag("c", cmd)
}

func getPipPackages(cmd shell.Command) []string {
	// Get all arguments
	args := shell.GetArgs(cmd)

	// Remove flags that take values
	filtered := filterPipFlags(args)

	// Strip "install" prefix
	packages := stripPipInstallPrefix(filtered)

	return packages
}

// filterPipFlags removes pip install flags and their values.
func filterPipFlags(args []string) []string {
	// Flags that take a value
	flagsWithValues := map[string]bool{
		"--trusted-host": true, "--abi": true, "--build": true, "-b": true,
		"--editable": true, "-e": true, "--extra-index-url": true,
		"--find-links": true, "-f": true, "--index-url": true, "-i": true,
		"--implementation": true, "--no-binary": true, "--only-binary": true,
		"--platform": true, "--prefix": true, "--progress-bar": true,
		"--proxy": true, "--python": true, "--python-version": true, "--root": true,
		"--root-user-action": true,
		"--src":              true, "--target": true, "-t": true, "--upgrade-strategy": true,
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

func stripPipInstallPrefix(args []string) []string {
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

func isPipVersionFixed(pkg string) bool {
	// Has version symbol
	if hasVersionSymbol(pkg) {
		return true
	}

	// Is versioned VCS
	if isVersionedVcs(pkg) {
		return true
	}

	// Is local package file
	if isLocalPackage(pkg) {
		return true
	}

	// Is non-VCS path source
	if isNoVcsPathSource(pkg) {
		return true
	}

	return false
}

//nolint:gochecknoglobals // read-only lookup table, effectively constant
var pipVersionSymbols = []string{"==", ">=", "<=", ">", "<", "!=", "~=", "==="}

func hasVersionSymbol(pkg string) bool {
	for _, symbol := range pipVersionSymbols {
		if strings.Contains(pkg, symbol) {
			return true
		}
	}

	return false
}

//nolint:gochecknoglobals // read-only lookup table, effectively constant
var pipVcsSchemes = []string{
	"git+file", "git+https", "git+ssh", "git+http", "git+git", "git",
	"hg+file", "hg+http", "hg+https", "hg+ssh", "hg+static-http",
	"svn", "svn+svn", "svn+http", "svn+https", "svn+ssh",
	"bzr+http", "bzr+https", "bzr+ssh", "bzr+sftp", "bzr+ftp", "bzr+lp",
}

func isVcs(pkg string) bool {
	for _, scheme := range pipVcsSchemes {
		if strings.HasPrefix(pkg, scheme) {
			return true
		}
	}

	return false
}

func isVersionedVcs(pkg string) bool {
	return isVcs(pkg) && strings.Contains(pkg, "@")
}

//nolint:gochecknoglobals // read-only lookup table, effectively constant
var pipLocalPackageExtensions = []string{".whl", tarGzSuffix}

func isLocalPackage(pkg string) bool {
	for _, ext := range pipLocalPackageExtensions {
		if strings.HasSuffix(pkg, ext) {
			return true
		}
	}

	return false
}

func isNoVcsPathSource(pkg string) bool {
	return !isVcs(pkg) && strings.Contains(pkg, "/")
}
