package process_test

import (
	"strings"
	"testing"

	"github.com/forkcloser/godolint/internal/parser"
	"github.com/forkcloser/godolint/internal/process"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// seen is a rule that reports, once every instruction is read, the names of
// the instructions it was shown, in one failure: what a state of the rule
// saw.
type seen struct {
	code rule.Code
}

func (r seen) Code() rule.Code        { return r.code }
func (seen) Severity() rule.Severity  { return rule.Info }
func (seen) Message() string          { return "" }
func (seen) InitialState() rule.State { return rule.EmptyState("") }

func (seen) Check(_ int, state rule.State, instruction syntax.Instruction) rule.State {
	return state.ReplaceData(strings.TrimSpace(rule.Data[string](state) + " " + instruction.Name()))
}

func (r seen) Finalize(state rule.State) rule.State {
	return state.AddFailure(rule.CheckFailure{Code: r.code, Message: rule.Data[string](state), Line: 1})
}

// saysSo is a rule that says of itself that it reads what ONBUILD wraps.
type saysSo struct{ seen }

func (saysSo) InOnBuild() bool { return true }

// TestRunAppliesARuleInsideOnBuildAsHadolintComposesIt: a rule hadolint
// composes with its onbuild combinator (DL3009 is one, and a rule may say
// so of itself) has two states: one shown every instruction as written, an
// ONBUILD as an ONBUILD, and one shown the instructions ONBUILD wraps and no
// other. Any other rule (DL3000 is one) has the first alone.
func TestRunAppliesARuleInsideOnBuildAsHadolintComposesIt(t *testing.T) {
	t.Parallel()

	instructions, err := parser.NewBuildkitParser().Parse(
		[]byte("FROM a\nONBUILD RUN b\nRUN c\nONBUILD COPY d e\nUSER f\n"),
	)
	if err != nil {
		t.Fatal(err)
	}

	const (
		written = "FROM ONBUILD RUN ONBUILD USER"
		wrapped = "RUN COPY"
	)

	for name, tc := range map[string]struct {
		applied rule.Rule
		want    []string
	}{
		"a rule hadolint applies inside ONBUILD": {seen{code: "DL3009"}, []string{written, wrapped}},
		"a rule hadolint does not":               {seen{code: "DL3000"}, []string{written}},
		"a rule that says so of itself":          {saysSo{seen{code: "SHELLCHECK"}}, []string{written, wrapped}},
		"a rule of no known code":                {seen{code: "DL9999"}, []string{written}},
	} {
		var got []string
		for _, failure := range process.NewProcessor([]rule.Rule{tc.applied}).Run(instructions) {
			got = append(got, failure.Message)
		}

		if strings.Join(got, " | ") != strings.Join(tc.want, " | ") {
			t.Errorf("%s: its states saw %q, want %q", name, got, tc.want)
		}
	}
}

// counted is a rule that counts the instructions it is shown, and reports
// one failure under each of its codes once they are all read.
type counted struct {
	code    rule.Code
	reports []rule.Code
	shown   *int
}

func (r counted) Code() rule.Code        { return r.code }
func (counted) Severity() rule.Severity  { return rule.Info }
func (counted) Message() string          { return "" }
func (counted) InitialState() rule.State { return rule.EmptyState("") }

func (r counted) Check(_ int, state rule.State, _ syntax.Instruction) rule.State {
	*r.shown++

	return state
}

func (r counted) Finalize(state rule.State) rule.State {
	for _, code := range r.reports {
		state = state.AddFailure(rule.CheckFailure{Code: code, Line: 1})
	}

	return state
}

// TestRunDoesNotRunAnIgnoredRule: a rule whose code is ignored is shown no
// instruction, so it costs nothing and does nothing; a failure under an
// ignored code is dropped whichever rule reports it, and that rule, whose
// own code is not ignored, still runs.
func TestRunDoesNotRunAnIgnoredRule(t *testing.T) {
	t.Parallel()

	instructions, err := parser.NewBuildkitParser().Parse([]byte("FROM a\nRUN b\nRUN c\n"))
	if err != nil {
		t.Fatal(err)
	}

	for name, tc := range map[string]struct {
		ignored      []rule.Code
		off, on, sub int
		want         string
	}{
		"nothing ignored":             {nil, 3, 3, 3, "DL0001 DL0002 SC1 SC2"},
		"a rule's code":               {[]rule.Code{"DL0001"}, 0, 3, 3, "DL0002 SC1 SC2"},
		"a code another rule reports": {[]rule.Code{"SC1"}, 3, 3, 3, "DL0001 DL0002 SC2"},
		"both":                        {[]rule.Code{"DL0001", "SC1", "SC2"}, 0, 3, 3, "DL0002"},
		"the reporting rule's code":   {[]rule.Code{"SHELLCHECK"}, 3, 3, 0, "DL0001 DL0002"},
	} {
		var off, on, sub int

		ignored := map[rule.Code]bool{}
		for _, code := range tc.ignored {
			ignored[code] = true
		}

		var got []string
		for _, failure := range process.NewProcessor([]rule.Rule{
			counted{code: "DL0001", reports: []rule.Code{"DL0001"}, shown: &off},
			counted{code: "DL0002", reports: []rule.Code{"DL0002"}, shown: &on},
			counted{code: "SHELLCHECK", reports: []rule.Code{"SC1", "SC2"}, shown: &sub},
		}).WithIgnored(ignored).Run(instructions) {
			got = append(got, string(failure.Code))
		}

		if strings.Join(got, " ") != tc.want {
			t.Errorf("%s: reported %q, want %q", name, got, tc.want)
		}

		if off != tc.off || on != tc.on || sub != tc.sub {
			t.Errorf("%s: the rules were shown %d, %d and %d instructions, want %d, %d and %d",
				name, off, on, sub, tc.off, tc.on, tc.sub)
		}
	}
}
