// gen-rules extracts metadata from hadolint Haskell rules and generates Go stubs.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"
)

// ErrRuleGeneration is the base error for rule generation failures.
var ErrRuleGeneration = errors.New("rule metadata generation error")

// RuleMetadata contains extracted metadata from hadolint rules.
type RuleMetadata struct {
	Code         string
	Severity     string
	Message      string
	SourceFile   string
	Implemented  bool
	HaskellCheck string // Raw Haskell check function for pattern detection
	CanGenerate  bool   // Whether we can auto-generate implementation
}

var (
	// Regex patterns for extracting Haskell rule metadata.
	codePattern     = regexp.MustCompile(`code\s*=\s*"(DL\d+)"`)
	severityPattern = regexp.MustCompile(`severity\s*=\s*(DL\w+)`)
	// A message is one string literal, or literals and variables joined by
	// <> (the label rules: "Label `" <> label <> "` is missing."), any literal
	// optionally under Text.pack.
	messageStartPattern   = regexp.MustCompile(`message\s*=\s*`)
	messageLiteralPattern = regexp.MustCompile(`^(?:Text\.pack\s*)?"([^"]*)"`)
	messageVarPattern     = regexp.MustCompile(`^[a-z][A-Za-z0-9_']*`)
	messageConcatPattern  = regexp.MustCompile(`^\s*<>\s*`)
	// stringGapPattern is a Haskell string gap inside a message literal.
	stringGapPattern = regexp.MustCompile(`\\\s+\\`)
)

func main() {
	if len(os.Args) != 2 {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s <hadolint-rule-dir>\n", os.Args[0])

		os.Exit(1)
	}

	hadolintRuleDir := os.Args[1]

	pattern := filepath.Join(hadolintRuleDir, "DL*.hs")

	files, err := filepath.Glob(pattern)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to glob rules: %v\n", err)

		os.Exit(1)
	}

	if len(files) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "No rule files found in %s\n", hadolintRuleDir)

		os.Exit(1)
	}

	var rules []RuleMetadata

	for _, file := range files {
		rule, err := parseRuleFile(file)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", file, err)

			continue
		}

		rules = append(rules, rule)
	}

	slices.SortFunc(rules, func(left, right RuleMetadata) int {
		return strings.Compare(left.Code, right.Code)
	})

	// Check which rules are already implemented (check for _impl.go files)
	for idx := range rules {
		implFile := strings.ToLower(rules[idx].Code) + "_impl.go"
		// The "taint" is the maintainer's own hadolint checkout, and the code
		// is regex-constrained to DL\d+ — no traversal is possible.
		// #nosec G703 -- dev-only codegen; path derives from a regex-matched rule code.
		if _, err := os.Stat(implFile); err == nil {
			rules[idx].Implemented = true
		}
	}

	fmt.Print("# Hadolint Rules Status\n\n")
	fmt.Printf("Total rules: %d\n", len(rules))

	implemented := 0

	for _, r := range rules {
		if r.Implemented {
			implemented++
		}
	}

	fmt.Printf("Implemented: %d\n", implemented)
	fmt.Printf("Not implemented: %d\n", len(rules)-implemented)
	fmt.Println()

	metadataGenerated := 0

	for _, rule := range rules {
		err := generateMetadata(rule)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to generate metadata for %s: %v\n", rule.Code, err)
		} else {
			metadataGenerated++
		}
	}

	fmt.Printf("Generated %d metadata files\n", metadataGenerated)

	implGenerated := 0

	for _, rule := range rules {
		if !rule.Implemented && rule.CanGenerate {
			err := generateImplementation(rule)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to generate implementation for %s: %v\n", rule.Code, err)
			} else {
				implGenerated++
			}
		}
	}

	fmt.Printf("Generated %d working implementations\n", implGenerated)

	fmt.Print("\n## Rules by Status\n\n")

	fmt.Printf("### Implemented (%d)\n", implemented)

	for _, r := range rules {
		if r.Implemented {
			fmt.Printf("- ✅ %s: %s\n", r.Code, r.Message)
		}
	}

	fmt.Printf("\n### Not Implemented (%d)\n", len(rules)-implemented)

	for _, r := range rules {
		if !r.Implemented {
			fmt.Printf("- ⏳ %s: %s\n", r.Code, r.Message)
		}
	}
}

