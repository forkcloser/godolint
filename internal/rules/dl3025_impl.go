package rules

import (
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3025 creates a rule for checking CMD, ENTRYPOINT and HEALTHCHECK CMD use
// JSON notation.
func DL3025() rule.Rule {
	return rule.NewSimpleRule(
		DL3025Meta.Code,
		DL3025Meta.Severity,
		DL3025Meta.Message,
		checkDL3025,
	)
}

func checkDL3025(instruction syntax.Instruction) bool {
	switch inst := instruction.(type) {
	case *syntax.Cmd:
		return inst.IsJSON
	case *syntax.Entrypoint:
		return inst.IsJSON
	case *syntax.Healthcheck:
		// HEALTHCHECK NONE has no command to write in either form.
		return inst.None || inst.IsJSON
	}

	return true
}
