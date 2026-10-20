package sdk_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/forkcloser/godolint/sdk"
)

// TestContractLintFileReadsAndLints: LintFile lints the file at the path
// as Lint lints its bytes; a missing path is an error, never a panic.
func TestContractLintFileReadsAndLints(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "Dockerfile")
	if err := os.WriteFile(path, []byte("FROM ubuntu\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	r, err := sdk.New().LintFile(t.Context(), path)
	if err != nil || !slices.Contains(codes(r), "DL3006") {
		t.Errorf("LintFile: %v, %v", err, r)
	}

	if _, err := sdk.New().LintFile(t.Context(), filepath.Join(t.TempDir(), "nonesuch")); err == nil {
		t.Error("a missing file: no error")
	}
}

// TestContractWhatHadolintRefusesIsAParseError: `FROM` without an image,
// an instruction no Dockerfile has, and bytes that are no Dockerfile are
// a ParseError naming the line, never a clean result, as hadolint refuses
// them at parse.
func TestContractWhatHadolintRefusesIsAParseError(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"FROM\n", "\x00\xff", "FROM debian\nFOO bar\n"} {
		r, err := sdk.New().Lint(t.Context(), []byte(in))

		var perr *sdk.ParseError
		if !errors.As(err, &perr) || r != nil {
			t.Errorf("%q: %v (result %v), want a ParseError and no result", in, err, r)
		}
	}
}
