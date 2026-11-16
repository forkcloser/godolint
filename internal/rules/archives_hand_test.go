package rules_test

import (
	"testing"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/rules"
	"github.com/forkcloser/godolint/internal/testutils"
)

// TestArchiveSuffixesHand: an archive, to the two rules that ask, is what a
// build's ADD unpacks: a tar, plain or under gzip, bzip2, xz or zstd,
// whatever the case of its name. DL3020 asks for COPY when ADD brings
// anything else, and DL3010 for ADD when a COPY of one is then unpacked by
// tar.
func TestArchiveSuffixesHand(t *testing.T) {
	t.Parallel()

	for name, unpacked := range map[string]bool{
		"a.tar": true, "a.tar.gz": true, "a.tgz": true, "a.taz": true,
		"a.tar.bz2": true, "a.tz2": true, "a.tbz": true, "a.tbz2": true, "a.tb2": true,
		"a.tar.xz": true, "a.txz": true,
		"a.tar.zst": true, "a.tzst": true,
		"A.TAR.GZ": true, "a.Tgz": true,

		"a.tar.Z": false, "a.tZ": false,
		"a.tar.lz": false, "a.tar.lzma": false, "a.tlz": false,
		"a.tar.lzo": false, "a.tzo": false, "a.t7z": false, "a.7z": false,
		"a.tpz": false, "a.tz": false, "a.zip": false, "a.txt": false, "tar": false,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			added := testutils.LintDockerfile("FROM a\nADD "+name+" /\n", []rule.Rule{rules.DL3020()})
			if asksForCopy := len(added) != 0; asksForCopy == unpacked {
				t.Errorf("ADD %s: DL3020 reported %v, want %v", name, asksForCopy, !unpacked)
			}

			copied := testutils.LintDockerfile(
				"FROM a\nCOPY "+name+" /tmp/\nRUN tar -xf /tmp/"+name+"\n", []rule.Rule{rules.DL3010()},
			)
			if asksForAdd := len(copied) != 0; asksForAdd != unpacked {
				t.Errorf("COPY %s then tar -x: DL3010 reported %v, want %v", name, asksForAdd, unpacked)
			}
		})
	}
}
