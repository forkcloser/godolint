// Package process orchestrates running hadolint rules over a parsed Dockerfile AST.
// Ported from Hadolint/Process.hs
package process

import (
	"github.com/forkcloser/godolint/internal/pragma"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/syntax"
)

// Processor runs rules against a Dockerfile AST and collects violations.
type Processor struct {
	rules                []rule.Rule
	disableIgnorePragmas bool
}

// NewProcessor creates a new processor with the given rules.
func NewProcessor(ruleSet []rule.Rule) *Processor {
	return &Processor{
		rules:                ruleSet,
		disableIgnorePragmas: false,
	}
}

// WithDisableIgnorePragmas configures whether the ignore pragmas (line, stage
// and global) are read at all.
func (p *Processor) WithDisableIgnorePragmas(disable bool) *Processor {
	p.disableIgnorePragmas = disable

	return p
}

// Run processes a Dockerfile AST and returns all rule violations found.
// Uses fold-style accumulation with state for each rule.
// Ported from Hadolint's Rule fold pattern.
func (p *Processor) Run(instructions []syntax.InstructionPos) []rule.CheckFailure {
	allFailures := []rule.CheckFailure{}

	// For each rule, fold over all instructions with state. A rule hadolint
	// composes with its onbuild combinator runs a second time, with a state
	// of its own, over the instructions ONBUILD wraps and no other: an
	// ONBUILD RUN is no RUN of the stage it stands in. Any other rule never
	// looks inside an ONBUILD.
	for _, currentRule := range p.rules {
		state := currentRule.InitialState()

		inOnBuild := rules.InOnBuild(currentRule)
		wrapped := currentRule.InitialState()

		for _, instrPos := range instructions {
			state = currentRule.Check(instrPos.LineNumber, state, instrPos.Instruction)

			if onbuild, ok := instrPos.Instruction.(*syntax.OnBuild); ok && inOnBuild {
				wrapped = currentRule.Check(instrPos.LineNumber, wrapped, onbuild.Inner)
			}
		}

		// Finalize the state (some rules add failures only at the end)
		allFailures = append(allFailures, currentRule.Finalize(state).Failures...)

		if inOnBuild {
			allFailures = append(allFailures, currentRule.Finalize(wrapped).Failures...)
		}
	}

	// Drop what a line, stage or global ignore pragma silences.
	if !p.disableIgnorePragmas {
		directives := pragma.Parse(instructions)
		allFailures = filterIgnored(allFailures, directives)
	}

	return allFailures
}

// WithoutIgnored drops the failures of Ignore severity, as hadolint's Lint
// does after Process has run: a rule such as DL3057 reports them, a test
// sees them, and a linter's output does not.
func WithoutIgnored(failures []rule.CheckFailure) []rule.CheckFailure {
	return Settle(failures, nil, nil)
}

// Settle gives the failures of a run the severity overrides names for their
// code, then drops those whose code is in ignored and those at the Ignore
// severity, as hadolint's fixSeverity does. The override comes first, so a
// rule that is off (DL3057) is reported once it is given a severity.
func Settle(
	failures []rule.CheckFailure, overrides map[rule.Code]rule.Severity, ignored map[rule.Code]bool,
) []rule.CheckFailure {
	settled := []rule.CheckFailure{}

	for _, failure := range failures {
		if severity, overridden := overrides[failure.Code]; overridden {
			failure.Severity = severity
		}

		if !ignored[failure.Code] && failure.Severity != rule.Ignore {
			settled = append(settled, failure)
		}
	}

	return settled
}

// filterIgnored removes failures that are suppressed by ignore pragmas.
func filterIgnored(failures []rule.CheckFailure, directives pragma.Directives) []rule.CheckFailure {
	filtered := []rule.CheckFailure{}

	for _, failure := range failures {
		if !directives.ShouldIgnore(failure) {
			filtered = append(filtered, failure)
		}
	}

	return filtered
}
