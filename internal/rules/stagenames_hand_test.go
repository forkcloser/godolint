package rules_test

import (
	"slices"
	"testing"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestStageNamesAreLowerCase: a stage's name is what a build keeps of it,
// its lower case. BuildKit lowers the name after AS when it reads a FROM,
// looks the base of a FROM up as written among those names, and looks the
// source of a COPY --from or of a mount up in lower case. So a stage named
// Base is the base of `FROM base`, the source of `COPY --from=BASE`, and
// the same name as base given twice.
func TestStageNamesAreLowerCase(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		dockerfile string
		rule       rule.Rule
		code       rule.Code
		reported   bool
	}{
		{
			"a FROM on a stage named in capitals is no untagged image",
			"FROM a:1 AS Base\nFROM base\n", rules.DL3006(), "DL3006", false,
		},
		{
			"a FROM on another name is one",
			"FROM a:1 AS Base\nFROM other\n", rules.DL3006(), "DL3006", true,
		},
		{
			"a COPY from a stage named in capitals, asked for in lower case",
			"FROM a:1 AS Build\nFROM b:1\nCOPY --from=build /x /y\n", rules.DL3022(), "DL3022", false,
		},
		{
			"a COPY from a stage named in lower case, asked for in capitals",
			"FROM a:1 AS build\nFROM b:1\nCOPY --from=BUILD /x /y\n", rules.DL3022(), "DL3022", false,
		},
		{
			"a COPY from the stage it is in, asked for in another case",
			"FROM a:1 AS build\nCOPY --from=Build /x /y\n", rules.DL3023(), "DL3023", true,
		},
		{
			"one name given twice, in two cases",
			"FROM a:1 AS base\nFROM b:1 AS Base\n", rules.DL3024(), "DL3024", true,
		},
		{
			"a COPY from a stage, asked for in capitals, is from no registry",
			"FROM example.com/a:1 AS build\nFROM example.com/b:1\nCOPY --from=BUILD /x /y\n",
			rules.DL3026WithConfig(&config.Config{AllowedRegistries: []string{"example.com"}}), "DL3026", false,
		},
		{
			"a COPY from an image keeps its registry as written",
			"FROM example.com/a:1\nCOPY --from=EXAMPLE.com/b:1 /x /y\n",
			rules.DL3026WithConfig(&config.Config{AllowedRegistries: []string{"example.com"}}), "DL3026", true,
		},
		{
			"a COPY from a stage, asked for in capitals, answers for its label",
			"FROM a:1 AS base\nFROM b:1\nLABEL foo=\"bar\"\nCOPY --from=BASE /x /y\n",
			rules.DL3049WithConfig(&config.Config{LabelSchema: map[string]config.LabelType{"foo": config.LabelTypeRawText}}),
			"DL3049", false,
		},
		{
			"a mount from a stage, asked for in capitals, is from no registry",
			"FROM example.com/a:1 AS build\nFROM example.com/b:1\nRUN --mount=type=bind,from=BUILD,target=/x true\n",
			rules.DL3026WithConfig(&config.Config{AllowedRegistries: []string{"example.com"}}), "DL3026", false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			failures := testutils.LintDockerfile(tc.dockerfile, []rule.Rule{tc.rule})

			reported := slices.ContainsFunc(failures, func(failure rule.CheckFailure) bool {
				return failure.Code == tc.code
			})
			if reported != tc.reported {
				t.Errorf("%s reported: %t, want %t", tc.code, reported, tc.reported)
			}
		})
	}
}
