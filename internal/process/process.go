// Package process orchestrates running hadolint rules over a parsed Dockerfile AST.
// Ported from Hadolint/Process.hs
package process

import (
	"cmp"
	"slices"

	"github.com/forkcloser/godolint/internal/pragma"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/syntax"
)

// Processor runs rules against a Dockerfile AST and collects violations.
type Processor struct {
	rules                []rule.Rule
	disableIgnorePragmas bool
	ignored              map[rule.Code]bool
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

// WithIgnored turns the codes off. A rule whose code is among them is not
// run at all: it costs nothing, and what it would do on a Dockerfile is not
// done. A failure with one of the codes is dropped, whichever rule reports
// it: the shellcheck rule reports under shellcheck's codes, not its own.
func (p *Processor) WithIgnored(codes map[rule.Code]bool) *Processor {
	p.ignored = codes

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
		if p.ignored[currentRule.Code()] {
			continue
		}

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

	return slices.DeleteFunc(allFailures, func(failure rule.CheckFailure) bool { return p.ignored[failure.Code] })
}

// Settle gives the failures of a run the severity overrides names for their
// code, then drops those at the Ignore severity, as hadolint's fixSeverity
// does. The override comes first, so a rule that is off (DL3057) is reported
// once it is given a severity.
//
// What it returns is ordered by line, as hadolint orders what it prints, and
// by code on one line, so that a run gives the same order every time: the
// rules do not all report in the order of their lines.
func Settle(failures []rule.CheckFailure, overrides map[rule.Code]rule.Severity) []rule.CheckFailure {
	settled := []rule.CheckFailure{}

	for _, failure := range failures {
		if severity, overridden := overrides[failure.Code]; overridden {
			failure.Severity = severity
		}

		if failure.Severity != rule.Ignore {
			settled = append(settled, failure)
		}
	}

	slices.SortStableFunc(settled, func(left, right rule.CheckFailure) int {
		return cmp.Or(cmp.Compare(left.Line, right.Line), cmp.Compare(left.Code, right.Code))
	})

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
