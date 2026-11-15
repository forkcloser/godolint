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
