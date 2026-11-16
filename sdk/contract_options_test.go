package sdk_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/forkcloser/godolint/sdk"
)

// at is the violation of code in r, and whether there is one.
func at(r *sdk.Result, code string) (sdk.Violation, bool) {
	for _, v := range r.Violations {
		if v.Code == code {
			return v, true
		}
	}

	return sdk.Violation{}, false
}

// TestContractLabelSchema: WithLabelSchema names the labels an image must
// carry (DL3049) and how each value is written (DL3051 to DL3056, DL3058);
// without it no label is asked for; the map is read when the option is made;
// given again it adds; a type that is none of the constants is a
// *ConfigError naming the option and the value, and nothing is linted.
func TestContractLabelSchema(t *testing.T) {
	t.Parallel()

	const bare = "FROM debian:12\n"

	if got := codes(lint(t, bare)); slices.Contains(got, "DL3049") {
		t.Fatalf("no schema, and a label is missing: %v", got)
	}

	schema := map[string]sdk.LabelType{"maintainer": sdk.LabelText}
	option := sdk.WithLabelSchema(schema)
	schema["later"] = sdk.LabelText

	missing, ok := at(lint(t, bare, option), "DL3049")
	if !ok || missing.Line != 1 || missing.Message != "Label `maintainer` is missing." {
		t.Errorf("a required label is missing: %+v, %v", missing, ok)
	}

	if r := lint(t, bare, option); len(r.Violations) != 1 {
		t.Errorf("the map changed after the option was made, and it shows: %v", r.Violations)
	}

	twice := lint(t, bare, option, sdk.WithLabelSchema(map[string]sdk.LabelType{"version": sdk.LabelSemVer}))
	if len(twice.Violations) != 2 {
		t.Errorf("two schemas: %v; want both labels missing", twice.Violations)
	}

	for _, tc := range []struct {
		labelType sdk.LabelType
		good, bad string
		code      string
	}{
		{sdk.LabelText, "anything", "", "DL3051"},
		{sdk.LabelURL, "https://example.com", "not a url", "DL3052"},
		{sdk.LabelRFC3339, "2026-11-15T10:00:00Z", "yesterday", "DL3053"},
		{sdk.LabelSPDX, "Apache-2.0", "my own license", "DL3054"},
		{sdk.LabelHash, "1d901b9c0b0d0445d19b1a9128d89fa55bb4604e", "not a hash", "DL3055"},
		{sdk.LabelSemVer, "1.2.3", "one", "DL3056"},
		{sdk.LabelEmail, "someone@example.com", "someone", "DL3058"},
	} {
		typed := sdk.WithLabelSchema(map[string]sdk.LabelType{"l": tc.labelType})

		if r := lint(t, "FROM debian:12\nLABEL l=\""+tc.good+"\"\n", typed); len(r.Violations) != 0 {
			t.Errorf("%s %q: %v; want nothing", tc.labelType, tc.good, r.Violations)
		}

		if got := codes(
			lint(t, "FROM debian:12\nLABEL l=\""+tc.bad+"\"\n", typed),
		); !slices.Equal(
			got,
			[]string{tc.code},
		) {
			t.Errorf("%s %q: %v; want %s alone", tc.labelType, tc.bad, got, tc.code)
		}
	}

	linter := sdk.New(sdk.WithLabelSchema(map[string]sdk.LabelType{"l": "markdown"}))

	var refused *sdk.ConfigError
	if r, err := linter.Lint(t.Context(), []byte(bare)); !errors.As(err, &refused) || r != nil {
		t.Fatalf("an unknown label type: %v, %v; want a *ConfigError and no result", r, err)
	}

	if refused.Option != "WithLabelSchema" || refused.Value != "markdown" ||
		!strings.Contains(refused.Error(), "markdown") {
		t.Errorf("the error names %q and %q: %v", refused.Option, refused.Value, refused)
	}
}

