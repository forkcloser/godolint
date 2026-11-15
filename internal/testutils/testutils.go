// Package testutils provides common testing utilities for rule validation.
package testutils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/forkcloser/godolint/internal/parser"
	"github.com/forkcloser/godolint/internal/process"
	"github.com/forkcloser/godolint/internal/rule"
)

// labelPlace is where a label rule's message holds the label a failure names.
const labelPlace = "<label>"

// LintDockerfile lints a Dockerfile string with the given rules and returns violations.
//
// It panics on a failure that carries the code of one of the rules without
// that rule's severity, or without its message (the label standing where the
// rule's message says `<label>`): what a rule declares is what its failures
// carry, for every rule a test lints with.
func LintDockerfile(dockerfile string, rules []rule.Rule) []rule.CheckFailure {
	p := parser.NewBuildkitParser()

	instructions, err := p.Parse([]byte(dockerfile))
	if err != nil {
		// Return empty if parse fails - some tests intentionally use invalid syntax
		return nil
	}

	processor := process.NewProcessor(rules)

	failures := processor.Run(instructions)

	for _, failure := range failures {
		for _, declared := range rules {
			if declared.Code() != failure.Code {
				continue
			}

			if failure.Severity != declared.Severity() || !carries(failure.Message, declared.Message()) {
				panic(fmt.Sprintf(
					"failure %+v does not carry what its rule declares: severity %s, message %q",
					failure, declared.Severity(), declared.Message(),
				))
			}
		}
	}

	return failures
}

// carries reports whether a failure's message is its rule's, a label standing
// where the rule's says `<label>`.
func carries(message, declared string) bool {
	before, after, named := strings.Cut(declared, labelPlace)
	if !named {
		return message == declared
	}

	return len(message) >= len(before)+len(after) &&
		strings.HasPrefix(message, before) && strings.HasSuffix(message, after)
}

// AssertContainsViolation asserts that violations contains a failure with the given rule code.
func AssertContainsViolation(t *testing.T, violations []rule.CheckFailure, ruleCode string) {
	t.Helper()

	for _, v := range violations {
		if string(v.Code) == ruleCode {
			return
		}
	}

	t.Errorf("Expected violation %s not found. Got violations: %v", ruleCode, violations)
}

// AssertNoViolation asserts that violations does NOT contain the given rule code.
func AssertNoViolation(t *testing.T, violations []rule.CheckFailure, ruleCode string) {
	t.Helper()

	for _, v := range violations {
		if string(v.Code) == ruleCode {
			t.Errorf("Unexpected violation %s found: %+v", ruleCode, v)

			return
		}
	}
}
