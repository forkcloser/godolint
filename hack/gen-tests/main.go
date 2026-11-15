// gen-tests extracts tests from hadolint Haskell test files and generates Go tests.
package main

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/template"
)

// ErrTestGeneration is the base error for test generation failures.
var ErrTestGeneration = errors.New("test generation error")

// TestCase represents a single test case extracted from hadolint.
type TestCase struct {
	Name       string
	RuleCode   string
	Dockerfile string
	ShouldFail bool // true if ruleCatches, false if ruleCatchesNot

	// OnBuild is a case every RUN of which is wrapped in ONBUILD.
	OnBuild bool
	// Exactly is a case that counts the failures: Count of them.
	Exactly bool
	Count   int
	// OnLines is a case every failure of which must be on a line.
	OnLines bool
	// Config is the configuration in force for the case, nil for none.
	Config *HadolintConfig
}

// RuleTests contains all test cases for a rule.
type RuleTests struct {
	RuleCode  string
	TestCases []TestCase
	Config    *HadolintConfig
}

// HadolintConfig represents hadolint configuration for test cases.
type HadolintConfig struct {
	LabelSchema       map[string]string // label name -> type (RawText, Email, etc.)
	StrictLabels      bool
	AllowedRegistries []string
}

// failing names the cases of hadolint that godolint does not pass, by rule
// and by name, each with the issue that says why: a case is ported, or it
// is here. An entry goes when its issue is fixed, and the generator stops on
// an entry that names no case, so none outlives its reason. It is empty when
// godolint passes every case the generator can write.
//
//nolint:gochecknoglobals // a table, written to by its test alone
var failing = map[string]map[string]string{}

// ruleCodePattern is a rule's code.
var ruleCodePattern = regexp.MustCompile(`^DL\d{4}$`)

// misfiled names the codes hadolint's specs assert that are no rule's, each
// with what it is. A code that is no rule's and is not here stops the
// generator.
//
//nolint:gochecknoglobals // a constant table
var misfiled = map[string]string{
	"FROM scratch\nDL3045": "hadolint's own slip (DL3045Spec.hs: the Dockerfile's first line is in the code), " +
		"where the assertion holds whatever the rule does, no rule having that code",
}

// notPorted is the reason a case is left out, when there is one. Each is
// counted and named in the report: nothing is left out without a word.
func notPorted(testCase TestCase) string {
	if testCase.Config != nil && len(testCase.Config.AllowedRegistries) > 0 {
		return "sets allowedRegistries, which DL3026 takes once it has a WithConfig constructor " +
			"(https://github.com/forkcloser/godolint/issues/125)"
	}

	if issue, known := failing[testCase.RuleCode][testCase.Name]; known {
		return "godolint does not pass it (" + issue + ")"
	}

	return ""
}

func main() {
	if len(os.Args) != 2 {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s <hadolint-test-dir>\n", os.Args[0])

		os.Exit(1)
	}

	if err := generate(os.Args[1], os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)

		os.Exit(1)
	}
}

// generate reads every DL spec under dir and writes the test file of every
// implemented rule into the working directory. A spec it cannot read whole
// stops it: a case is ported, or named with the reason it is not.
func generate(dir string, report io.Writer) error {
	files, err := filepath.Glob(filepath.Join(dir, "DL*Spec.hs"))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTestGeneration, err)
	}

	if len(files) == 0 {
		return fmt.Errorf("%w: no spec in %s", ErrTestGeneration, dir)
	}

	read := make(map[string][]TestCase)
	assertions := 0
	allTests := make(map[string][]TestCase)
	left := map[string]int{}
	totalTests := 0

	for _, file := range files {
		// #nosec G304 G703 -- the specs of the pinned hadolint, named by the caller
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrTestGeneration, err)
		}

		cases, err := readSpec(string(content))
		if err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(file), err)
		}

		for _, testCase := range cases {
			read[testCase.RuleCode] = append(read[testCase.RuleCode], testCase)
		}

		assertions += len(cases)
	}

	for _, ruleCode := range slices.Sorted(maps.Keys(read)) {
		nameCases(read[ruleCode])

		// A case is filed under the code its assertion gives. One that is
		// no rule's, or a rule's with no implementation, has no test file
		// to go to: it is named, or the generator stops.
		if reason, known := misfiled[ruleCode]; known {
			left[strconv.Quote(ruleCode)+": "+reason] += len(read[ruleCode])

			continue
		}

		if !ruleCodePattern.MatchString(ruleCode) {
			return fmt.Errorf("%w: a case asserts the code %q, which is no rule's", ErrTestGeneration, ruleCode)
		}

		if _, err := os.Stat(strings.ToLower(ruleCode) + "_impl.go"); err != nil {
			left[ruleCode+": the rule has no implementation"] += len(read[ruleCode])

			continue
		}

		named := map[string]bool{}

		for _, testCase := range read[ruleCode] {
			named[testCase.Name] = true

			if reason := notPorted(testCase); reason != "" {
				left[ruleCode+": "+reason]++

				continue
			}

			allTests[ruleCode] = append(allTests[ruleCode], testCase)
			totalTests++
		}

		for name := range failing[ruleCode] {
			if !named[name] {
				return fmt.Errorf(
					"%w: %s has no case %q, which is listed as failing",
					ErrTestGeneration,
					ruleCode,
					name,
				)
			}
		}
	}

	fmt.Fprint(report, "# Hadolint Test Generation\n\n")
	fmt.Fprintf(report, "Total test files parsed: %d\n", len(files))
	fmt.Fprintf(report, "Total assertions read: %d\n", assertions)
	fmt.Fprintf(report, "Total test cases extracted: %d\n", totalTests)
	fmt.Fprintf(report, "Rules with tests: %d\n\n", len(allTests))

	for _, reason := range slices.Sorted(maps.Keys(left)) {
		fmt.Fprintf(report, "Not ported: %d case(s) of %s\n", left[reason], reason)
	}

	generated := 0

	for _, ruleCode := range slices.Sorted(maps.Keys(allTests)) {
		if err := generateTestFile(ruleCode, allTests[ruleCode]); err != nil {
			return fmt.Errorf("%s: %w", ruleCode, err)
		}

		fmt.Fprintf(report, "Generated tests for %s (%d cases)\n", ruleCode, len(allTests[ruleCode]))

		generated++
	}

	fmt.Fprintf(report, "\nGenerated test files: %d\n", generated)

	// Every assertion read is in a test file or named above: the sum says
	// none fell between.
	named := 0
	for _, count := range left {
		named += count
	}

	if totalTests+named != assertions {
		return fmt.Errorf(
			"%w: %d assertions read, %d generated and %d named as not ported",
			ErrTestGeneration, assertions, totalTests, named,
		)
	}

	return nil
}