// TestContractStrictLabels: a label the schema does not name is a violation
// (DL3050) with WithStrictLabels, and none without it.
func TestContractStrictLabels(t *testing.T) {
	t.Parallel()

	const dockerfile = "FROM debian:12\nLABEL maintainer=\"x\" other=\"y\"\n"

	schema := sdk.WithLabelSchema(map[string]sdk.LabelType{"maintainer": sdk.LabelText})

	if r := lint(t, dockerfile, schema); len(r.Violations) != 0 {
		t.Errorf("not strict: %v; want nothing", r.Violations)
	}

	extra, ok := at(lint(t, dockerfile, schema, sdk.WithStrictLabels()), "DL3050")
	if !ok || extra.Line != 2 {
		t.Errorf("strict: %+v, %v; want DL3050 at line 2", extra, ok)
	}

	// With no schema, every label is one the schema does not name.
	if got := codes(
		lint(t, "FROM debian:12\nLABEL a=\"b\"\n", sdk.WithStrictLabels()),
	); !slices.Equal(
		got,
		[]string{"DL3050"},
	) {
		t.Errorf("strict and no schema: %v; want DL3050 alone", got)
	}
}

// TestContractTrustedRegistries: with WithTrustedRegistries, DL3026 reports
// the image of a FROM, of a COPY --from and of a bind or cache mount that
// comes from another registry; an image with no registry is Docker Hub's;
// a `*` stands for the start or the end of a name; given again the option
// adds; without it every registry is trusted.
func TestContractTrustedRegistries(t *testing.T) {
	t.Parallel()

	const dockerfile = "FROM mine.example.com/base:1 AS base\n" +
		"FROM other.io/tools:1\n" +
		"COPY --from=other.io/tools:1 /a /b\n" +
		"RUN --mount=type=cache,from=other.io/tools:1,target=/c true\n" +
		"COPY --from=base /a /b\n"

	lines := func(opts ...sdk.Option) []int {
		var out []int

		for _, v := range lint(t, dockerfile, opts...).Violations {
			if v.Code == "DL3026" {
				out = append(out, v.Line)
			}
		}

		return out
	}

	for _, tc := range []struct {
		name string
		opts []sdk.Option
		want []int
	}{
		{"none named", nil, nil},
		{"one named", []sdk.Option{sdk.WithTrustedRegistries("mine.example.com")}, []int{2, 3, 4}},
		{"the other named", []sdk.Option{sdk.WithTrustedRegistries("other.io")}, []int{1}},
		{"both, in one option", []sdk.Option{sdk.WithTrustedRegistries("mine.example.com", "other.io")}, nil},
		{
			"both, in two options",
			[]sdk.Option{sdk.WithTrustedRegistries("mine.example.com"), sdk.WithTrustedRegistries("other.io")},
			nil,
		},
		{"the end of a name", []sdk.Option{sdk.WithTrustedRegistries("*.example.com")}, []int{2, 3, 4}},
		{"the start of a name", []sdk.Option{sdk.WithTrustedRegistries("other.*")}, []int{1}},
		{"any", []sdk.Option{sdk.WithTrustedRegistries("*")}, nil},
		{"Docker Hub alone", []sdk.Option{sdk.WithTrustedRegistries("docker.io")}, []int{1, 2, 3, 4}},
	} {
		if got := lines(tc.opts...); !slices.Equal(got, tc.want) {
			t.Errorf("%s: DL3026 at lines %v, want %v", tc.name, got, tc.want)
		}
	}

	hub := lint(t, "FROM debian:12\n", sdk.WithTrustedRegistries("hub.docker.com"))
	if _, reported := at(hub, "DL3026"); reported {
		t.Errorf("an image with no registry, hub.docker.com trusted: %v", hub.Violations)
	}
}

// TestContractOptionsReachTheRulesChosen: the options configure the rules of
// the linter whichever option chose them and in whichever order, and one
// linter's configuration is not another's.
func TestContractOptionsReachTheRulesChosen(t *testing.T) {
	t.Parallel()

	const dockerfile = "FROM other.io/tools:1\n"

	for name, opts := range map[string][]sdk.Option{
		"the registries alone": {sdk.WithTrustedRegistries("docker.io")},
		"the registries, then a rule set": {
			sdk.WithTrustedRegistries("docker.io"), sdk.WithRuleSet(sdk.RuleSetRecommended),
		},
		"a rule set, then the registries": {
			sdk.WithRuleSet(sdk.RuleSetRecommended), sdk.WithTrustedRegistries("docker.io"),
		},
	} {
		if _, reported := at(lint(t, dockerfile, opts...), "DL3026"); !reported {
			t.Errorf("%s: DL3026 not reported", name)
		}
	}

	if _, reported := at(lint(t, dockerfile, sdk.WithRuleSet(sdk.RuleSetRecommended)), "DL3026"); reported {
		t.Error("a linter given no registries carries those of another")
	}
}

