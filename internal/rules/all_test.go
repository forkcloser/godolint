package rules_test

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/forkcloser/godolint/internal/rules"
)

// TestAllHoldsEveryRule: All lists every rule that has an implementation
// here, once, and nothing else. A rule left out of the list is linted by
// nothing, and no other test would say so.
func TestAllHoldsEveryRule(t *testing.T) {
	t.Parallel()

	files, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}

	var implemented []string

	for _, file := range files {
		if code, isRule := strings.CutSuffix(file.Name(), "_impl.go"); isRule {
			implemented = append(implemented, strings.ToUpper(code))
		}
	}

	var listed []string
	for _, r := range rules.All() {
		listed = append(listed, string(r.Code()))
	}

	slices.Sort(implemented)
	slices.Sort(listed)

	if len(implemented) == 0 || !slices.Equal(listed, implemented) {
		t.Errorf("All lists\n%v\nthe rules implemented are\n%v", listed, implemented)
	}
}
