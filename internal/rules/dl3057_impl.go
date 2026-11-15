package rules

import (
	"maps"
	"slices"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// dl3057State tracks stages and their HEALTHCHECK status.
type dl3057State struct {
	currentStage *stageID
	goodStages   map[stageID]bool // stages with HEALTHCHECK or inherited
	badStages    map[stageID]bool // stages without HEALTHCHECK
}

type stageID struct {
	src  string // source image name (what this stage is FROM)
	name string // alias name (or same as src if no alias)
	line int    // line number where stage is defined
}

// DL3057Rule checks for missing HEALTHCHECK instructions.
type DL3057Rule struct{}

// DL3057 creates the rule for checking missing HEALTHCHECK.
func DL3057() rule.Rule {
	return &DL3057Rule{}
}

// Code returns the rule code.
func (*DL3057Rule) Code() rule.Code {
	return DL3057Meta.Code
}

// Severity returns the rule severity.
func (*DL3057Rule) Severity() rule.Severity {
	return DL3057Meta.Severity
}

// Message returns the rule message.
func (*DL3057Rule) Message() string {
	return DL3057Meta.Message
}

// InitialState returns the initial state for this rule.
func (*DL3057Rule) InitialState() rule.State {
	return rule.EmptyState(dl3057State{
		currentStage: nil,
		goodStages:   make(map[stageID]bool),
		badStages:    make(map[stageID]bool),
	})
}

// Check tracks stages and flags images that never declare a HEALTHCHECK.
func (*DL3057Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	currentState := rule.Data[dl3057State](state)

	switch inst := instruction.(type) {
	case *syntax.From:
		// Create stage ID
		imageName := inst.Image.Image

		stageName := imageName
		if inst.Image.Alias != nil {
			stageName = *inst.Image.Alias
		}

		newStage := stageID{
			src:  imageName,
			name: stageName,
			line: line,
		}

		// Check if this stage inherits from a good stage
		inherited := false

		for goodStage := range currentState.goodStages {
			if goodStage.name == imageName {
				inherited = true

				break
			}
		}

		if inherited {
			// Mark as good since it inherits from a good stage
			currentState.goodStages[newStage] = true
		} else {
			// Mark as bad for now (can be updated if HEALTHCHECK found)
			currentState.badStages[newStage] = true
		}

		currentState.currentStage = &newStage

		return state.ReplaceData(currentState)

	case *syntax.Healthcheck:
		// Mark current stage and all its ancestors as good
		if currentState.currentStage != nil {
			currentState = markGood(currentState, *currentState.currentStage)
		}

		return state.ReplaceData(currentState)
	}

	return state
}

// markGood marks a stage and all its ancestors as good.
func markGood(state dl3057State, stage stageID) dl3057State {
	for ancestor := range ancestors(state.badStages, stage) {
		state.goodStages[ancestor] = true
		delete(state.badStages, ancestor)
	}

	state.goodStages[stage] = true
	delete(state.badStages, stage)

	return state
}

// ancestors is the stages of bad the stage is built on, and those they are
// built on in turn: hadolint's recurseGood. Each round looks only among the
// stages no earlier round took. A stage with no alias is named by its image
// and so is built on itself: looked for among all of bad again, it would be
// found without end.
func ancestors(bad map[stageID]bool, stage stageID) map[stageID]bool {
	taken := map[stageID]bool{}
	rest := map[stageID]bool{}

	for candidate := range bad {
		if candidate.name == stage.src {
			taken[candidate] = true
		} else {
			rest[candidate] = true
		}
	}

	for _, parent := range slices.Collect(maps.Keys(taken)) {
		maps.Copy(taken, ancestors(rest, parent))
	}

	return taken
}

// Finalize performs final checks after processing all instructions.
func (*DL3057Rule) Finalize(state rule.State) rule.State {
	currentState := rule.Data[dl3057State](state)

	// Report failures for all bad stages
	for badStage := range currentState.badStages {
		state = state.AddFailure(rule.CheckFailure{
			Code:     DL3057Meta.Code,
			Severity: DL3057Meta.Severity,
			Message:  DL3057Meta.Message,
			Line:     badStage.line,
			Column:   1, // Hardcoded to 1 (matches hadolint)
		})
	}

	return state
}
