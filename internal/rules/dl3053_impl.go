package rules

import (
	"fmt"
	"time"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3053Rule checks that RFC3339 timestamp labels are valid.
type DL3053Rule struct {
	rule.StatefulRuleBase

	cfg *config.Config
}

// DL3053 creates the rule for checking RFC3339 labels.
func DL3053() rule.Rule {
	return &DL3053Rule{
		StatefulRuleBase: rule.NewStatefulRuleBase(DL3053Meta),
		cfg:              config.Default(),
	}
}

// DL3053WithConfig creates the rule with custom configuration.
func DL3053WithConfig(cfg *config.Config) rule.Rule {
	return &DL3053Rule{
		StatefulRuleBase: rule.NewStatefulRuleBase(DL3053Meta),
		cfg:              cfg,
	}
}

// InitialState returns the initial state for this rule.
func (*DL3053Rule) InitialState() rule.State {
	return rule.EmptyState(nil)
}

// Check validates that RFC3339 timestamp labels are valid.
func (r *DL3053Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	label, ok := instruction.(*syntax.Label)
	if !ok {
		return state
	}

	ofKind := func(key string) bool { return r.cfg.LabelSchema[key] == config.LabelTypeRFC3339 }

	for _, key := range failingLabels(label.Pairs, ofKind, isValidRFC3339) {
		//nolint:gocritic // sprintfQuotedString: hadolint prints the key raw; %#q would escape one holding a backquote
		state = state.AddFailure(
			labelFailure(r, line, fmt.Sprintf("Label `%s` is not a valid time format - must conform to RFC3339.", key)),
		)
	}

	return state
}

// isValidRFC3339 checks if a string is a valid RFC3339 timestamp.
func isValidRFC3339(timestamp string) bool {
	_, err := time.Parse(time.RFC3339, timestamp)

	return err == nil
}
