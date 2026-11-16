package rules

import (
	"maps"
	"slices"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// failingLabels is the labels of one LABEL instruction, among those wanted,
// that have a value valid refuses: each label once, by name. hadolint has one
// rule a label of the schema, so a LABEL with two bad labels is reported
// twice, and one that gives a bad label twice is reported once.
func failingLabels(pairs []syntax.LabelPair, wanted func(key string) bool, valid func(value string) bool) []string {
	failing := map[string]bool{}

	for _, pair := range pairs {
		if wanted(pair.Key) && !valid(pair.Value) {
			failing[pair.Key] = true
		}
	}

	return slices.Sorted(maps.Keys(failing))
}

// labelFailure is the failure a label rule reports at a line, under the
// message that names the label.
func labelFailure(r rule.Rule, line int, message string) rule.CheckFailure {
	failure := rule.Fail(r, line)
	failure.Message = message

	return failure
}
