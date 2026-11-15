package main

import (
	"os"
	"path/filepath"
	"testing"
)

// A Haskell string gap (backslash, line break and indentation, backslash)
// is deleted whole: the message keeps the one space before the gap, not the
// indentation inside it.
func TestParseRuleFileCollapsesStringGaps(t *testing.T) {
	t.Parallel()

	src := `rule :: Rule args
rule = simpleRule code severity message check
  where
    code = "DL3007"
    severity = DLWarningC
    message =
      "Using latest is prone to errors if the image will ever update. Pin the version explicitly \
      \to a release tag"

    check (From BaseImage {tag = Just t}) = t /= "latest"
    check _ = True
`

	path := filepath.Join(t.TempDir(), "DL3007.hs")
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	rule, err := parseRuleFile(path)
	if err != nil {
		t.Fatal(err)
	}

	want := "Using latest is prone to errors if the image will ever update. Pin the version explicitly to a release tag"
	if rule.Message != want {
		t.Errorf("Message = %q\nwant      %q", rule.Message, want)
	}
}

// A message joined from literals and a variable is the literals with the
// variable's name in angle brackets, Text.pack or not: the whole sentence, not
// its first literal.
func TestParseRuleFileJoinsConcatenatedMessage(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		message string
		want    string
	}{
		"plain": {
			"\"Label `\" <> label <> \"` is missing.\"",
			"Label `<label>` is missing.",
		},
		"under Text.pack": {
			"Text.pack \"Label `\" <> label <> Text.pack \"` does not conform to semantic versioning.\"",
			"Label `<label>` does not conform to semantic versioning.",
		},
		"broken over lines": {
			"\n      \"Label `\"\n        <> label\n        <> \"` is empty.\"",
			"Label `<label>` is empty.",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			src := "rule = veryCustomRule check (emptyState Empty) markFailure\n  where\n" +
				"    code = \"DL3049\"\n    severity = DLInfoC\n    message = " + tc.message + "\n" +
				"    check _ state _ = state\n"

			path := filepath.Join(t.TempDir(), "DL3049.hs")
			if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
				t.Fatal(err)
			}

			rule, err := parseRuleFile(path)
			if err != nil {
				t.Fatal(err)
			}

			if rule.Message != tc.want {
				t.Errorf("Message = %q\nwant      %q", rule.Message, tc.want)
			}
		})
	}
}

// A message the generator does not read whole is refused, never recorded in
// part: one that does not start with a literal, be it a name alone, another
// operator after a literal, a function applied to an argument, and an
// expression that goes on on the next line.
func TestParseRuleFileRefusesWhatItDoesNotRead(t *testing.T) {
	t.Parallel()

	for name, message := range map[string]string{
		"a computed message":                  "describe label",
		"a name alone":                        "describedElsewhere",
		"a variable first":                    "label <> \" is missing.\"",
		"another operator":                    "\"Found \" ++ label ++ \" things.\"",
		"a function between literals":         "\"Found \" <> show n <> \" things.\"",
		"a function at the end":               "\"Found \" <> show n",
		"a qualified function":                "\"Found \" <> Text.pack label",
		"a parenthesis":                       "\"Found \" <> (label)",
		"another operator on the next line":   "\"Found \"\n        ++ label",
		"another operator below a blank line": "\"Found \"\n\n        ++ label",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			src := "rule = simpleRule code severity message check\n  where\n" +
				"    code = \"DL9999\"\n    severity = DLInfoC\n    message = " + message + "\n" +
				"    check _ = True\n"

			path := filepath.Join(t.TempDir(), "DL9999.hs")
			if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
				t.Fatal(err)
			}

			if rule, err := parseRuleFile(path); err == nil {
				t.Errorf("parseRuleFile read %q as %q", message, rule.Message)
			}
		})
	}
}
