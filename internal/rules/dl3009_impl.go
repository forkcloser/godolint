package rules

import (
	"maps"
	"slices"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// dl3009State tracks apt list cleanup per stage.
type dl3009State struct {
	lastFrom    *syntax.BaseImage // Last FROM instruction
	dockerClean bool              // Whether docker-clean is enabled
	stages      map[string]int    // map[alias]line - tracks stage names
	// forgets is the lines of the RUNs that forgot the cleanup, each with
	// the stage it stands in, nil before any FROM.
	forgets map[int]*syntax.BaseImage
}

// DL3009Rule checks for deletion of apt lists after apt update.
type DL3009Rule struct {
	rule.StatefulRuleBase
}

// DL3009 creates the rule for checking apt list cleanup.
func DL3009() rule.Rule {
	return &DL3009Rule{StatefulRuleBase: rule.NewStatefulRuleBase(DL3009Meta)}
}

// InitialState returns the initial state for this rule.
func (*DL3009Rule) InitialState() rule.State {
	return rule.EmptyState(dl3009State{
		dockerClean: true,
		stages:      make(map[string]int),
		forgets:     make(map[int]*syntax.BaseImage),
	})
}

// Check tracks apt-get/apk usage and flags stages that install packages
// without cleaning the package lists afterwards.
func (*DL3009Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	currentState := rule.Data[dl3009State](state)

	switch inst := instruction.(type) {
	case *syntax.From:
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

		if forgotToCleanup(parsed) {
			if hasCacheOrTmpfsMount(inst.Flags, "/var/lib/apt/lists") {
				return state
			}

			if hasCacheOrTmpfsMount(inst.Flags, "/var/lib/apt") &&
				hasCacheOrTmpfsMount(inst.Flags, "/var/cache/apt") {
				return state
			}

			currentState.forgets[line] = currentState.lastFrom
		} else if disabledDockerClean(parsed) {
			currentState.dockerClean = false
		}

		return state.ReplaceData(currentState)
	}

	return state
}

// Finalize reports the forgotten cleanups of the stages that reach the
// image, in the order of their lines: the last stage, which is any stage
// written as the last FROM is, and the stages a later one is built on.
func (r *DL3009Rule) Finalize(state rule.State) rule.State {
	currentState := rule.Data[dl3009State](state)

	for _, line := range slices.Sorted(maps.Keys(currentState.forgets)) {
		stage := currentState.forgets[line]

		// No image has the empty name, so a stage with no alias is built on
		// by none.
		_, builtOn := currentState.stages[aliasOf(stage)]
		if sameStage(stage, currentState.lastFrom) || builtOn {
			state = state.AddFailure(rule.Fail(r, line))
		}
	}

	return state
}

// aliasOf is a stage's alias, empty when it has none or there is no stage.
func aliasOf(stage *syntax.BaseImage) string {
	if stage == nil || stage.Alias == nil {
		return ""
	}

	return *stage.Alias
}

func forgotToCleanup(parsed *shell.ParsedShell) bool {
	hasUpdate := false
	hasCleanup := false

	for _, cmd := range parsed.PresentCommands {
		if shell.CmdHasArgs("apt", []string{updateArg}, cmd) ||
			shell.CmdHasArgs("apt-get", []string{updateArg}, cmd) ||
			shell.CmdHasArgs("aptitude", []string{updateArg}, cmd) {
			hasUpdate = true
		}

		if shell.IsRecursiveRemoval(
			[]string{"/var/lib/apt/lists", "/var/lib/apt/lists/", "/var/lib/apt/lists/*"}, cmd,
		) {
			hasCleanup = true
		}
	}

	return hasUpdate && !hasCleanup
}

func disabledDockerClean(parsed *shell.ParsedShell) bool {
	for _, cmd := range parsed.PresentCommands {
		if shell.CmdHasArgs("rm", []string{"/etc/apt/apt.conf.d/docker-clean"}, cmd) {
			return true
		}

		if shell.CmdHasArgs("echo", []string{"'Binary::apt::APT::Keep-Downloaded-Packages \"true\";'"}, cmd) {
			return true
		}
	}

	return false
}
