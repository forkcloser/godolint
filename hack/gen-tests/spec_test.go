package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// spec wraps the body of a describe in what a spec starts with.
func spec(body string) string {
	return "spec :: SpecWith ()\nspec = do\n  let ?config = def\n\n  describe \"a rule\" $ do\n" + body
}

type expected struct {
	name       string
	code       string
	dockerfile string
	kind       string // catches, not, count N, with "onbuild " before it
}

func kindOf(testCase TestCase) string {
	kind := "not"

	switch {
	case testCase.ShouldFail:
		kind = "catches"
	case testCase.Exactly && testCase.OnLines:
		kind = "one on a line"
	case testCase.Exactly:
		kind = "count " + string(rune('0'+testCase.Count))
	}

	if testCase.OnBuild {
		kind = "onbuild " + kind
	}

	return kind
}

// TestReadSpecReadsEveryShape: each shape hadolint's specs write a case in
// gives the cases its assertions make, with the Dockerfile whole.
func TestReadSpecReadsEveryShape(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		body string
		want []expected
	}{
		{
			"an assertion on the line of the it",
			`    it "explicit latest" $ ruleCatches "DL3007" "FROM debian:latest"` + "\n",
			[]expected{{"explicit latest", "DL3007", "FROM debian:latest", "catches"}},
		},
		{
			"an assertion over lines, its string continued by a gap",
			"    it \"explicit SHA\" $\n      ruleCatchesNot\n        \"DL3007\"\n" +
				"        \"FROM hub.docker.io/debian@sha256:\\\n        \\7959ed6f\"\n",
			[]expected{{"explicit SHA", "DL3007", "FROM hub.docker.io/debian@sha256:7959ed6f", "not"}},
		},
		{
			"a do block, escapes in the string",
			"    it \"two runs\" $ do\n      ruleCatches \"DL3059\" \"RUN /foo.sh\\n# a \\\"comment\\\"\\tx\\\\\\nRUN /bar.sh\"\n",
			[]expected{{"two runs", "DL3059", "RUN /foo.sh\n# a \"comment\"\tx\\\nRUN /bar.sh", "catches"}},
		},
		{
			"a list bound under any name, joined by Text.unlines",
			"    it \"flags\" $ do\n      let dfile = [ \"RUN a\",\n                    \"RUN b\"\n                  ]\n" +
				"       in ruleCatchesNot \"DL3059\" $ Text.unlines dfile\n",
			[]expected{{"flags", "DL3059", "RUN a\nRUN b", "not"}},
		},
		{
			"let, in, do, and the two assertions on one Dockerfile",
			"    it \"both\" $\n      let dockerFile =\n            [ \"FROM a\",\n              \"RUN b\"\n            ]\n" +
				"       in do\n            ruleCatches \"DL3009\" $ Text.unlines dockerFile\n" +
				"            onBuildRuleCatches \"DL3009\" $ Text.unlines dockerFile\n",
			[]expected{
				{"both", "DL3009", "FROM a\nRUN b", "catches"},
				{"both", "DL3009", "FROM a\nRUN b", "onbuild one on a line"},
			},
		},
		{
			"a text bound whole, a signature above the binding, the two Not",
			"    it \"text\" $ do\n      let dockerfile :: Text\n          dockerfile = Text.unlines [\"FROM a\", \"RUN b\"]\n" +
				"      ruleCatchesNot \"DL3064\" dockerfile\n      onBuildRuleCatchesNot \"DL3064\" dockerfile\n",
			[]expected{
				{"text", "DL3064", "FROM a\nRUN b", "not"},
				{"text", "DL3064", "FROM a\nRUN b", "onbuild not"},
			},
		},
		{
			"an assertion mapped over a list of Dockerfiles",
			"    it \"each\" $ do\n      let dockerFile :: [Text]\n          dockerFile =\n            [ \"RUN a\",\n              \"RUN b\"\n            ]\n" +
				"      mapM_ (ruleCatches \"DL3009\") dockerFile\n      mapM_ (onBuildRuleCatchesNot \"DL3009\") dockerFile\n",
			[]expected{
				{"each", "DL3009", "RUN a", "catches"},
				{"each", "DL3009", "RUN b", "catches"},
				{"each", "DL3009", "RUN a", "onbuild not"},
				{"each", "DL3009", "RUN b", "onbuild not"},
			},
		},
		{
			"a count, with and without the line breaks",
			"    it \"twice\" $\n      let dockerFile =\n            [ \"FROM a\",\n              \"FROM b\"\n            ]\n" +
				"       in assertChecks\n            (Text.unlines dockerFile)\n            (failsWith 2 \"DL3049\")\n" +
				"    it \"thrice\" $\n      let dockerFile = [\"FROM a\"]\n" +
				"       in assertChecks ( Text.unlines dockerFile ) ( failsWith 3 \"DL3026\" )\n" +
				"    it \"none, under ONBUILD\" $ assertOnBuildChecks \"RUN a\" $ failsWith 0 \"DL3026\"\n",
			[]expected{
				{"twice", "DL3049", "FROM a\nFROM b", "count 2"},
				{"thrice", "DL3026", "FROM a", "count 3"},
				{"none, under ONBUILD", "DL3026", "RUN a", "onbuild count 0"},
			},
		},
		{
			"a rule code bound once for the spec, strings joined",
			"    let rule = \"DL3065\"\n    it \"bound\" $ ruleCatches rule (\"FROM a\" <> \"\\n\" <> \"RUN b\")\n",
			[]expected{{"bound", "DL3065", "FROM a\nRUN b", "catches"}},
		},
		{
			"comments are not read",
			"    -- it \"gone\" $ ruleCatches \"DL3000\" \"x\"\n    {- it \"gone too\" $\n       ruleCatches \"DL3000\" \"y\" -}\n" +
				"    it \"kept\" $ ruleCatches \"DL3000\" \"WORKDIR a -- not a comment\"\n",
			[]expected{{"kept", "DL3000", "WORKDIR a -- not a comment", "catches"}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cases, err := readSpec(spec(tc.body))
			if err != nil {
				t.Fatalf("readSpec: %v", err)
			}

			var got []expected
			for _, testCase := range cases {
				got = append(got, expected{testCase.Name, testCase.RuleCode, testCase.Dockerfile, kindOf(testCase)})
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("cases:\n%+v\nwant:\n%+v", got, tc.want)
			}
		})
	}
}