func parseRuleFile(path string) (RuleMetadata, error) {
	// #nosec G304 G703
	content, err := os.ReadFile(path)
	if err != nil {
		return RuleMetadata{}, fmt.Errorf("%w: %w", ErrRuleGeneration, err)
	}

	text := string(content)

	rule := RuleMetadata{
		SourceFile: filepath.Base(path),
	}

	match := codePattern.FindStringSubmatch(text)
	if len(match) <= 1 {
		return rule, fmt.Errorf("%w: could not find rule code", ErrRuleGeneration)
	}

	rule.Code = match[1]

	match = severityPattern.FindStringSubmatch(text)
	if len(match) <= 1 {
		return rule, fmt.Errorf("%w: could not find severity", ErrRuleGeneration)
	}

	rule.Severity = mapSeverity(match[1])

	msg, ok := parseMessage(text)
	if !ok {
		return rule, fmt.Errorf("%w: could not find message", ErrRuleGeneration)
	}

	rule.Message = msg

	// Match patterns like: check (Maintainer _) = False
	//                      check _ = True
	checkPattern := regexp.MustCompile(`(?m)^\s*check\s+(.+?)\s*=`)

	checkMatches := checkPattern.FindAllStringSubmatch(text, -1)
	if len(checkMatches) > 0 {
		// Collect all check patterns
		var checkLines []string

		for _, match := range checkMatches {
			if len(match) > 1 {
				pattern := strings.TrimSpace(match[1])
				// Skip the "where" clause pattern
				if !strings.HasPrefix(pattern, "where") {
					checkLines = append(checkLines, pattern)
				}
			}
		}

		rule.HaskellCheck = strings.Join(checkLines, "\n")
	}

	// Detect if we can auto-generate this rule
	rule.CanGenerate = canGenerateImplementation(rule.HaskellCheck)

	return rule, nil
}

// parseMessage is the message a rule declares, as hadolint's documentation
// writes it: the literals as they are, and a variable joined between them as
// its name in angle brackets, so "Label `" <> label <> "` is missing." is
// "Label `<label>` is missing.". Anything else is refused rather than read in
// part: a first term that is not a literal, another operator, a function
// applied to an argument, or an expression that goes on past what was read,
// on the same line or on a deeper-indented one below.
func parseMessage(text string) (string, bool) {
	loc := messageStartPattern.FindStringIndex(text)
	if loc == nil {
		return "", false
	}

	indent := loc[0] - (strings.LastIndex(text[:loc[0]], "\n") + 1)
	rest := text[loc[1]:]

	var msg strings.Builder

	for first := true; ; first = false {
		if match := messageLiteralPattern.FindStringSubmatch(rest); match != nil {
			// A Haskell string gap (a backslash, whitespace including the
			// line break, a backslash) is deleted whole by the compiler,
			// indentation included; the message hadolint emits has one
			// space where the source breaks its line.
			literal := stringGapPattern.ReplaceAllString(match[1], "")
			_, _ = msg.WriteString(strings.ReplaceAll(literal, "\\", ""))
			rest = rest[len(match[0]):]
		} else if name := messageVarPattern.FindString(rest); name != "" && !first {
			_, _ = msg.WriteString("<" + name + ">")
			rest = rest[len(name):]
		} else {
			return "", false
		}

		concat := messageConcatPattern.FindString(rest)
		if concat == "" {
			break
		}

		rest = rest[len(concat):]
	}

	line, below, _ := strings.Cut(rest, "\n")
	if strings.TrimSpace(line) != "" || continues(below, indent) {
		return "", false
	}

	return strings.TrimSpace(msg.String()), true
}

// continues reports whether below, the text under a declaration indented by
// indent, starts with a line of that declaration: its first line that is not
// blank is indented deeper.
func continues(below string, indent int) bool {
	for line := range strings.SplitSeq(below, "\n") {
		if text := strings.TrimLeft(line, " \t"); text != "" {
			return len(line)-len(text) > indent
		}
	}

	return false
}

// canGenerateImplementation determines if we can auto-generate Go code from Haskell patterns.
func canGenerateImplementation(haskellCheck string) bool {
	if haskellCheck == "" {
		return false
	}

	lines := strings.Split(haskellCheck, "\n")

	// Only generate for the absolute simplest pattern:
	// Pattern: (InstructionName _) followed by _ (default case)
	// Example: (Maintainer _) and _
	// This means: check specific instruction type, always fail, everything else passes

	if len(lines) != 2 {
		return false
	}

	// First line must be: (InstructionName _) with ONLY underscore, no nested patterns
	simplePattern := regexp.MustCompile(`^\((\w+)\s+_\)$`)
	if !simplePattern.MatchString(strings.TrimSpace(lines[0])) {
		return false
	}

	// Second line must be: _ (default wildcard)
	if strings.TrimSpace(lines[1]) != "_" {
		return false
	}

	return true
}

