package rules

import (
	"slices"
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// dl3042State tracks PIP_NO_CACHE_DIR env var per stage.
type dl3042State struct {
	currentStage string
	noCacheSet   map[string]bool // map[stage]bool - tracks if PIP_NO_CACHE_DIR is set
}

// DL3042Rule checks for pip install without --no-cache-dir.
type DL3042Rule struct{}

// DL3042 creates the rule for checking pip cache usage.
func DL3042() rule.Rule {
	return &DL3042Rule{}
}

// Code returns the rule code.
func (*DL3042Rule) Code() rule.Code {
	return DL3042Meta.Code
}

// Severity returns the rule severity.
func (*DL3042Rule) Severity() rule.Severity {
	return DL3042Meta.Severity
}

// Message returns the rule message.
func (*DL3042Rule) Message() string {
	return DL3042Meta.Message
}

// InitialState returns the initial state for this rule.
func (*DL3042Rule) InitialState() rule.State {
	return rule.EmptyState(dl3042State{
		currentStage: "",
		noCacheSet:   make(map[string]bool),
	})
}

// Check flags pip install runs that do not disable the pip cache.
//
//nolint:gocognit // one case per instruction kind, each read in place; moving the case bodies out left the logic the same, only further away
func (*DL3042Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	currentState := rule.Data[dl3042State](state)

	switch inst := instruction.(type) {
	case *syntax.From:
		// Update current stage
		if inst.Image.Alias != nil {
			currentState.currentStage = *inst.Image.Alias
		} else {
			currentState.currentStage = inst.Image.Image
		}

		return state.ReplaceData(currentState)

	case *syntax.Env:
		// Check if PIP_NO_CACHE_DIR is set to truthy value
		for _, pair := range inst.Pairs {
			if pair.Key == "PIP_NO_CACHE_DIR" && isTruthy(pair.Value) {
				currentState.noCacheSet[currentState.currentStage] = true
			}
		}

		return state.ReplaceData(currentState)

	case *syntax.Run:
		// Skip if PIP_NO_CACHE_DIR is already set for this stage
		if currentState.noCacheSet[currentState.currentStage] {
			return state
		}

		// Skip if pip's cache is a cache or tmpfs mount
		if hasCacheOrTmpfsMount(inst.Flags, ".cache/pip") {
			return state
		}

		// Check if PIP_NO_CACHE_DIR is set to truthy value in the RUN command itself
		if pipNoCacheDirSetInCommand(inst.Command) {
			return state
		}

		parsed, err := shell.ParseShell(inst.Command)
		if err != nil {
			return state
		}

		// Check all pip install commands
		if slices.ContainsFunc(parsed.PresentCommands, forgotPipNoCacheDir) {
			return state.AddFailure(rule.CheckFailure{
				Code:     DL3042Meta.Code,
				Severity: DL3042Meta.Severity,
				Message:  DL3042Meta.Message,
				Line:     line,
				Column:   1, // Hardcoded to 1 (matches hadolint)
			})
		}

		return state
	}

	return state
}

// Finalize performs final checks after processing all instructions.
func (*DL3042Rule) Finalize(state rule.State) rule.State {
	return state
}

func forgotPipNoCacheDir(cmd shell.Command) bool {
	// Must be pip install
	if !isPipInstall(cmd) {
		return false
	}

	// Skip if it's a pip wrapper (pipx, pipenv)
	if isPipWrapper(cmd) {
		return false
	}

	// Check if has --no-cache-dir flag
	args := shell.GetArgs(cmd)

	return !slices.Contains(args, "--no-cache-dir")
}

func isPipInstall(cmd shell.Command) bool {
	return shell.CmdHasArgs("pip", []string{installArg}, cmd) ||
		shell.CmdHasArgs("pip2", []string{installArg}, cmd) ||
		shell.CmdHasArgs("pip3", []string{installArg}, cmd)
}

func isPipWrapper(cmd shell.Command) bool {
	// Check for pipx or pipenv in command name
	if strings.Contains(cmd.Name, "pipx") || strings.Contains(cmd.Name, "pipenv") {
		return true
	}

	// Check for python -m pipx/pipenv
	if strings.HasPrefix(cmd.Name, "python") {
		args := shell.GetArgs(cmd)
		for i, arg := range args {
			if arg == "-m" && i+1 < len(args) {
				next := args[i+1]
				if next == "pipx" || next == "pipenv" {
					return true
				}
			}
		}
	}

	return false
}

func isTruthy(value string) bool {
	truthy := map[string]bool{
		"1":    true,
		"true": true,
		"True": true,
		"TRUE": true,
		"on":   true,
		"On":   true,
		"ON":   true,
		"yes":  true,
		"Yes":  true,
		"YES":  true,
	}

	return truthy[value]
}

func pipNoCacheDirSetInCommand(command string) bool {
	// Check if PIP_NO_CACHE_DIR=<value> appears in the command
	_, after, ok := strings.Cut(command, "PIP_NO_CACHE_DIR=")
	if !ok {
		return false
	}

	// Extract the value after PIP_NO_CACHE_DIR=
	rest := after

	// Get the value (up to space or end of string)
	value := ""

	for i, ch := range rest {
		if ch == ' ' || ch == '\t' || ch == '\n' {
			value = rest[:i]

			break
		}
	}

	if value == "" {
		value = rest
	}

	// Check if value is truthy
	return isTruthy(value)
}
