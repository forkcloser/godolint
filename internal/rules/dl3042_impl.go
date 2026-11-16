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
type DL3042Rule struct {
	rule.StatefulRuleBase
}

// DL3042 creates the rule for checking pip cache usage.
func DL3042() rule.Rule {
	return &DL3042Rule{StatefulRuleBase: rule.NewStatefulRuleBase(DL3042Meta)}
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
func (r *DL3042Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	currentState := rule.Data[dl3042State](state)

	switch inst := instruction.(type) {
	case *syntax.From:
		// A stage under an alias starts with what the image it is built on
		// has: a stage built on one that turned pip's cache off has it off.
		// A stage with no alias goes by its image's name, and so has that
		// already.
		if inst.Image.Alias != nil {
			currentState.noCacheSet[*inst.Image.Alias] = currentState.noCacheSet[inst.Image.Image]
			currentState.currentStage = *inst.Image.Alias
		} else {
			currentState.currentStage = inst.Image.Image
		}

		return state.ReplaceData(currentState)

	case *syntax.Env:
		// An ENV that sets PIP_NO_CACHE_DIR to a truthy value turns the
		// cache off for the stage. Given twice in one ENV, the last value is
		// the one the build keeps.
		for _, pair := range slices.Backward(inst.Pairs) {
			if pair.Key == "PIP_NO_CACHE_DIR" {
				if isTruthy(pair.Value) {
					currentState.noCacheSet[currentState.currentStage] = true
				}

				break
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
			return state.AddFailure(rule.Fail(r, line))
		}

		return state
	}

	return state
}

func forgotPipNoCacheDir(cmd shell.Command) bool {
	// Must be pip install
	if !shell.IsPipInstall(cmd) {
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

// truthy is the values that turn PIP_NO_CACHE_DIR on.
//
//nolint:gochecknoglobals // a constant list
var truthy = []string{"1", "true", "True", "TRUE", "on", "On", "ON", "yes", "Yes", "YES"}

func isTruthy(value string) bool {
	return slices.Contains(truthy, value)
}

// pipNoCacheDirSetInCommand reports whether the RUN's own text sets
// PIP_NO_CACHE_DIR on: what follows the first `PIP_NO_CACHE_DIR=` begins
// with one of the truthy values, whatever comes after it (`=1;`, `=1 pip`).
func pipNoCacheDirSetInCommand(command string) bool {
	_, value, found := strings.Cut(command, "PIP_NO_CACHE_DIR=")

	return found && slices.ContainsFunc(truthy, func(word string) bool { return strings.HasPrefix(value, word) })
}
