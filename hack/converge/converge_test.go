// Package converge_test holds .converge to what the generators write.
package converge_test

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// generated is the line a Go file that is generated opens with.
var generated = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)

// TestConvergeNamesWhatIsGenerated: .converge names every Go file of the
// repository that says it is generated, and no other. A file a generator
// starts to write and nobody declares, or one that is declared and is not
// generated, would be left alone by `just converge`.
func TestConvergeNamesWhatIsGenerated(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")

	list, err := os.ReadFile(filepath.Join(root, ".converge"))
	if err != nil {
		t.Fatal(err)
	}

	var declared []string

	for line := range strings.Lines(string(list)) {
		if path := strings.TrimSpace(line); path != "" && !strings.HasPrefix(path, "#") {
			declared = append(declared, path)
		}
	}

	var marked []string

	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// What git does not track: its own directory, the build's, a tool's.
		if entry.IsDir() && path != root && strings.HasPrefix(entry.Name(), ".") {
			return filepath.SkipDir
		}

		if entry.IsDir() || filepath.Ext(path) != ".go" || !saysGenerated(t, path) {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		marked = append(marked, filepath.ToSlash(rel))

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	slices.Sort(declared)
	slices.Sort(marked)

	for _, path := range marked {
		if !slices.Contains(declared, path) {
			t.Errorf("%s says it is generated and .converge does not name it", path)
		}
	}

	for _, path := range declared {
		if !slices.Contains(marked, path) {
			t.Errorf(".converge names %s, which is not there or does not say it is generated", path)
		}
	}
}

// saysGenerated reports whether the file's first line is the mark of
// generated code.
func saysGenerated(t *testing.T, path string) bool {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)

	return scanner.Scan() && generated.MatchString(scanner.Text())
}