// TestReadSpecConfigurations: `?config` is in force for the blocks below it
// at its depth or deeper, a block's own for that block alone, and `def` is
// none.
func TestReadSpecConfigurations(t *testing.T) {
	t.Parallel()

	src := "spec = do\n" +
		"  let ?config = def { labelSchema = Map.fromList [(\"foo\", Rule.RawText), (\"bar\", Url)], strictLabels = True }\n" +
		"  describe \"one\" $ do\n" +
		"    it \"under the spec's\" $ ruleCatches \"DL3049\" \"FROM a\"\n" +
		"    it \"under its own\" $ do\n" +
		"      let dockerFile = [\"FROM x.com/a\"]\n" +
		"      let ?config = def { allowedRegistries = [\"x.com\", \"*.random.com\"] }\n" +
		"      ruleCatchesNot \"DL3026\" $ Text.unlines dockerFile\n" +
		"    it \"under the spec's again\" $ ruleCatches \"DL3049\" \"FROM b\"\n" +
		"    it \"under none\" $\n" +
		"      let ?config = def\n" +
		"       in ruleCatchesNot \"DL3049\" \"FROM c\"\n" +
		"  describe \"two\" $ do\n" +
		"    let ?config = def { strictLabels = False }\n" +
		"    it \"under the describe's\" $ ruleCatchesNot \"DL3050\" \"FROM d\"\n" +
		"  describe \"three\" $ do\n" +
		"    it \"under the spec's, after a deeper one\" $ ruleCatches \"DL3049\" \"FROM e\"\n"

	cases, err := readSpec(src)
	if err != nil {
		t.Fatalf("readSpec: %v", err)
	}

	whole := &HadolintConfig{LabelSchema: map[string]string{"foo": "RawText", "bar": "Url"}, StrictLabels: true}
	want := []*HadolintConfig{
		whole,
		{LabelSchema: map[string]string{}, AllowedRegistries: []string{"x.com", "*.random.com"}},
		whole,
		nil,
		{LabelSchema: map[string]string{}},
		whole,
	}

	if len(cases) != len(want) {
		t.Fatalf("%d cases, want %d", len(cases), len(want))
	}

	for i, testCase := range cases {
		if !reflect.DeepEqual(testCase.Config, want[i]) {
			t.Errorf("%q: configuration %+v, want %+v", testCase.Name, testCase.Config, want[i])
		}
	}

	if cases[0].Config != cases[2].Config || cases[0].Config != cases[5].Config {
		t.Error("the spec's configuration is not one value for the blocks under it")
	}
}

