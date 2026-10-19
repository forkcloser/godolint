package rules

// The argv vocabulary the DL30xx rules share. Around twenty rules match the
// same package-manager subcommands and flags, so the words are named once
// here, beside the per-manager binaries the rules that need only one keep
// locally (yumCommand, zypperCommand): a rule matches a literal or it
// silently stops firing, and a typo in a literal is invisible.
const (
	// dnf and microdnf are matched together by every DL304x rule: microdnf is
	// the minimal-container build of dnf and takes the same subcommands.
	dnfCommand      = "dnf"
	microdnfCommand = "microdnf"

	// The subcommands that install packages, clean a cache, or refresh an
	// index. Spelled the same by apt, apk, gem, npm, pip, yum, zypper and dnf.
	installArg = "install"
	addArg     = "add"
	updateArg  = "update"
	cleanArg   = "clean"
	moduleArg  = "module"
	allArg     = "all"

	// The empty image, which is also the stage name that would shadow it.
	scratchImage = "scratch"

	// The flags of the `rm` that the cache-cleaning rules expect in the same
	// RUN as the install.
	recursiveForceFlag = "-rf"

	// Archive suffixes several package rules read as a local artifact.
	tarSuffix   = ".tar"
	tarGzSuffix = ".tar.gz"
	tgzSuffix   = ".tgz"
)
