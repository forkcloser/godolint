package rules

import (
	"fmt"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3051Rule checks that labels are not empty.
type DL3051Rule struct {
	rule.StatefulRuleBase

	cfg *config.Config
}

// DL3051 creates the rule for checking empty labels.
func DL3051() rule.Rule {
	return &DL3051Rule{
		StatefulRuleBase: rule.NewStatefulRuleBase(DL3051Meta),
		cfg:              config.Default(),
	}
}

// DL3051WithConfig creates the rule with custom configuration.
func DL3051WithConfig(cfg *config.Config) rule.Rule {
	return &DL3051Rule{
		StatefulRuleBase: rule.NewStatefulRuleBase(DL3051Meta),
		cfg:              cfg,
	}
}

// InitialState returns the initial state for this rule.
func (*DL3051Rule) InitialState() rule.State {
	return rule.EmptyState(nil)
}

// Check validates that label values are not empty.
func (r *DL3051Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	label, ok := instruction.(*syntax.Label)
	if !ok {
		return state
	}

	inSchema := func(key string) bool {
		_, exists := r.cfg.LabelSchema[key]

		return exists
	}

	for _, key := range failingLabels(label.Pairs, inSchema, func(value string) bool { return value != "" }) {
		//nolint:gocritic // sprintfQuotedString: hadolint prints the key raw; %#q would escape one holding a backquote
		state = state.AddFailure(labelFailure(r, line, fmt.Sprintf("label `%s` is empty.", key)))
	}

	return state
}