func mapSeverity(haskellSeverity string) string {
	// The explicit DLWarningC case documents the hadolint mapping; Warning is
	// also the fallback for unknown severities, so default mirrors it on purpose.
	//revive:disable:identical-switch-branches intentional, see above.
	switch haskellSeverity {
	case "DLErrorC":
		return "rule.Error"
	case "DLWarningC":
		return "rule.Warning"
	case "DLInfoC":
		return "rule.Info"
	case "DLStyleC":
		return "rule.Style"
	case "DLIgnoreC":
		return "rule.Ignore"
	default:
		return "rule.Warning" // default
	}
}

// generateMetadata generates metadata-only file (always overwrites).
func generateMetadata(rule RuleMetadata) error {
	tmpl := `// Code generated by go generate; DO NOT EDIT.
// This file is auto-generated from hadolint rules.

package rules

import "github.com/forkcloser/godolint/internal/rule"

// {{.Code}}Meta contains metadata for rule {{.Code}}.
// Source: hadolint/src/Hadolint/Rule/{{.SourceFile}}
var {{.Code}}Meta = rule.Meta{
	Code:     "{{.Code}}",
	Severity: {{.Severity}},
	Message:  "{{.Message}}",
}
`

	tpl, err := template.New("metadata").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRuleGeneration, err)
	}

	filename := strings.ToLower(rule.Code) + ".go"

	// #nosec G304
	outputFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRuleGeneration, err)
	}

	defer func() {
		_ = outputFile.Close()
	}()

	err = tpl.Execute(outputFile, rule)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRuleGeneration, err)
	}

	return nil
}

// generateImplementation generates working implementation from Haskell patterns (only if file doesn't exist).
func generateImplementation(rule RuleMetadata) error {
	filename := strings.ToLower(rule.Code) + "_impl.go"

	// Check if implementation already exists
	if _, err := os.Stat(filename); err == nil {
		return nil // File exists, don't overwrite
	}

	// Generate implementation code based on detected pattern
	implCode := generateCheckFunction(rule)

	tmpl := `package rules

import (
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// {{.Code}} creates a rule from the generated metadata.
// Auto-generated from hadolint Haskell source.
func {{.Code}}() rule.Rule {
	return rule.NewSimpleRule(
		{{.Code}}Meta.Code,
		{{.Code}}Meta.Severity,
		{{.Code}}Meta.Message,
		check{{.Code}},
	)
}

{{.ImplCode}}
`

	tpl, err := template.New("impl").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRuleGeneration, err)
	}

	// #nosec G304
	outputFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRuleGeneration, err)
	}

	defer func() {
		_ = outputFile.Close()
	}()

	data := struct {
		Code     string
		ImplCode string
	}{
		Code:     rule.Code,
		ImplCode: implCode,
	}

	err = tpl.Execute(outputFile, data)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRuleGeneration, err)
	}

	return nil
}

// generateCheckFunction generates Go check function from Haskell patterns.
func generateCheckFunction(rule RuleMetadata) string {
	lines := strings.Split(rule.HaskellCheck, "\n")

	// Pattern 1: Simple instruction type check
	// Example: (Maintainer _) followed by _
	simplePattern := regexp.MustCompile(`\((\w+)\s+\w+\)`)
	if len(lines) == 2 && simplePattern.MatchString(lines[0]) && strings.TrimSpace(lines[1]) == "_" {
		match := simplePattern.FindStringSubmatch(lines[0])
		if len(match) > 1 {
			instructionType := match[1]

			return fmt.Sprintf(`func check%s(instruction syntax.Instruction) bool {
	_, ok := instruction.(*syntax.%s)
	if ok {
		return false // %s instruction found -> fail
	}
	return true
}`, rule.Code, instructionType, instructionType)
		}
	}

	// Fallback - should not reach here if canGenerateImplementation is correct
	return fmt.Sprintf(`func check%s(instruction syntax.Instruction) bool {
	// TODO: Auto-generation failed, implement manually
	return true
}`, rule.Code)
}
