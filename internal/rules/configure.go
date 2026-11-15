package rules

import (
	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
)

// WithConfig returns the rule of given's code made with cfg, when the code is one
// a configuration reaches (the registries of DL3026, the label schema of
// DL3049 to DL3056 and DL3058), and given itself otherwise.
func WithConfig(given rule.Rule, cfg *config.Config) rule.Rule {
	configured := map[rule.Code]func(*config.Config) rule.Rule{
		DL3026Meta.Code: DL3026WithConfig,
		DL3049Meta.Code: DL3049WithConfig,
		DL3050Meta.Code: DL3050WithConfig,
		DL3051Meta.Code: DL3051WithConfig,
		DL3052Meta.Code: DL3052WithConfig,
		DL3053Meta.Code: DL3053WithConfig,
		DL3054Meta.Code: DL3054WithConfig,
		DL3055Meta.Code: DL3055WithConfig,
		DL3056Meta.Code: DL3056WithConfig,
		DL3058Meta.Code: DL3058WithConfig,
	}

	if withConfig, ok := configured[given.Code()]; ok {
		return withConfig(cfg)
	}

	return given
}
