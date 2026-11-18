package rules

import (
	"maps"
	"slices"
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// dl4006State tracks whether pipes are safe in the stage being processed, and
// the same for every named stage so far, so that a stage built on one of them
// inherits its SHELL.
type dl4006State struct {
	current     string          // alias of the stage being processed, empty when it has none
	pipefailSet bool            // whether the current stage's shell fails a pipe
	stages      map[string]bool // pipefailSet as each named stage left it
}

// setPipefail records value for the current stage, and under its alias when
// it has one. The map is copied: a state is a value, and the previous one
// must not see the write.
func (s dl4006State) setPipefail(value bool) dl4006State {
	s.pipefailSet = value

	if s.current != "" {
		s.stages = maps.Clone(s.stages)
		s.stages[s.current] = value
	}

	return s
}

// enterStage starts a stage FROM image. It inherits the pipefail of the named
// stage it is built on; an image that is not a stage name, one behind a
// registry included, starts without it.
func (s dl4006State) enterStage(image syntax.BaseImage) dl4006State {
	next := dl4006State{stages: s.stages}
	if image.Alias != nil {
		next.current = *image.Alias
	}

	return next.setPipefail(s.stages[image.Image])
}

// DL4006Rule checks for pipefail with pipes.
type DL4006Rule struct {
	rule.StatefulRuleBase
}

// DL4006 creates the rule for checking pipefail.
func DL4006() rule.Rule {
	return &DL4006Rule{StatefulRuleBase: rule.NewStatefulRuleBase(DL4006Meta)}
}

// InitialState returns the initial state for this rule.
func (*DL4006Rule) InitialState() rule.State {
	return rule.EmptyState(dl4006State{
		stages: make(map[string]bool),
	})
}

// Check flags RUN instructions that pipe without `-o pipefail` set by a
// preceding SHELL instruction.
func (r *DL4006Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	currentState := rule.Data[dl4006State](state)

	switch inst := instruction.(type) {
	case *syntax.From:
		return state.ReplaceData(currentState.enterStage(inst.Image))

	case *syntax.Shell:
		if len(inst.Arguments) > 0 {
			shellCmd := strings.Join(inst.Arguments, " ")

			currentState = currentState.setPipefail(isNonPosixShell(inst.Arguments[0]) || hasPipefailOption(shellCmd))
		}

		return state.ReplaceData(currentState)

	case *syntax.Run:
		if !currentState.pipefailSet && hasPipes(inst.Command) {
			return state.AddFailure(rule.Fail(r, line))
		}

		return state
	}

	return state
}

// isNonPosixShell reports whether program, the first word of a SHELL, is a
// shell with no pipefail to ask for: fish, Windows' cmd, or a PowerShell,
// under any path, with or without .exe, in any case. A PowerShell is
// installed under more names than two (pwsh-preview, pwsh-lts): any name
// that begins as one is one.
func isNonPosixShell(program string) bool {
	name := strings.TrimSuffix(strings.ToLower(program[strings.LastIndexAny(program, `/\`)+1:]), ".exe")

	return name == "fish" || name == "cmd" || strings.HasPrefix(name, "pwsh") || strings.HasPrefix(name, "powershell")
}

// hasPipes checks if a command contains pipes.
func hasPipes(command string) bool {
	parsed, err := shell.ParseShell(command)
	if err != nil {
		// Can't parse = can't reliably detect pipes
		return false
	}

	return shell.HasPipes(parsed)
}

// hasPipefailOption checks if a shell command sets pipefail.
func hasPipefailOption(shellCmd string) bool {
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

		if !shell.HasFlag("o", cmd) {
			continue
		}

		args := shell.GetArgs(cmd)
		if slices.Contains(args, "pipefail") {
			return true
		}
	}

	return false
}