// TestContractDisableIgnorePragma: with WithDisableIgnorePragma the linter
// reads none of the three ignore pragmas, and what they silence is reported.
func TestContractDisableIgnorePragma(t *testing.T) {
	t.Parallel()

	for dockerfile, code := range map[string]string{
		"# hadolint ignore=DL3006\nFROM ubuntu\n":                       "DL3006",
		"# hadolint stage ignore=DL3002\nFROM alpine:3.7\nUSER root\n":  "DL3002",
		"# hadolint global ignore=DL3002\nFROM alpine:3.7\nUSER root\n": "DL3002",
	} {
		if got := codes(lint(t, dockerfile)); slices.Contains(got, code) {
			t.Errorf("%q: the pragma silences nothing: %v", dockerfile, got)
		}

		if got := codes(lint(t, dockerfile, sdk.WithDisableIgnorePragma())); !slices.Contains(got, code) {
			t.Errorf("%q, pragmas not read: %v; want %s", dockerfile, got, code)
		}
	}
}

// TestContractSeverity: WithSeverity gives the violations of a code another
// severity; a code given several gets the gravest, in whatever order; a
// rule that is off is reported once given one; a severity that is none of
// the four is a *ConfigError.
func TestContractSeverity(t *testing.T) {
	t.Parallel()

	const untagged = "FROM ubuntu\n"

	severity := func(code string, opts ...sdk.Option) sdk.Severity {
		v, _ := at(lint(t, untagged, opts...), code)

		return v.Severity
	}

	if got := severity("DL3006"); got != sdk.SeverityWarning {
		t.Fatalf("the premise: DL3006 is a warning, got %q", got)
	}

	for _, given := range []sdk.Severity{sdk.SeverityError, sdk.SeverityWarning, sdk.SeverityInfo, sdk.SeverityStyle} {
		if got := severity("DL3006", sdk.WithSeverity(given, "DL3000", "DL3006")); got != given {
			t.Errorf("DL3006 given %q is reported as %q", given, got)
		}
	}

	for name, opts := range map[string][]sdk.Option{
		"the gravest last":  {sdk.WithSeverity(sdk.SeverityStyle, "DL3006"), sdk.WithSeverity(sdk.SeverityError, "DL3006")},
		"the gravest first": {sdk.WithSeverity(sdk.SeverityError, "DL3006"), sdk.WithSeverity(sdk.SeverityInfo, "DL3006")},
	} {
		if got := severity("DL3006", opts...); got != sdk.SeverityError {
			t.Errorf("%s: DL3006 is reported as %q, want the gravest", name, got)
		}
	}

	// No HEALTHCHECK, which DL3057 is about, and DL3057 is off.
	if got := severity("DL3057", sdk.WithSeverity(sdk.SeverityInfo, "DL3057")); got != sdk.SeverityInfo {
		t.Errorf("DL3057 given a severity is reported as %q, want info", got)
	}

	if other := severity("DL3006", sdk.WithSeverity(sdk.SeverityError, "DL3007")); other != sdk.SeverityWarning {
		t.Errorf("another code's severity changed DL3006's: %q", other)
	}

	for _, refusedValue := range []sdk.Severity{sdk.SeverityIgnore, "fatal", ""} {
		var refused *sdk.ConfigError

		_, err := sdk.New(sdk.WithSeverity(refusedValue, "DL3006")).Lint(t.Context(), []byte(untagged))
		if !errors.As(err, &refused) || refused.Option != "WithSeverity" || refused.Value != string(refusedValue) {
			t.Errorf("WithSeverity(%q): %v; want a *ConfigError naming it", refusedValue, err)
		}
	}
}

