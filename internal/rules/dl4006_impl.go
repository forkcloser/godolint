package rules

import (
	"slices"
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// dl4006State tracks whether pipefail is set.
type dl4006State struct {
	pipefailSet bool
}

// DL4006Rule checks for pipefail with pipes.
type DL4006Rule struct{}

// DL4006 creates the rule for checking pipefail.
func DL4006() rule.Rule {
	return &DL4006Rule{}
}

// Code returns the rule code.
func (*DL4006Rule) Code() rule.Code {
	return DL4006Meta.Code
}

// Severity returns the rule severity.
func (*DL4006Rule) Severity() rule.Severity {
	return DL4006Meta.Severity
}

// Message returns the rule message.
func (*DL4006Rule) Message() string {
	return DL4006Meta.Message
}

// InitialState returns the initial state for this rule.
func (*DL4006Rule) InitialState() rule.State {
	return rule.EmptyState(dl4006State{
		pipefailSet: false,
	})
}

// Check flags RUN instructions that pipe without `-o pipefail` set by a
// preceding SHELL instruction.
func (*DL4006Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	currentState := rule.Data[dl4006State](state)

	switch inst := instruction.(type) {
	case *syntax.From:
		// Reset state on new FROM
		currentState.pipefailSet = false

		return state.ReplaceData(currentState)

	case *syntax.Shell:
		// Check if SHELL sets pipefail
		if len(inst.Arguments) > 0 {
			shellCmd := strings.Join(inst.Arguments, " ")

			// Check if it's a non-POSIX shell (fish, powershell, etc)
			if isNonPosixShell(shellCmd) {
				currentState.pipefailSet = true // Skip checks for non-POSIX shells

				return state.ReplaceData(currentState)
			}

			// Check if pipefail is set
			currentState.pipefailSet = hasPipefailOption(shellCmd)
		}

		return state.ReplaceData(currentState)

	case *syntax.Run:
		// If pipefail is not set and command has pipes, fail
		if !currentState.pipefailSet && hasPipes(inst.Command) {
			return state.AddFailure(rule.CheckFailure{
				Code:     DL4006Meta.Code,
				Severity: DL4006Meta.Severity,
				Message:  DL4006Meta.Message,
				Line:     line,
				Column:   1, // Hardcoded to 1 (matches hadolint)
			})
		}

		return state
	}

	return state
}

// Finalize performs final checks after processing all instructions.
func (*DL4006Rule) Finalize(state rule.State) rule.State {
	return state
}

// isNonPosixShell checks if shell is non-POSIX (fish, powershell, etc).
func isNonPosixShell(shellCmd string) bool {
	nonPosixShells := []string{
		"/usr/bin/fish",
		"/bin/fish",
		"fish",
		"/usr/bin/pwsh",
		"/bin/pwsh",
		"pwsh",
		"powershell",
		"cmd.exe",
	}

	for _, shellName := range nonPosixShells {
		if strings.Contains(shellCmd, shellName) {
			return true
		}
	}

	return false
}

// hasPipes checks if a command contains pipes.
func hasPipes(command string) bool {
	// Parse the shell script to check for actual pipe operators
	parsed, err := shell.ParseShell(command)
	if err != nil {
		// Can't parse = can't reliably detect pipes
		return false
	}

	// Use shell package to detect pipes properly
	return shell.HasPipes(parsed)
}

// hasPipefailOption checks if a shell command sets pipefail.
func hasPipefailOption(shellCmd string) bool {
	// Parse the shell command to check for -o pipefail
	parsed, err := shell.ParseShell(shellCmd)
	if err != nil {
		return false
	}

	// The shell is known by its basename, so bash under any path counts.
	validShells := []string{"bash", "zsh", "ash"}

	for _, cmd := range parsed.PresentCommands {
		isValidShell := slices.Contains(validShells, shell.Basename(cmd.Name))

		if !isValidShell {
			continue
		}

		// Check for -o flag
		if !shell.HasFlag("o", cmd) {
			continue
		}

		// Check if pipefail is in arguments
		args := shell.GetArgs(cmd)
		if slices.Contains(args, "pipefail") {
			return true
		}
	}

	return false
}
