package rules

import (
	"maps"
	"slices"
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/shell"
	"github.com/forkcloser/godolint/internal/syntax"
)

// dl3010State is the archives the current stage has copied, each under its
// base name with the line of its COPY, and the lines of the COPYs whose
// archive a RUN went on to unpack, in any stage.
type dl3010State struct {
	archives  map[string]int
	extracted map[int]struct{}
}

// DL3010Rule checks for using COPY instead of ADD for archive extraction.
type DL3010Rule struct {
	rule.StatefulRuleBase
}

// DL3010 creates the rule for checking ADD usage for archives.
func DL3010() rule.Rule {
	return &DL3010Rule{StatefulRuleBase: rule.NewStatefulRuleBase(DL3010Meta)}
}

// InitialState returns the initial state for this rule.
func (*DL3010Rule) InitialState() rule.State {
	return rule.EmptyState(dl3010State{
		archives:  make(map[string]int),
		extracted: make(map[int]struct{}),
	})
}

// Check tracks COPYed archives and flags those later extracted by RUN
// instead of being extracted directly by ADD.
//
//nolint:gocognit // one case per instruction kind, each read in place; moving the case bodies out left the logic the same, only further away
func (*DL3010Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	currentState := rule.Data[dl3010State](state)

	switch inst := instruction.(type) {
	case *syntax.From:
		// A stage starts with no file of the one before. What was found
		// there stays found.
		currentState.archives = make(map[string]int)

		return state.ReplaceData(currentState)

	case *syntax.Copy:
		// An ADD reads the build context: it cannot stand for a COPY that
		// reads another stage.
		if inst.From != nil {
			return state
		}

		destBase := basename(inst.Destination)

		if isArchive(destBase) {
			currentState.archives[destBase] = line
		} else {
			for _, src := range inst.Source {
				srcBase := basename(src)
				if isArchive(srcBase) {
					currentState.archives[srcBase] = line
				}
			}
		}

		return state.ReplaceData(currentState)

	case *syntax.Run:
		parsed, err := shell.ParseShell(inst.Command)
		if err != nil {
			return state
		}

		for _, cmd := range parsed.PresentCommands {
			if isTarExtractCommand(cmd) || isUnzipCommand(cmd) {
				args := shell.GetArgsNoFlags(cmd)
				for _, arg := range args {
					base := basename(arg)
					if copied, tracked := currentState.archives[base]; tracked {
						currentState.extracted[copied] = struct{}{}
					}
				}
			}
		}

		return state.ReplaceData(currentState)
	}

	return state
}

// Finalize performs final checks after processing all instructions.
func (r *DL3010Rule) Finalize(state rule.State) rule.State {
	currentState := rule.Data[dl3010State](state)

	finalState := state

	// A COPY is reported once, however many of its archives are unpacked.
	for _, line := range slices.Sorted(maps.Keys(currentState.extracted)) {
		finalState = finalState.AddFailure(rule.Fail(r, line))
	}

	return finalState
}

func basename(path string) string {
	path = dropQuotes(path)

	idx := strings.LastIndexAny(path, "/\\")
	if idx != -1 {
		return path[idx+1:]
	}

	return path
}

// isArchive is defined in dl3020_impl.go and shared

func isTarExtractCommand(cmd shell.Command) bool {
	if cmd.Name != "tar" {
		return false
	}

	for i, arg := range shell.GetArgs(cmd) {
		switch {
		case arg == "--extract" || arg == "--get":
			return true
		case strings.HasPrefix(arg, "--"):
		case strings.HasPrefix(arg, "-"):
			if clusterExtracts(arg[1:]) {
				return true
			}
		case i == 0:
			// tar's old form: its first argument is its options, with no dash.
			if strings.Contains(arg, "x") {
				return true
			}
		}
	}

	return false
}

// tarValueOptions is tar's one-letter options that take a value.
const tarValueOptions = "bCfFgHIKLNTVX"

// clusterExtracts reads one-letter options written together after a dash.
// What follows an option that takes a value is the value: the x of
// -cfx.tar is in a file's name.
func clusterExtracts(cluster string) bool {
	for _, option := range cluster {
		if option == 'x' {
			return true
		}

		if strings.ContainsRune(tarValueOptions, option) {
			return false
		}
	}

	return false
}

//nolint:gochecknoglobals // read-only lookup table, effectively constant
var unzipCommands = map[string]bool{
	"unzip":      true,
	"gunzip":     true,
	"bunzip2":    true,
	"unlzma":     true,
	"unxz":       true,
	"zgz":        true,
	"uncompress": true,
	"zcat":       true,
	"gzcat":      true,
}

func isUnzipCommand(cmd shell.Command) bool {
	return unzipCommands[cmd.Name]
}
