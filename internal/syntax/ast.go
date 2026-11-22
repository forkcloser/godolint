// Package syntax defines the AST (Abstract Syntax Tree) for Dockerfiles.
// Ported from Language.Docker.Syntax
//
//revive:disable:max-public-structs
package syntax

// Instruction is ported from Instruction in Language.Docker.Syntax.
type Instruction interface {
	// Name returns the instruction name (FROM, RUN, COPY, etc.)
	Name() string
}

// InstructionPos is ported from InstructionPos in Language.Docker.Syntax.
type InstructionPos struct {
	Instruction Instruction
	LineNumber  int
	// EndLine is the instruction's last line: its continued lines and the
	// bodies of its heredocs, down to the last delimiter.
	EndLine int
}

// BaseImage field names match Haskell: image, tag, digest, alias, platform (all lowercase).
type BaseImage struct {
	Image    string  // Image name (e.g., "debian", "alpine")
	Tag      *string // Optional tag (e.g., "latest", "3.20")
	Digest   *string // Optional digest (e.g., "sha256:...")
	Alias    *string // Optional alias for multi-stage builds (e.g., "AS builder")
	Platform *string // Optional platform (e.g., "linux/amd64")
}

// From is ported from From in Language.Docker.Syntax.
type From struct {
	Image BaseImage
}

// Name returns the instruction name.
func (*From) Name() string {
	return "FROM"
}

// Run is ported from Run in Language.Docker.Syntax.
type Run struct {
	Command string   // The shell command to execute
	Flags   []string // RUN instruction flags (e.g., --mount)
	IsJSON  bool     // true if using JSON/exec form, false if shell form
	// ScriptOffset is how many lines below the instruction's line the first
	// line of Command stands: 0, but for a heredoc's body, the one Command
	// with lines of its own in the Dockerfile.
	ScriptOffset int
}

// Name returns the instruction name.
func (*Run) Name() string {
	return "RUN"
}

// Copy is ported from Copy in Language.Docker.Syntax.
type Copy struct {
	Source      []string // Source paths
	Destination string   // Destination path
	From        *string  // Optional --from flag for multi-stage
}

// Name returns the instruction name.
func (*Copy) Name() string {
	return "COPY"
}

// Add is ported from Add in Language.Docker.Syntax.
type Add struct {
	Source      []string // Source paths or URLs
	Destination string   // Destination path
}

// Name returns the instruction name.
func (*Add) Name() string {
	return "ADD"
}

// Env is ported from Env in Language.Docker.Syntax.
type Env struct {
	Pairs []EnvPair // Environment variable key-value pairs
}

// EnvPair represents a key-value pair for ENV instruction.
type EnvPair struct {
	Key   string
	Value string
}

// Name returns the instruction name.
func (*Env) Name() string {
	return "ENV"
}

// Label is ported from Label in Language.Docker.Syntax.
type Label struct {
	Pairs []LabelPair // Label key-value pairs
}

// LabelPair represents a key-value pair for LABEL instruction.
type LabelPair struct {
	Key   string
	Value string
}

// Name returns the instruction name.
func (*Label) Name() string {
	return "LABEL"
}

// Workdir is ported from Workdir in Language.Docker.Syntax.
type Workdir struct {
	Directory string
}

// Name returns the instruction name.
func (*Workdir) Name() string {
	return "WORKDIR"
}

// User is ported from User in Language.Docker.Syntax.
type User struct {
	User string
}

// Name returns the instruction name.
func (*User) Name() string {
	return "USER"
}

// Expose is ported from Expose in Language.Docker.Syntax.
type Expose struct {
	Ports []string // Port specifications
}

// Name returns the instruction name.
func (*Expose) Name() string {
	return "EXPOSE"
}

// Volume is ported from Volume in Language.Docker.Syntax.
type Volume struct {
	Volumes []string // Volume mount points
}

// Name returns the instruction name.
func (*Volume) Name() string {
	return "VOLUME"
}

// Cmd is ported from Cmd in Language.Docker.Syntax.
type Cmd struct {
	Arguments []string // Command arguments
	IsJSON    bool     // true if using JSON/exec form, false if shell form
}

// Name returns the instruction name.
func (*Cmd) Name() string {
	return "CMD"
}

// Entrypoint is ported from Entrypoint in Language.Docker.Syntax.
type Entrypoint struct {
	Arguments []string // Entrypoint arguments
	IsJSON    bool     // true if using JSON/exec form, false if shell form
}

// Name returns the instruction name.
func (*Entrypoint) Name() string {
	return "ENTRYPOINT"
}

// Healthcheck is ported from Healthcheck in Language.Docker.Syntax.
type Healthcheck struct {
	Command string // Health check command
	// None is HEALTHCHECK NONE, which checks nothing.
	None bool
	// IsJSON is whether the command of HEALTHCHECK CMD is in JSON (exec) form.
	IsJSON bool
}

// Name returns the instruction name.
func (*Healthcheck) Name() string {
	return "HEALTHCHECK"
}

// Maintainer represents the MAINTAINER instruction.
// Haskell: Maintainer !Text (single unnamed field).
type Maintainer struct {
	MaintainerName string // Maintainer name/email
}

// Name returns the instruction name.
func (*Maintainer) Name() string {
	return "MAINTAINER"
}

// Arg represents the ARG instruction.
// Haskell: Arg !Text !(Maybe Text) (two unnamed fields: name and optional default value).
type Arg struct {
	ArgName string
	Value   *string // Optional default value
}

// Name returns the instruction name.
func (*Arg) Name() string {
	return "ARG"
}

// StopSignal is ported from StopSignal in Language.Docker.Syntax.
type StopSignal struct {
	Signal string
}

// Name returns the instruction name.
func (*StopSignal) Name() string {
	return "STOPSIGNAL"
}

// Shell is ported from Shell in Language.Docker.Syntax.
type Shell struct {
	Arguments []string
}

// Name returns the instruction name.
func (*Shell) Name() string {
	return "SHELL"
}

// OnBuild is ported from OnBuild in Language.Docker.Syntax.
type OnBuild struct {
	Inner Instruction // The wrapped instruction
}

// Name returns the instruction name.
func (*OnBuild) Name() string {
	return "ONBUILD"
}

// Comment is ported from Comment in Language.Docker.Syntax.
type Comment struct {
	Text string // Comment text (without # prefix)
}

// Name returns the instruction name.
func (*Comment) Name() string {
	return "COMMENT"
}
