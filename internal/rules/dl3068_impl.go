package rules

import (
	"slices"
	"strings"
	"unicode"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3068 is the rule that pecl installs name a version: `<name>-<version>`,
// with or without a channel (`channel/name-1.2`, `channel://host/name-1.2`),
// where the version starts with a digit or is a shell variable; a stability
// word (`-beta`) floats and is not a version. A local archive or descriptor
// (`.tgz`, `.tar`, `.tar.gz`, `.tar.bz2`, `.zip`, `.xml`) and an http(s) URL
// are fixed already. The value of `-D`/`--define` is not a package.
// Ported from Hadolint.Rule.DL3068.
func DL3068() rule.Rule {
	return rule.NewSimpleRule(
		DL3068Meta.Code,
		DL3068Meta.Severity,
		DL3068Meta.Message,
		checkDL3068,
	)
}

// checkDL3068 passes a RUN whose every `pecl install` pins its packages.
func checkDL3068(instruction syntax.Instruction) bool {
	run, ok := instruction.(*syntax.Run)
	if !ok {
		return true
	}

	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	return !slices.ContainsFunc(parsed.PresentCommands, peclForgotToPin)
}

// peclForgotToPin is a `pecl install` whose first word is install and whose
// packages are not all at a version.
func peclForgotToPin(cmd shell.Command) bool {
	if !shell.CmdHasArgs("pecl", []string{installArg}, cmd) {
		return false
	}

	// -D and --define take a value argument that is not a package.
	args := shell.GetArgsNoFlags(shell.DropFlagArg([]string{"D", "define"}, cmd))
	if len(args) == 0 || args[0] != installArg {
		return false
	}

	for _, pkg := range peclPackages(args) {
		if !peclVersionFixed(pkg) {
			return true
		}
	}

	return false
}

// peclPackages are the arguments after the install word(s).
func peclPackages(args []string) []string {
	i := 0
	for i < len(args) && args[i] != installArg {
		i++
	}

	for i < len(args) && args[i] == installArg {
		i++
	}

	return args[i:]
}

// peclVersionFixed: a local artifact, a URL, or a name-version with a version.
func peclVersionFixed(pkg string) bool {
	if peclIsLocalArtifact(pkg) || strings.HasPrefix(pkg, "http://") || strings.HasPrefix(pkg, "https://") {
		return true
	}

	return peclHasPinnedVersion(peclDropChannel(pkg))
}

// peclHasPinnedVersion: the part after the last `-` starts with a digit or
// `$`; no `-`, or nothing after it, is no version.
func peclHasPinnedVersion(pkg string) bool {
	i := strings.LastIndex(pkg, "-")
	if i < 0 || i+1 >= len(pkg) {
		return false
	}

	first := rune(pkg[i+1])

	return unicode.IsDigit(first) || first == '$'
}

// peclDropChannel keeps the `<name>-<version>` segment: what follows a
// `://` scheme, then what follows the last `/`; a package without `/` is
// returned as it is.
func peclDropChannel(pkg string) string {
	bare := pkg
	if _, after, ok := strings.Cut(bare, "://"); ok && after != "" {
		bare = after
	}

	if i := strings.LastIndex(bare, "/"); i >= 0 {
		return bare[i+1:]
	}

	return pkg
}

// peclIsLocalArtifact is a release archive or a package.xml descriptor.
func peclIsLocalArtifact(pkg string) bool {
	for _, suffix := range []string{tgzSuffix, tarSuffix, tarGzSuffix, ".tar.bz2", ".zip", ".xml"} {
		if strings.HasSuffix(pkg, suffix) {
			return true
		}
	}

	return false
}
