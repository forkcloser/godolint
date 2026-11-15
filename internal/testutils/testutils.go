// Package testutils provides common testing utilities for rule validation.
package testutils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/forkcloser/godolint/internal/parser"
	"github.com/forkcloser/godolint/internal/process"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
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
	return lint(dockerfile, rules, false)
}

// LintDockerfileOnBuild lints as LintDockerfile does, with every RUN of the
// Dockerfile wrapped in ONBUILD, as hadolint's onBuildRuleCatches does: a
// rule about RUN holds for an ONBUILD RUN.
func LintDockerfileOnBuild(dockerfile string, rules []rule.Rule) []rule.CheckFailure {
	return lint(dockerfile, rules, true)
}

func lint(dockerfile string, rules []rule.Rule, onBuild bool) []rule.CheckFailure {
	p := parser.NewBuildkitParser()

	instructions, err := p.Parse([]byte(dockerfile))
	if err != nil {
		// Return empty if parse fails - some tests intentionally use invalid syntax
		return nil
	}

	if onBuild {
		for i, instruction := range instructions {
			if run, isRun := instruction.Instruction.(*syntax.Run); isRun {
				instructions[i].Instruction = &syntax.OnBuild{Inner: run}
			}
		}
	}

	failures := process.NewProcessor(rules).Run(instructions)
	holdToRules(failures, rules)

	return failures
}

// holdToRules panics on a failure under a rule's code that does not carry
// what the rule declares.
func holdToRules(failures []rule.CheckFailure, rules []rule.Rule) {
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

// AssertContainsViolation asserts that violations contains a failure with the
// given rule code, and that every failure is on a line, as hadolint's
// ruleCatches does: a failure at line 0 or below points at nothing.
func AssertContainsViolation(t *testing.T, violations []rule.CheckFailure, ruleCode string) {
	t.Helper()

	AssertOnLines(t, violations)

	for _, v := range violations {
		if string(v.Code) == ruleCode {
			return
		}
	}

	t.Errorf("Expected violation %s not found. Got violations: %v", ruleCode, violations)
}

// AssertOnLines asserts that every failure is on a line of the Dockerfile.
func AssertOnLines(t *testing.T, violations []rule.CheckFailure) {
	t.Helper()

	for _, v := range violations {
		if v.Line <= 0 {
			t.Errorf("Violation %s is on no line: %+v", v.Code, v)
		}
	}
}

// AssertViolationCount asserts that violations holds exactly count failures
// with the given rule code, as hadolint's failsWith does.
func AssertViolationCount(t *testing.T, violations []rule.CheckFailure, ruleCode string, count int) {
	t.Helper()

	var matched []rule.CheckFailure

	for _, v := range violations {
		if string(v.Code) == ruleCode {
			matched = append(matched, v)
		}
	}

	if len(matched) != count {
		t.Errorf("Expected exactly %d violation(s) %s, found %d: %v", count, ruleCode, len(matched), matched)
	}
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
