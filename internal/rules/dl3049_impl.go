package rules

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// dl3049State follows each required label through the stages. A stage is
// known by the line of its FROM. For a label, a stage is good when it carries
// the label, itself or through the stage it is built on, and missing
// otherwise, until a stage built on it sets the label or a COPY --from takes
// from it: a stage that only feeds another is not the image that ships.
type dl3049State struct {
	current int                      // line of the current stage's FROM, 0 before any
	stages  map[int]syntax.BaseImage // every stage so far, by line
	good    map[dl3049Mark]bool      // the stages that carry a label
	missing map[dl3049Mark]bool      // the stages to report a label on
}

// dl3049Mark is one required label on one stage.
type dl3049Mark struct {
	label string
	line  int
}

// enter starts the stage FROM image at line: for each label, good when the
// image is the alias of a good stage, missing otherwise.
func (s dl3049State) enter(line int, image syntax.BaseImage, labels []string) {
	for _, label := range labels {
		mark := dl3049Mark{label: label, line: line}
		s.missing[mark] = true

		for parent := range s.good {
			alias := s.stages[parent.line].Alias
			if parent.label == label && alias != nil && *alias == image.Image {
				delete(s.missing, mark)

				s.good[mark] = true

				break
			}
		}
	}

	s.stages[line] = image
}

// excuse stops reporting label as missing on the stages aliased name.
func (s dl3049State) excuse(label, name string) {
	for mark := range s.missing {
		if alias := s.stages[mark.line].Alias; mark.label == label && alias != nil && *alias == name {
			delete(s.missing, mark)
		}
	}
}

// name is what a later instruction calls the current stage: its alias, or
// its image when it has none, which names the stage it is built on.
func (s dl3049State) name() string {
	image := s.stages[s.current]
	if image.Alias != nil {
		return *image.Alias
	}

	return image.Image
}

// DL3049Rule checks for missing required labels.
type DL3049Rule struct {
	cfg *config.Config
}

// DL3049 creates a rule that checks for missing required labels.
func DL3049() rule.Rule {
	return &DL3049Rule{
		cfg: config.Default(),
	}
}

// DL3049WithConfig creates the rule with custom configuration.
func DL3049WithConfig(cfg *config.Config) rule.Rule {
	return &DL3049Rule{
		cfg: cfg,
	}
}

// Code returns the rule code.
func (*DL3049Rule) Code() rule.Code {
	return DL3049Meta.Code
}

// Severity returns the rule severity.
func (*DL3049Rule) Severity() rule.Severity {
	return DL3049Meta.Severity
}

// Message returns the rule message.
func (*DL3049Rule) Message() string {
	return DL3049Meta.Message
}

// InitialState returns the initial state for this rule.
func (*DL3049Rule) InitialState() rule.State {
	return rule.EmptyState(dl3049State{
		stages:  make(map[int]syntax.BaseImage),
		good:    make(map[dl3049Mark]bool),
		missing: make(map[dl3049Mark]bool),
	})
}

// Check follows the required labels through the stages; Finalize reports,
// at its FROM, each stage a label is still missing from.
// Ported from missingLabelRule in DL3049.hs.
func (r *DL3049Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	current := rule.Data[dl3049State](state)

	switch inst := instruction.(type) {
	case *syntax.From:
		current.enter(line, inst.Image, slices.Collect(maps.Keys(r.cfg.LabelSchema)))
		current.current = line

	case *syntax.Copy:
		if inst.From == nil || current.current == 0 {
			return state
		}

		for label := range r.cfg.LabelSchema {
			current.excuse(label, *inst.From)
		}

	case *syntax.Label:
		if current.current == 0 {
			return state
		}

		// A label nobody requires leaves marks nothing reads.
		for _, pair := range inst.Pairs {
			mark := dl3049Mark{label: pair.Key, line: current.current}

			current.excuse(pair.Key, current.name())
			current.good[mark] = true
			delete(current.missing, mark)
		}

	default:
		return state
	}

	return state.ReplaceData(current)
}

// Finalize reports each label on each stage it is missing from: label by
// label, each in the order of the stages.
func (*DL3049Rule) Finalize(state rule.State) rule.State {
	current := rule.Data[dl3049State](state)

	marks := slices.SortedFunc(maps.Keys(current.missing), func(a, b dl3049Mark) int {
		return cmp.Or(strings.Compare(a.label, b.label), cmp.Compare(a.line, b.line))
	})

	for _, mark := range marks {
		state = state.AddFailure(rule.CheckFailure{
			Code:     DL3049Meta.Code,
			Severity: DL3049Meta.Severity,
			//nolint:gocritic // sprintfQuotedString: hadolint prints the key raw; %#q would escape one holding a backquote
			Message: fmt.Sprintf("Label `%s` is missing.", mark.label),
			Line:    mark.line,
			Column:  1, // Hardcoded to 1 (matches hadolint)
		})
	}

	return state
}
