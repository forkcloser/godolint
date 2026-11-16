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
type DL3010Rule struct{}

// DL3010 creates the rule for checking ADD usage for archives.
func DL3010() rule.Rule {
	return &DL3010Rule{}
}

// Code returns the rule code.
func (*DL3010Rule) Code() rule.Code {
	return DL3010Meta.Code
}

// Severity returns the rule severity.
func (*DL3010Rule) Severity() rule.Severity {
	return DL3010Meta.Severity
}

// Message returns the rule message.
func (*DL3010Rule) Message() string {
	return DL3010Meta.Message
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

		// If destination looks like an archive file, track its basename
		if isArchive(destBase) {
			currentState.archives[destBase] = line
		} else {
			// If destination is a directory, track source basenames
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

		// Check if any tracked archives are extracted
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
func (*DL3010Rule) Finalize(state rule.State) rule.State {
	currentState := rule.Data[dl3010State](state)

	finalState := state

	// A COPY is reported once, however many of its archives are unpacked.
	for _, line := range slices.Sorted(maps.Keys(currentState.extracted)) {
		finalState = finalState.AddFailure(rule.CheckFailure{
			Code:     DL3010Meta.Code,
			Severity: DL3010Meta.Severity,
			Message:  DL3010Meta.Message,
			Line:     line,
			Column:   1, // Hardcoded to 1 (matches hadolint)
		})
	}

	return finalState
}

func basename(path string) string {
	// Remove quotes
	path = dropQuotes(path)

	// Get last component after / or \
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

	args := shell.GetArgs(cmd)

	// Check for long extract flags
	for _, arg := range args {
		if arg == "--extract" || arg == "--get" {
			return true
		}
	}

	// Check for short extract flags (-x)
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") && strings.Contains(arg, "x") {
			return true
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