// TestReadSpecRefusesWhatItCannotRead: a block of a shape this does not know
// is an error naming the block, never a case left out.
func TestReadSpecRefusesWhatItCannotRead(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		body  string
		named string
	}{
		{"an assertion it does not know", `    it "odd" $ assertFormatter Json [] ""` + "\n", `it "odd"`},
		{"a count that is not failsWith", `    it "odd" $ assertChecks "RUN a" passesShellcheck` + "\n", "passesShellcheck"},
		{"a name bound nowhere", `    it "odd" $ ruleCatches "DL3000" dockerFile` + "\n", `"dockerFile" is not bound`},
		{"a block with no assertion", "    it \"odd\" $ do\n      let dockerFile = [\"a\"]\n", "no assertion"},
		{"a list where a Dockerfile goes", "    it \"odd\" $ do\n      let d = [\"a\"]\n      ruleCatches \"DL3000\" d\n", "a list where"},
		{"Text.unlines of a text", `    it "odd" $ ruleCatches "DL3000" $ Text.unlines "a"` + "\n", "not a list"},
		{"mapM_ over a text", `    it "odd" $ mapM_ (ruleCatches "DL3000") "a"` + "\n", "not a list"},
		{"mapM_ of something else", `    it "odd" $ mapM_ (print) ["a"]` + "\n", "mapM_ of"},
		{"a configuration field it does not know", "    it \"odd\" $ do\n      let ?config = def { noFail = True }\n      ruleCatches \"DL3000\" \"a\"\n", `"noFail"`},
		{"a string that does not end", `    it "odd" $ ruleCatches "DL3000" "a` + "\n", "does not end"},
		{"an escape it does not know", `    it "odd" $ ruleCatches "DL3000" "a\qb"` + "\n", `\q`},
		{"an it that does not start its line", `    describe "inner" $ it "odd" $ ruleCatches "DL3000" "a"` + "\n", "1 `it` blocks written, 0 read"},
		{"an it with no $", `    it "odd" ruleCatches "DL3000" "a"` + "\n", `"$" is expected`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cases, err := readSpec(spec(tc.body))
			if !errors.Is(err, ErrSpec) || cases != nil {
				t.Fatalf("readSpec = %v, %v; want an unreadable spec", cases, err)
			}

			if !strings.Contains(err.Error(), tc.named) {
				t.Errorf("error %q; want it to name %s", err, tc.named)
			}
		})
	}
}

// TestNameCases: each case of a rule gets its own name: the block's, with
// ONBUILD said and a number after the first of a name.
func TestNameCases(t *testing.T) {
	t.Parallel()

	cases := []TestCase{
		{Name: "a"}, {Name: "a", OnBuild: true}, {Name: "b"}, {Name: "a"}, {Name: "a", OnBuild: true}, {Name: "a"},
	}
	nameCases(cases)

	var got []string
	for _, testCase := range cases {
		got = append(got, testCase.Name)
	}

	want := []string{"a", "a (ONBUILD)", "b", "a [2]", "a (ONBUILD) [2]", "a [3]"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("names %q, want %q", got, want)
	}
}
