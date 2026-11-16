package rules

import (
	"strings"
	"unicode"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3048 creates a rule for checking label keys are valid.
func DL3048() rule.Rule {
	return rule.NewSimpleRule(
		DL3048Meta.Code,
		DL3048Meta.Severity,
		DL3048Meta.Message,
		checkDL3048,
	)
}

func checkDL3048(instruction syntax.Instruction) bool {
	label, ok := instruction.(*syntax.Label)
	if !ok {
		return true
	}

	for _, pair := range label.Pairs {
		if !isValidLabelKey(pair.Key) {
			return false
		}
	}

	return true
}

// isValidLabelKey checks if a label key follows Docker naming conventions.
// Ported from DL3048.hs validation logic.
func isValidLabelKey(key string) bool {
	if key == "" {
		return false
	}

	firstChar := rune(key[0])
	if !unicode.IsLower(firstChar) {
		return false
	}

	lastChar := rune(key[len(key)-1])
	if !unicode.IsLower(lastChar) && !unicode.IsDigit(lastChar) {
		return false
	}

	if strings.HasPrefix(key, "com.docker.") ||
		strings.HasPrefix(key, "io.docker.") ||
		strings.HasPrefix(key, "org.dockerproject.") {
		return false
	}

	if strings.Contains(key, "..") || strings.Contains(key, "--") {
		return false
	}

	for _, ch := range key {
		if !isValidLabelChar(ch) {
			return false
		}
	}

	return true
}

// isValidLabelChar checks if a character is valid in a label key.
func isValidLabelChar(char rune) bool {
	return unicode.IsDigit(char) ||
		unicode.IsLower(char) ||
		char == '.' ||
		char == '-' ||
		char == '_' ||
		char == '/'
}
