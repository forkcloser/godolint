package rules

import (
	"slices"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3046 checks for useradd without -l flag with high UID.
func DL3046() rule.Rule {
	return rule.NewSimpleRule(
		DL3046Meta.Code,
		DL3046Meta.Severity,
		DL3046Meta.Message,
		checkDL3046,
	)
}

func checkDL3046(instruction syntax.Instruction) bool {
	run, ok := instruction.(*syntax.Run)
	if !ok {
		return true
	}

	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	// Check all useradd commands
	return !slices.ContainsFunc(parsed.PresentCommands, forgotUseraddFlagL)
}

// forgotUseraddFlagL reports a useradd that gives a UID of more than five
// digits and no -l: the UID after -u, alone or last in a cluster (-mu), or
// given to --uid.
func forgotUseraddFlagL(cmd shell.Command) bool {
	if cmd.Name != "useradd" || shell.HasAnyFlag([]string{"l", "no-log-init"}, cmd) {
		return false
	}

	uids := append(shell.GetFlagArg("u", cmd), shell.GetFlagArg("uid", cmd)...)

	return slices.ContainsFunc(uids, func(uid string) bool { return len(uid) > 5 })
}
