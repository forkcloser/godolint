package rules

import (
	"maps"
	"strings"
	"unicode"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// dl3006State tracks FROM aliases to allow untagged images that reference aliases.
type dl3006State struct {
	aliases map[string]bool // Set of defined FROM aliases
}

// DL3006Rule checks that images have explicit tags.
// Ported from Hadolint.Rule.DL3006.
type DL3006Rule struct{}

// DL3006 creates the rule for checking images have explicit tags.
func DL3006() rule.Rule {
	return &DL3006Rule{}
}

// Code returns the rule code.
func (*DL3006Rule) Code() rule.Code {
	return DL3006Meta.Code
}

// Severity returns the rule severity.
func (*DL3006Rule) Severity() rule.Severity {
	return DL3006Meta.Severity
}

// Message returns the rule message.
func (*DL3006Rule) Message() string {
	return DL3006Meta.Message
}

// InitialState returns the initial state for this rule.
func (*DL3006Rule) InitialState() rule.State {
	return rule.EmptyState(dl3006State{
		aliases: make(map[string]bool),
	})
}

// Check examines FROM instructions and tracks aliases.
// Ported from the check function in DL3006.hs.
func (*DL3006Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	from, ok := instruction.(*syntax.From)
	if !ok {
		return state
	}

	currentState := rule.Data[dl3006State](state)

	// The stages before this one: a stage's own alias is not a previous stage,
	// so `FROM app-${X} AS app-final` must not match itself.
	previous := currentState.aliases

	// Add alias to set if present
	if from.Image.Alias != nil {
		newAliases := make(map[string]bool)
		maps.Copy(newAliases, currentState.aliases)

		newAliases[*from.Image.Alias] = true
		currentState.aliases = newAliases
		state = state.ReplaceData(currentState)
	}

	// Check if image needs explicit tag
	// Scratch image - OK
	if from.Image.Image == scratchImage {
		return state
	}

	// Has digest - OK
	if from.Image.Digest != nil {
		return state
	}

	// Has tag - OK
	if from.Image.Tag != nil {
		return state
	}

	// A previous stage, or a variable that may name one - OK
	if isPreviousStage(dropQuotes(from.Image.Image), previous) {
		return state
	}

	// No tag and not an exception - fail
	return state.AddFailure(rule.CheckFailure{
		Code:     DL3006Meta.Code,
		Severity: DL3006Meta.Severity,
		Message:  DL3006Meta.Message,
		Line:     line,
		Column:   1, // Hardcoded to 1 (matches hadolint)
	})
}

// Finalize performs final checks after processing all instructions.
func (*DL3006Rule) Finalize(state rule.State) rule.State {
	return state // No finalization needed
}

// isPreviousStage reports whether image refers to a stage in aliases. An
// image that starts with a variable cannot be resolved and is taken for one;
// a variable elsewhere in the name (base-${TARGETARCH}) is a wildcard, matched
// against the aliases.
// Ported from isPreviousStage in DL3006.hs.
func isPreviousStage(image string, aliases map[string]bool) bool {
	if strings.HasPrefix(image, "$") {
		return true
	}

	if !strings.Contains(image, "$") {
		return aliases[image]
	}

	literals := literalParts(image)

	for alias := range aliases {
		if matchesLiterals(literals, alias) {
			return true
		}
	}

	return false
}

// literalParts is the text of name around its variables, in order:
// base-${ARCH}-slim gives "base-" and "-slim". Never empty: a name with no
// variable is its own single part.
func literalParts(name string) []string {
	var parts []string

	for {
		literal, rest, found := strings.Cut(name, "$")
		parts = append(parts, literal)

		if !found {
			return parts
		}

		if braced, ok := strings.CutPrefix(rest, "{"); ok {
			_, name, _ = strings.Cut(braced, "}")
		} else {
			name = strings.TrimLeftFunc(rest, func(char rune) bool {
				return unicode.IsLetter(char) || unicode.IsNumber(char) || char == '_'
			})
		}
	}
}

// matchesLiterals reports whether name is the literals in order with anything,
// or nothing, between them: the first a prefix, the last a suffix.
func matchesLiterals(literals []string, name string) bool {
	if len(literals) == 0 {
		return name == ""
	}

	rest, ok := strings.CutPrefix(name, literals[0])
	if !ok {
		return false
	}

	middle := literals[1:]
	if len(middle) == 0 {
		return rest == ""
	}

	for _, literal := range middle[:len(middle)-1] {
		_, rest, ok = strings.Cut(rest, literal)
		if !ok {
			return false
		}
	}

	return strings.HasSuffix(rest, middle[len(middle)-1])
}
