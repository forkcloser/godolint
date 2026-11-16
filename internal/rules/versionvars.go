package rules

import (
	"maps"
	"slices"
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// versionVars is the set of variables a RUN may write a package version
// with, as hadolint's DL3033 and DL3041 keep it: the ENVs of the current
// stage, where a stage built FROM an earlier stage's alias starts with that
// stage's ENVs, and every ARG of the file.
type versionVars struct {
	stage   int
	aliases map[int]string
	envs    map[int]map[string]bool
	args    map[string]bool
}

// newVersionVars is the state before any instruction: one unnamed stage
// with nothing set.
func newVersionVars() versionVars {
	return versionVars{
		stage:   0,
		aliases: map[int]string{0: ""},
		envs:    map[int]map[string]bool{0: {}},
		args:    map[string]bool{},
	}
}

// withEnvs is v with pairs' keys set in the current stage.
func (v versionVars) withEnvs(pairs []syntax.EnvPair) versionVars {
	envs := make(map[int]map[string]bool, len(v.envs))
	maps.Copy(envs, v.envs)

	stage := make(map[string]bool, len(envs[v.stage])+len(pairs))
	maps.Copy(stage, envs[v.stage])

	for _, pair := range pairs {
		stage[pair.Key] = true
	}

	envs[v.stage] = stage

	return versionVars{stage: v.stage, aliases: v.aliases, envs: envs, args: v.args}
}

// withArg is v with name declared.
func (v versionVars) withArg(name string) versionVars {
	args := make(map[string]bool, len(v.args)+1)
	maps.Copy(args, v.args)
	args[name] = true

	return versionVars{stage: v.stage, aliases: v.aliases, envs: v.envs, args: args}
}

// withStage is v in the stage from opens: a new index, the alias if any,
// and the ENVs of the stage whose alias the image names, or none.
func (v versionVars) withStage(from *syntax.From) versionVars {
	next := v.stage + 1

	aliases := make(map[int]string, len(v.aliases)+1)
	maps.Copy(aliases, v.aliases)

	if from.Image.Alias != nil {
		aliases[next] = *from.Image.Alias
	} else {
		aliases[next] = ""
	}

	inherited := map[string]bool{}

	// The last stage under that alias, when two have it: a build takes a
	// name for the latest stage given it. In the order of the stages, not
	// of the map, which changes from one run to the next.
	for _, idx := range slices.Backward(slices.Sorted(maps.Keys(v.aliases))) {
		if alias := v.aliases[idx]; alias == from.Image.Image && alias != "" {
			inherited = v.envs[idx]

			break
		}
	}

	envs := make(map[int]map[string]bool, len(v.envs)+1)
	maps.Copy(envs, v.envs)
	envs[next] = inherited

	return versionVars{stage: next, aliases: aliases, envs: envs, args: v.args}
}

// defines reports whether text refers to one of the variables, as ${name} or
// as $name followed by nothing that could continue the name.
func (v versionVars) defines(text string) bool {
	for name := range v.envs[v.stage] {
		if refersTo(text, name) {
			return true
		}
	}

	for name := range v.args {
		if refersTo(text, name) {
			return true
		}
	}

	return false
}

// refersTo reports whether text contains a reference to the variable name.
func refersTo(text, name string) bool {
	if strings.Contains(text, "${"+name+"}") {
		return true
	}

	for rest := text; ; {
		idx := strings.Index(rest, "$"+name)
		if idx < 0 {
			return false
		}

		end := idx + 1 + len(name)
		if end == len(rest) || !isNameChar(rest[end]) {
			return true
		}

		rest = rest[end:]
	}
}

// isNameChar reports whether c can continue a shell variable name.
func isNameChar(c byte) bool {
	return c == '_' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// checkPinnedVersions is the Check of a rule that pins versions through
// versionVars: ENV, ARG and FROM feed the state, and a RUN fails as meta
// unless pinned holds for its commands.
func checkPinnedVersions(
	meta rule.Meta,
	line int,
	state rule.State,
	instruction syntax.Instruction,
	pinned func(versionVars, *shell.ParsedShell) bool,
) rule.State {
	vars := rule.Data[versionVars](state)

	switch instr := instruction.(type) {
	case *syntax.Env:
		return state.ReplaceData(vars.withEnvs(instr.Pairs))
	case *syntax.Arg:
		return state.ReplaceData(vars.withArg(instr.ArgName))
	case *syntax.From:
		return state.ReplaceData(vars.withStage(instr))
	case *syntax.Run:
		parsed, err := shell.ParseShell(instr.Command)
		if err != nil || pinned(vars, parsed) {
			return state
		}

		return state.AddFailure(rule.CheckFailure{
			Code:     meta.Code,
			Severity: meta.Severity,
			Message:  meta.Message,
			Line:     line,
			Column:   1, // Hardcoded to 1 (matches hadolint)
		})
	default:
		return state
	}
}