// TestContractFirstRefusalIsReturned: of several options given a value they
// do not take, Lint returns the first's error, in the order the options were
// given.
func TestContractFirstRefusalIsReturned(t *testing.T) {
	t.Parallel()

	badSeverity := sdk.WithSeverity("loud", "DL3006")
	badType := sdk.WithLabelSchema(map[string]sdk.LabelType{"l": "markdown"})

	for name, tc := range map[string]struct {
		opts   []sdk.Option
		option string
		value  string
	}{
		"the severity first":   {[]sdk.Option{badSeverity, badType}, "WithSeverity", "loud"},
		"the label type first": {[]sdk.Option{badType, badSeverity}, "WithLabelSchema", "markdown"},
	} {
		var refused *sdk.ConfigError

		_, err := sdk.New(tc.opts...).Lint(t.Context(), []byte("FROM debian:12\n"))
		if !errors.As(err, &refused) || refused.Option != tc.option || refused.Value != tc.value {
			t.Errorf("%s: %v; want %s refusing %q", name, err, tc.option, tc.value)
		}
	}
}

// TestContractDisabledRulesHoldWhereverTheyStand: a code given to
// WithDisabledRules reaches no Result, whether the option comes before or
// after the one that chooses the rules, and whatever severity it is given.
func TestContractDisabledRulesHoldWhereverTheyStand(t *testing.T) {
	t.Parallel()

	for name, opts := range map[string][]sdk.Option{
		"before the rule set": {sdk.WithDisabledRules("DL3006"), sdk.WithRuleSet(sdk.RuleSetAll)},
		"after the rule set":  {sdk.WithRuleSet(sdk.RuleSetAll), sdk.WithDisabledRules("DL3006")},
		"given a severity":    {sdk.WithSeverity(sdk.SeverityError, "DL3006"), sdk.WithDisabledRules("DL3006")},
	} {
		if got := codes(lint(t, "FROM ubuntu\nMAINTAINER x\n", opts...)); !slices.Equal(got, []string{"DL4000"}) {
			t.Errorf("%s: %v; want DL4000 alone", name, got)
		}
	}
}

// TestContractShellcheckCodes: a code shellcheck reports is disabled and
// given a severity as a rule's is; shellcheck leaves out SC2187, SC1090 and
// SC1091 unless WithShellcheckExclusions names other checks, or none; it
// reads the configuration file WithShellcheckRCFile names.
func TestContractShellcheckCodes(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("shellcheck"); err != nil {
		t.Skip("no shellcheck on PATH")
	}

	// SC2086: an unquoted variable. SC1091: a sourced file not followed.
	const dockerfile = "FROM debian:12\nRUN . ./env.sh && echo $HOME\n"

	plain := lint(t, dockerfile, sdk.WithShellcheck())
	if got := codes(plain); !slices.Contains(got, "SC2086") || slices.Contains(got, "SC1091") {
		t.Fatalf("the premise: SC2086 reported and SC1091 left out, got %v", got)
	}

	if got := codes(
		lint(t, dockerfile, sdk.WithShellcheck(), sdk.WithDisabledRules("SC2086")),
	); slices.Contains(
		got,
		"SC2086",
	) {
		t.Errorf("SC2086 disabled: %v", got)
	}

	raised, _ := at(lint(t, dockerfile, sdk.WithShellcheck(), sdk.WithSeverity(sdk.SeverityError, "SC2086")), "SC2086")
	if raised.Severity != sdk.SeverityError {
		t.Errorf("SC2086 given error is reported as %q", raised.Severity)
	}

	none := codes(lint(t, dockerfile, sdk.WithShellcheck(sdk.WithShellcheckExclusions())))
	if !slices.Contains(none, "SC1091") || !slices.Contains(none, "SC2086") {
		t.Errorf("nothing left out: %v; want SC1091 and SC2086", none)
	}

	other := codes(lint(t, dockerfile, sdk.WithShellcheck(sdk.WithShellcheckExclusions("SC2086"))))
	if !slices.Contains(other, "SC1091") || slices.Contains(other, "SC2086") {
		t.Errorf("SC2086 left out in place of the three: %v; want SC1091 and no SC2086", other)
	}

	rcfile := filepath.Join(t.TempDir(), "shellcheckrc")
	if err := os.WriteFile(rcfile, []byte("disable=SC2086\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	configured := codes(lint(t, dockerfile, sdk.WithShellcheck(sdk.WithShellcheckRCFile(rcfile))))
	if slices.Contains(configured, "SC2086") {
		t.Errorf("SC2086 disabled in the configuration file named: %v", configured)
	}
}