func escapeBackticks(s string) string {
	return strings.ReplaceAll(s, "`", "` + \"`\" + `")
}

// goLabelType maps Haskell LabelType names to Go LabelType names.
func goLabelType(haskellType string) string {
	// Map Haskell type names to Go type names (handle acronym casing differences)
	typeMap := map[string]string{
		"RawText": "RawText",
		"Email":   "Email",
		"Url":     "URL",     // Haskell: Url, Go: URL
		"Rfc3339": "RFC3339", // Haskell: Rfc3339, Go: RFC3339
		"Spdx":    "SPDX",    // Haskell: Spdx, Go: SPDX
		"GitHash": "GitHash",
		"SemVer":  "SemVer",
	}
	if goType, ok := typeMap[haskellType]; ok {
		return goType
	}

	return haskellType // fallback
}

// nameCases gives each case of a rule its own name: the `it` it comes from, with
// what tells it from the block's other cases.
func nameCases(cases []TestCase) {
	seen := map[string]int{}

	for i := range cases {
		name := cases[i].Name
		if cases[i].OnBuild {
			name += " (ONBUILD)"
		}

		seen[name]++
		if count := seen[name]; count > 1 {
			name += fmt.Sprintf(" [%d]", count)
		}

		cases[i].Name = name
	}
}

// sameConfig reports whether every case runs under one configuration: the
// spec's own, or none.
func sameConfig(cases []TestCase) bool {
	for _, testCase := range cases {
		if testCase.Config != cases[0].Config {
			return false
		}
	}

	return true
}

func generateTestFile(ruleCode string, cases []TestCase) error {
	if len(cases) == 0 {
		return fmt.Errorf("%w: no case to write", ErrTestGeneration)
	}

	if !sameConfig(cases) {
		return fmt.Errorf(
			"%w: cases under different configurations, which the template does not write",
			ErrTestGeneration,
		)
	}

	// Sort test cases by name for consistent output
	slices.SortStableFunc(cases, func(left, right TestCase) int {
		return strings.Compare(left.Name, right.Name)
	})

	tmpl := `package rules_test

import (
	"testing"

{{if .Config}}	"github.com/forkcloser/godolint/internal/config"
{{end}}	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// Auto-generated tests for {{.RuleCode}} ported from hadolint test suite.
// Source: hadolint/test/Hadolint/Rule/{{.RuleCode}}Spec.hs
//
// To regenerate: go generate ./internal/rules

func Test{{.RuleCode}}(t *testing.T) {
	t.Parallel()

{{if .Config}}	cfg := &config.Config{
{{if .Config.LabelSchema}}		LabelSchema: map[string]config.LabelType{
{{range $key, $val := .Config.LabelSchema}}			"{{$key}}": config.LabelType{{goLabelType $val}},
{{end}}		},
{{end}}{{if .Config.StrictLabels}}		StrictLabels: true,
{{end}}	}
	allRules := []rule.Rule{
		rules.{{.RuleCode}}WithConfig(cfg),
	}
{{else}}	allRules := []rule.Rule{
		rules.{{.RuleCode}}(),
	}
{{end}}{{range .TestCases}}
	t.Run(
		{{.Name | printf "%q"}},
		func(t *testing.T) {
			t.Parallel()

			dockerfile := ` + "`" + `{{.Dockerfile | escapeBackticks}}` + "`" + `
			violations := testutils.{{if .OnBuild}}LintDockerfileOnBuild{{else}}LintDockerfile{{end}}(dockerfile, allRules)
{{if .ShouldFail}}
			testutils.AssertContainsViolation(t, violations, "{{.RuleCode}}"){{else if and .Exactly (gt .Count 0)}}
			testutils.AssertViolationCount(t, violations, "{{.RuleCode}}", {{.Count}}){{if .OnLines}}
			testutils.AssertOnLines(t, violations){{end}}{{else}}
			testutils.AssertNoViolation(t, violations, "{{.RuleCode}}"){{end}}
		},
	)
{{end}}}
`

	funcMap := template.FuncMap{
		"escapeBackticks": escapeBackticks,
		"goLabelType":     goLabelType,
	}

	tpl, err := template.New("test").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTestGeneration, err)
	}

	filename := strings.ToLower(ruleCode) + "_test.go"

	// #nosec G304
	outputFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTestGeneration, err)
	}

	defer func() {
		_ = outputFile.Close()
	}()

	data := RuleTests{
		RuleCode:  ruleCode,
		TestCases: cases,
		Config:    cases[0].Config,
	}

	err = tpl.Execute(outputFile, data)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTestGeneration, err)
	}

	return nil
}
