package rules

import (
	"maps"
	"slices"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// dl3060State is the stages so far and the RUNs that left a yarn cache
// behind, each with the stage it stands in.
type dl3060State struct {
	// current is the stage being read, nil before any FROM.
	current *syntax.BaseImage
	// bases is the image names the stages are built on: a stage whose alias
	// is among them is built on later.
	bases map[string]bool
	// left is the lines of the RUNs that left a cache, with their stage.
	left map[int]*syntax.BaseImage
}

// DL3060Rule checks for yarn cache clean after yarn install.
type DL3060Rule struct {
	rule.StatefulRuleBase
}

// DL3060 checks for yarn cache clean after yarn install, in the stages that
// reach the image: the last one, and those a later stage is built on.
func DL3060() rule.Rule {
	return &DL3060Rule{StatefulRuleBase: rule.NewStatefulRuleBase(DL3060Meta)}
}

// InitialState returns the initial state for this rule.
func (*DL3060Rule) InitialState() rule.State {
	return rule.EmptyState(dl3060State{bases: map[string]bool{}, left: map[int]*syntax.BaseImage{}})
}

// Check remembers the stages and the RUNs that leave a yarn cache.
// Ported from DL3060.hs.
func (*DL3060Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	current := rule.Data[dl3060State](state)

	switch inst := instruction.(type) {
	case *syntax.From:
		current.current = &inst.Image
		current.bases[inst.Image.Image] = true
	case *syntax.Run:
		if checkDL3060(inst) {
			return state
		}

		current.left[line] = current.current
	default:
		return state
	}

	return state.ReplaceData(current)
}

// Finalize reports the RUNs that left a cache in a stage that reaches the
// image, in the order of their lines.
func (r *DL3060Rule) Finalize(state rule.State) rule.State {
	current := rule.Data[dl3060State](state)

	for _, line := range slices.Sorted(maps.Keys(current.left)) {
		stage := current.left[line]

		builtOn := stage != nil && stage.Alias != nil && current.bases[*stage.Alias]
		if sameStage(stage, current.current) || builtOn {
			state = state.AddFailure(rule.Fail(r, line))
		}
	}

	return state
}

// sameStage reports whether two stages are written the same, as hadolint
// compares them: image, tag, digest, alias and platform. Two FROMs written
// alike are one stage to it. No stage at all, which is where a RUN before
// the first FROM stands, is to hadolint an image written `scratch` and no
// more: a last stage written `FROM scratch` is that one.
func sameStage(left, right *syntax.BaseImage) bool {
	noStage := &syntax.BaseImage{Image: scratchImage}

	if left == nil {
		left = noStage
	}

	if right == nil {
		right = noStage
	}

	same := func(a, b *string) bool {
		if a == nil || b == nil {
			return a == b
		}

		return *a == *b
	}

	return left.Image == right.Image && same(left.Tag, right.Tag) && same(left.Digest, right.Digest) &&
		same(left.Alias, right.Alias) && same(left.Platform, right.Platform)
}

// checkDL3060 passes a RUN with no yarn install, one whose yarn cache is a
// mount, or one whose first yarn install comes before its first cache clean.
func checkDL3060(run *syntax.Run) bool {
	parsed, err := shell.ParseShell(run.Command)
	if err != nil {
		return true
	}

	if firstCommand(parsed.PresentCommands, isYarnInstall) < 0 {
		return true
	}

	if hasCacheOrTmpfsMount(run.Flags, ".cache/yarn") ||
		hasCacheOrTmpfsMount(run.Flags, "/root/.cache/yarn") {
		return true
	}

	return cleanFollowsInstall(parsed.PresentCommands, isYarnInstall, isYarnCacheClean)
}

func isYarnInstall(cmd shell.Command) bool {
	return shell.CmdHasArgs("yarn", []string{installArg}, cmd)
}

func isYarnCacheClean(cmd shell.Command) bool {
	return shell.CmdHasAllArgs("yarn", []string{"cache", cleanArg}, cmd)
}
