package rules

import (
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// dl3009State tracks apt list cleanup per stage.
type dl3009State struct {
	lastFrom    *syntax.BaseImage // Last FROM instruction
	dockerClean bool              // Whether docker-clean is enabled
	stages      map[string]int    // map[alias]line - tracks stage names
	forgets     map[int]string    // map[line]alias - tracks lines that forgot cleanup
}

// stageAlias is the alias of the stage being processed, empty when the stage
// has none: what a forgotten cleanup is recorded and reported against.
func (s dl3009State) stageAlias() string {
	if s.lastFrom == nil || s.lastFrom.Alias == nil {
		return ""
	}

	return *s.lastFrom.Alias
}

// DL3009Rule checks for deletion of apt lists after apt update.
type DL3009Rule struct{}

// DL3009 creates the rule for checking apt list cleanup.
func DL3009() rule.Rule {
	return &DL3009Rule{}
}

// Code returns the rule code.
func (*DL3009Rule) Code() rule.Code {
	return DL3009Meta.Code
}

// Severity returns the rule severity.
func (*DL3009Rule) Severity() rule.Severity {
	return DL3009Meta.Severity
}

// Message returns the rule message.
func (*DL3009Rule) Message() string {
	return DL3009Meta.Message
}

// InitialState returns the initial state for this rule.
func (*DL3009Rule) InitialState() rule.State {
	return rule.EmptyState(dl3009State{
		dockerClean: true,
		stages:      make(map[string]int),
		forgets:     make(map[int]string),
	})
}

// Check tracks apt-get/apk usage and flags stages that install packages
// without cleaning the package lists afterwards.
func (*DL3009Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	currentState := rule.Data[dl3009State](state)

	switch inst := instruction.(type) {
	case *syntax.From:
		// Remember new stage
		currentState.lastFrom = &inst.Image
		currentState.dockerClean = true
		// Track which image names are referenced by FROM instructions
		// This is used to detect if a stage (by its alias) is used later
		currentState.stages[inst.Image.Image] = line

		return state.ReplaceData(currentState)

	case *syntax.Run:
		parsed, err := shell.ParseShell(inst.Command)
		if err != nil {
			return state
		}

		// Check if forgot to cleanup apt lists
		if forgotToCleanup(parsed) {
			// Skip if has cache/tmpfs mount for /var/lib/apt/lists
			if hasCacheOrTmpfsMount(inst.Flags, "/var/lib/apt/lists") {
				return state
			}

			// Skip if has cache/tmpfs mount for BOTH /var/lib/apt AND /var/cache/apt
			if hasCacheOrTmpfsMount(inst.Flags, "/var/lib/apt") &&
				hasCacheOrTmpfsMount(inst.Flags, "/var/cache/apt") {
				return state
			}

			// Record this line as forgetting cleanup
			currentState.forgets[line] = currentState.stageAlias()
		} else if disabledDockerClean(parsed) {
			currentState.dockerClean = false
		}

		return state.ReplaceData(currentState)
	}

	return state
}

// Finalize performs final checks after processing all instructions.
func (*DL3009Rule) Finalize(state rule.State) rule.State {
	currentState := rule.Data[dl3009State](state)

	finalState := state

	lastAlias := currentState.stageAlias()

	// Add failures for forgets that matter
	for line, alias := range currentState.forgets {
		// Fail if this is the last stage
		if alias == lastAlias {
			finalState = finalState.AddFailure(rule.CheckFailure{
				Code:     DL3009Meta.Code,
				Severity: DL3009Meta.Severity,
				Message:  DL3009Meta.Message,
				Line:     line,
				Column:   1, // Hardcoded to 1 (matches hadolint)
			})

			continue
		}

		// Fail if this stage is used later (alias appears in stages)
		if alias != "" {
			if _, used := currentState.stages[alias]; used {
				finalState = finalState.AddFailure(rule.CheckFailure{
					Code:     DL3009Meta.Code,
					Severity: DL3009Meta.Severity,
					Message:  DL3009Meta.Message,
					Line:     line,
					Column:   1, // Hardcoded to 1 (matches hadolint)
				})
			}
		}
	}

	return finalState
}

func forgotToCleanup(parsed *shell.ParsedShell) bool {
	hasUpdate := false
	hasCleanup := false

	for _, cmd := range parsed.PresentCommands {
		// Check for apt/apt-get/aptitude update
		if shell.CmdHasArgs("apt", []string{updateArg}, cmd) ||
			shell.CmdHasArgs("apt-get", []string{updateArg}, cmd) ||
			shell.CmdHasArgs("aptitude", []string{updateArg}, cmd) {
			hasUpdate = true
		}

		// Check for cleanup
		if shell.CmdHasAllArgs("rm", []string{recursiveForceFlag, "/var/lib/apt/lists/*"}, cmd) {
			hasCleanup = true
		}
	}

	return hasUpdate && !hasCleanup
}

func disabledDockerClean(parsed *shell.ParsedShell) bool {
	for _, cmd := range parsed.PresentCommands {
		// Check if removes docker-clean script
		if shell.CmdHasArgs("rm", []string{"/etc/apt/apt.conf.d/docker-clean"}, cmd) {
			return true
		}

		// Check if configures to keep packages
		if shell.CmdHasArgs("echo", []string{"'Binary::apt::APT::Keep-Downloaded-Packages \"true\";'"}, cmd) {
			return true
		}
	}

	return false
}
