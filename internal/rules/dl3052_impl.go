package rules

import (
	"fmt"
	"net/url"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3052Rule checks that URL labels are valid.
type DL3052Rule struct {
	rule.StatefulRuleBase

	cfg *config.Config
}

// DL3052 creates the rule for checking URL labels.
func DL3052() rule.Rule {
	return &DL3052Rule{
		StatefulRuleBase: rule.NewStatefulRuleBase(DL3052Meta),
		cfg:              config.Default(),
	}
}

// DL3052WithConfig creates the rule with custom configuration.
func DL3052WithConfig(cfg *config.Config) rule.Rule {
	return &DL3052Rule{
		StatefulRuleBase: rule.NewStatefulRuleBase(DL3052Meta),
		cfg:              cfg,
	}
}

// InitialState returns the initial state for this rule.
func (*DL3052Rule) InitialState() rule.State {
	return rule.EmptyState(nil)
}

// Check validates that URL label values are valid URLs.
func (r *DL3052Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	label, ok := instruction.(*syntax.Label)
	if !ok {
		return state
	}

	ofKind := func(key string) bool { return r.cfg.LabelSchema[key] == config.LabelTypeURL }

	for _, key := range failingLabels(label.Pairs, ofKind, isValidURL) {
		//nolint:gocritic // sprintfQuotedString: hadolint prints the key raw; %#q would escape one holding a backquote
		state = state.AddFailure(labelFailure(r, line, fmt.Sprintf("Label `%s` is not a valid URL.", key)))
	}

	return state
}

// isValidURL checks if a string is a valid URL.
func isValidURL(urlStr string) bool {
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return false
	}

	return parsedURL.Scheme != ""
}
