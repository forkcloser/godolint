package rules

import (
	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3050Rule checks for superfluous labels (labels not in schema).
type DL3050Rule struct {
	rule.StatefulRuleBase

	cfg *config.Config
}

// DL3050 creates a rule that checks for superfluous labels.
// TODO: Wire up config from main application.
func DL3050() rule.Rule {
	return &DL3050Rule{
		StatefulRuleBase: rule.NewStatefulRuleBase(DL3050Meta),
		cfg:              config.Default(), // Use default config for now
	}
}

// DL3050WithConfig creates the rule with custom configuration.
func DL3050WithConfig(cfg *config.Config) rule.Rule {
	return &DL3050Rule{
		StatefulRuleBase: rule.NewStatefulRuleBase(DL3050Meta),
		cfg:              cfg,
	}
}

// InitialState returns the initial state for this rule.
func (*DL3050Rule) InitialState() rule.State {
	return rule.EmptyState(nil)
}

// Check validates that LABEL instructions only use labels defined in schema.
func (r *DL3050Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	label, ok := instruction.(*syntax.Label)
	if !ok {
		return state
	}

	if !r.cfg.StrictLabels {
		return state
	}

	for _, pair := range label.Pairs {
		if _, exists := r.cfg.LabelSchema[pair.Key]; !exists {
			return state.AddFailure(rule.Fail(r, line))
		}
	}

	return state
}
