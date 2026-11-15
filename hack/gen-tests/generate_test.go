package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// workspace is a directory of specs and, as the working directory, the place
// the test files go, with an implementation for each rule named.
func workspace(t *testing.T, specs map[string]string, implemented ...string) string {
	t.Helper()

	specDir := t.TempDir()
	for name, body := range specs {
		if err := os.WriteFile(filepath.Join(specDir, name), []byte(spec(body)), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	t.Chdir(t.TempDir())

	for _, code := range implemented {
		if err := os.WriteFile(strings.ToLower(code)+"_impl.go", []byte("package rules\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	return specDir
}

// TestGenerateAccountsForEveryAssertion: every assertion of the specs is in
// a test file or named in the report with its reason: a case of a rule with
// no implementation, and one whose code is hadolint's known slip. The
// report's counts add up.
//
//nolint:paralleltest // workspace changes the working directory
func TestGenerateAccountsForEveryAssertion(t *testing.T) {
	specDir := workspace(t, map[string]string{
		"DL3000Spec.hs": "    it \"relative\" $ ruleCatches \"DL3000\" \"WORKDIR a\"\n" +
			"    it \"absolute\" $ do\n      ruleCatchesNot \"DL3000\" \"WORKDIR /a\"\n" +
			"      onBuildRuleCatchesNot \"FROM scratch\\nDL3045\" \"WORKDIR /a\"\n",
		"DL3099Spec.hs": "    it \"not written yet\" $ ruleCatches \"DL3099\" \"RUN a\"\n" +
			"    it \"nor this\" $ ruleCatchesNot \"DL3099\" \"RUN b\"\n",
	}, "DL3000")

	var report strings.Builder
	if err := generate(specDir, &report); err != nil {
		t.Fatalf("generate: %v\n%s", err, report.String())
	}

	for _, line := range []string{
		"Total assertions read: 5\n",
		"Total test cases extracted: 2\n",
		"Not ported: 2 case(s) of DL3099: the rule has no implementation\n",
		"Not ported: 1 case(s) of \"FROM scratch\\nDL3045\": hadolint's own slip",
		"Generated test files: 1\n",
	} {
		if !strings.Contains(report.String(), line) {
			t.Errorf("the report lacks %q:\n%s", line, report.String())
		}
	}

	written, err := os.ReadFile("dl3000_test.go")
	if err != nil {
		t.Fatal(err)
	}

	if got := strings.Count(string(written), "t.Run("); got != 2 {
		t.Errorf("dl3000_test.go holds %d cases, want 2:\n%s", got, written)
	}

	if _, err := os.Stat("dl3099_test.go"); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a test file was written for a rule with no implementation: %v", err)
	}
}

// TestGenerateStopsOnACodeThatIsNoRules: a case whose code is no rule's, and
// is not one of hadolint's known slips, stops the generator, naming the
// code: it has no test file to go to and is not left out quietly.
//
//nolint:paralleltest // workspace changes the working directory
func TestGenerateStopsOnACodeThatIsNoRules(t *testing.T) {
	specDir := workspace(t, map[string]string{
		"DL3000Spec.hs": "    it \"relative\" $ ruleCatches \"DL3000\" \"WORKDIR a\"\n" +
			"    it \"mistyped\" $ ruleCatches \"DL300\" \"WORKDIR a\"\n",
	}, "DL3000")

	var report strings.Builder

	err := generate(specDir, &report)
	if err == nil || !errors.Is(err, ErrTestGeneration) || !strings.Contains(err.Error(), `"DL300"`) {
		t.Fatalf("generate = %v; want it to stop on the code \"DL300\"", err)
	}
}

// TestGenerateStopsOnAFailingEntryWithNoCase: an entry of the failing table
// that names no case of its rule stops the generator, so none outlives its
// reason.
//
//nolint:paralleltest // workspace changes the working directory
func TestGenerateStopsOnAFailingEntryWithNoCase(t *testing.T) {
	specDir := workspace(t, map[string]string{
		"DL4001Spec.hs": "    it \"another case\" $ ruleCatches \"DL4001\" \"RUN wget a && curl b\"\n",
	}, "DL4001")

	saved, had := failing["DL4001"]
	failing["DL4001"] = map[string]string{"a case that is gone": "https://example.com/issue"}

	t.Cleanup(func() {
		if had {
			failing["DL4001"] = saved
		} else {
			delete(failing, "DL4001")
		}
	})

	var report strings.Builder

	err := generate(specDir, &report)
	if err == nil || !errors.Is(err, ErrTestGeneration) || !strings.Contains(err.Error(), "a case that is gone") {
		t.Fatalf("generate = %v; want it to stop on the entry that names no case", err)
	}
}

// TestGenerateWritesEachCaseItsConfiguration: when the cases of a rule do
// not run under one configuration, each is written with its own: the rule
// built from what its `it` binds, or the plain rule for a case under none.
//
//nolint:paralleltest // workspace changes the working directory
func TestGenerateWritesEachCaseItsConfiguration(t *testing.T) {
	specDir := workspace(t, map[string]string{
		"DL3026Spec.hs": "    it \"under none\" $ ruleCatchesNot \"DL3026\" \"FROM a.com/one\"\n" +
			"    it \"under two registries\" $ do\n" +
			"      let ?config = def { allowedRegistries = [\"x.com\", \"*.random.com\"] }\n" +
			"      ruleCatchesNot \"DL3026\" \"FROM x.com/two\"\n" +
			"    it \"under another\" $ do\n" +
			"      let ?config = def { allowedRegistries = [\"docker.io\"] }\n" +
			"      ruleCatches \"DL3026\" \"FROM x.com/three\"\n",
	}, "DL3026")

	var report strings.Builder
	if err := generate(specDir, &report); err != nil {
		t.Fatalf("generate: %v\n%s", err, report.String())
	}

	if !strings.Contains(report.String(), "Total test cases extracted: 3\n") ||
		strings.Contains(report.String(), "Not ported") {
		t.Errorf("the report does not have the three cases generated and none left:\n%s", report.String())
	}

	written, err := os.ReadFile("dl3026_test.go")
	if err != nil {
		t.Fatal(err)
	}

	// What stands before the first case is shared by all of them: no rule
	// may be built there.
	cases := strings.Split(string(written), "t.Run(")
	if len(cases) != 4 {
		t.Fatalf("dl3026_test.go holds %d cases, want 3:\n%s", len(cases)-1, written)
	}

	if !strings.Contains(cases[0], `"github.com/forkcloser/godolint/internal/config"`) ||
		strings.Contains(cases[0], "allRules") {
		t.Errorf("above the cases: want the config import and no rule:\n%s", cases[0])
	}

	// The cases are written in the order of their names.
	for idx, want := range []struct {
		name, rule, dockerfile string
	}{
		{"under another", `rules.DL3026WithConfig(&config.Config{
					AllowedRegistries: []string{"docker.io"},
				}),`, "FROM x.com/three"},
		{"under none", "rules.DL3026(),", "FROM a.com/one"},
		{"under two registries", `rules.DL3026WithConfig(&config.Config{
					AllowedRegistries: []string{"x.com", "*.random.com"},
				}),`, "FROM x.com/two"},
	} {
		got := cases[idx+1]

		if !strings.Contains(got, `"`+want.name+`"`) || !strings.Contains(got, want.dockerfile) {
			t.Errorf("case %d is not %q on %q:\n%s", idx+1, want.name, want.dockerfile, got)
		}

		if !strings.Contains(got, want.rule) || strings.Count(got, "rules.DL3026") != 1 {
			t.Errorf("%q: want the one rule\n%s\ngot:\n%s", want.name, want.rule, got)
		}
	}
}

// TestGenerateWritesASharedConfigurationOnce: when every case of a rule runs
// under one configuration, it is written once, above the cases, with every
// field the spec sets, and no case builds a rule of its own.
//
//nolint:paralleltest // workspace changes the working directory
func TestGenerateWritesASharedConfigurationOnce(t *testing.T) {
	specDir := workspace(t, map[string]string{
		"DL3049Spec.hs": "    let ?config = def { labelSchema = Map.fromList " +
			"[(\"version\", Rule.SemVer), (\"url\", Url)], strictLabels = True }\n" +
			"    it \"one\" $ ruleCatches \"DL3049\" \"FROM a\"\n" +
			"    it \"two\" $ ruleCatchesNot \"DL3049\" \"FROM b\"\n",
	}, "DL3049")

	var report strings.Builder
	if err := generate(specDir, &report); err != nil {
		t.Fatalf("generate: %v\n%s", err, report.String())
	}

	written, err := os.ReadFile("dl3049_test.go")
	if err != nil {
		t.Fatal(err)
	}

	cases := strings.Split(string(written), "t.Run(")
	if len(cases) != 3 {
		t.Fatalf("dl3049_test.go holds %d cases, want 2:\n%s", len(cases)-1, written)
	}

	const shared = `	cfg := &config.Config{
		LabelSchema: map[string]config.LabelType{
			"url":     config.LabelTypeURL,
			"version": config.LabelTypeSemVer,
		},
		StrictLabels: true,
	}
	allRules := []rule.Rule{
		rules.DL3049WithConfig(cfg),
	}
`

	if !strings.Contains(cases[0], shared) {
		t.Errorf("above the cases: want\n%s\ngot:\n%s", shared, cases[0])
	}

	for _, got := range cases[1:] {
		if strings.Contains(got, "allRules :=") || strings.Contains(got, "config.Config") {
			t.Errorf("a case builds a rule of its own under a shared configuration:\n%s", got)
		}
	}
}
