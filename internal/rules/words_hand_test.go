package rules_test

import (
	"slices"
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestShellWordsHand: the rules read a RUN's words as hadolint does, where
// its cases say nothing: what an expansion holds is seen, a backslash
// escapes, a brace expansion is nothing. Each row is hadolint's verdict,
// read from simplify in Shell.hs and the rule.
func TestShellWordsHand(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		rule       func() rule.Rule
		dockerfile string
		lines      []int
	}{
		{"pip, the packages of a file", rules.DL3013, "FROM a\nRUN pip install $(cat /tmp/requirements.txt)\n", nil},
		{"pip, a version that may be given", rules.DL3013, "FROM a\nRUN pip install \"requests${V:+==$V}\"\n", nil},
		{"pip, a package and no version", rules.DL3013, "FROM a\nRUN pip install \"requests${EXTRAS}\"\n", []int{2}},
		{"npm, a version that may be given", rules.DL3016, "FROM a\nRUN npm install \"express${V:+@$V}\"\n", nil},
		{"apk, a version that may be given", rules.DL3018, "FROM a\nRUN apk add --no-cache \"curl${V:+=$V}\"\n", nil},
		{"apt-get, the packages of a file", rules.DL3008, "FROM a\nRUN apt-get install -y $(cat /tmp/pkgs.txt)\n", nil},
		{"apt-get, a default with a version", rules.DL3008, "FROM a\nRUN apt-get install -y ${PKGS:-foo=1.0}\n", nil},
		{"apt-get, a variable alone", rules.DL3008, "FROM a\nRUN apt-get install -y $PKGS\n", []int{2}},
		{"go install, a default with a version", rules.DL3062, "FROM a\nRUN go install ${TOOL:-example.com/x/cmd/tool@v1.2.3}\n", nil},
		{"yum, sibling packages of one version", rules.DL3033, "FROM a\nRUN yum install -y httpd-{devel,tools}-2.4.6 && yum clean all\n", nil},
		{"dnf, sibling packages of one version", rules.DL3041, "FROM a\nRUN dnf install -y httpd-{devel,tools}-2.4.6 && dnf clean all\n", nil},
		{"dnf, an escaped glob for the version", rules.DL3041, "FROM a\nRUN dnf install -y foo-1.2\\* && dnf clean all\n", nil},
		{"yum, the cache removed by an escaped glob", rules.DL3032, "FROM a\nRUN yum install -y foo-1.0 && rm -rf /var/cache/yum/\\*\n", nil},
		{
			"apt, the lists removed by an rm written with a backslash",
			rules.DL3009,
			"FROM a\nRUN apt-get update && apt-get install -y curl=1 && \\rm -rf /var/lib/apt/lists/*\n",
			nil,
		},
		{
			"curl written with a backslash, then wget",
			rules.DL4001,
			"FROM a\nRUN \\curl -sSL https://get.rvm.io -o /tmp/rvm.sh\nRUN wget -q https://example.com/x\n",
			[]int{3},
		},
		{"sudo written with a backslash", rules.DL3004, "FROM a\nRUN \\sudo true\n", []int{2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var lines []int
			for _, v := range testutils.LintDockerfile(tc.dockerfile, []rule.Rule{tc.rule()}) {
				lines = append(lines, v.Line)
			}

			slices.Sort(lines)

			if !slices.Equal(lines, tc.lines) {
				t.Errorf("reported at lines %v, want %v", lines, tc.lines)
			}
		})
	}
}
