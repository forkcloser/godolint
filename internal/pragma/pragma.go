// Package pragma reads hadolint's pragmas out of Dockerfile comments: the
// ignore pragmas that silence rules for a line, a stage or the whole file,
// with hadolint's grammar and scoping. Ported from Hadolint/Pragma.hs, and
// the filtering half of Hadolint/Process.hs.
//
// A pragma is a comment of the form
//
//	# hadolint ignore=DL3006,DL3008        the instruction on the next line
//	# hadolint stage ignore=DL3006         the stage the FROM below it starts
//	# hadolint global ignore=DL3006        the whole file
//
// with any amount of spaces or tabs around the keywords, the "=" and the
// commas, and an optional "# comment" after the list. A rule name is a run
// of the characters D, L, S, C and digits. A list with any other name in it
// is not a pragma at all, and the whole comment is left alone — hadolint's
// parser fails as a whole, rather than keeping the names it could read.
package pragma

import (
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// codes is a set of rule codes.
type codes map[rule.Code]struct{}

// Directives is the ignore pragmas of one Dockerfile, keyed the way
// hadolint's three folds key them: by the line a pragma applies to.
type Directives struct {
	// line holds a line pragma under each line it applies to: the comment's
	// line + 1, and the lines after it of the instruction that starts there.
	line map[int]codes
	// stage holds a stage pragma under the line it applies to, and an empty
	// set under every FROM's line that has no pragma already: the boundary
	// that ends the previous stage's pragma.
	stage map[int]codes
	// global holds every global pragma's codes.
	global codes
}

// Parse reads the pragmas out of a Dockerfile's comments. Ported from the
// ignored, stageIgnored and globalIgnored folds in Hadolint.Pragma.
func Parse(instructions []syntax.InstructionPos) Directives {
	directives := Directives{
		line:   make(map[int]codes),
		stage:  make(map[int]codes),
		global: make(codes),
	}

	for _, instr := range instructions {
		// A line pragma covers the instruction below it, to its last line: a
		// failure may stand on a line of a heredoc's body.
		if covered, ok := directives.line[instr.LineNumber]; ok {
			for line := instr.LineNumber + 1; line <= instr.EndLine; line++ {
				directives.line[line] = covered
			}
		}

		switch node := instr.Instruction.(type) {
		case *syntax.Comment:
			directives.readComment(instr.LineNumber, node.Text)
		case *syntax.From:
			if _, ok := directives.stage[instr.LineNumber]; !ok {
				directives.stage[instr.LineNumber] = make(codes)
			}
		default:
			// Any other instruction carries no pragma.
		}
	}

	return directives
}

// ShouldIgnore reports whether a pragma silences the failure: a global
// pragma for its code, a line pragma on the instruction its line is one of,
// or the stage pragma in force at its line. Ported from shouldKeep in
// Hadolint.Process.
func (d *Directives) ShouldIgnore(failure rule.CheckFailure) bool {
	if _, ok := d.global[failure.Code]; ok {
		return true
	}

	if _, ok := d.line[failure.Line][failure.Code]; ok {
		return true
	}

	_, ok := d.stageAt(failure.Line)[failure.Code]

	return ok
}

// stageAt is the stage pragma in force at a line: the entry with the
// greatest key below the line (hadolint's `last (0 : keys below line)`),
// until the next FROM, whose empty entry ends it. The pragma is keyed on the
// line after it and the lookup wants a key strictly below, so that line
// itself is not covered: written above a FROM, the FROM takes it and the
// stage's instructions are; written mid-stage, the instruction right after
// the pragma is not. That is hadolint's reach, kept.
func (d *Directives) stageAt(line int) codes {
	key := 0

	for candidate := range d.stage {
		if candidate < line && candidate > key {
			key = candidate
		}
	}

	return d.stage[key]
}

// readComment records the pragma a comment at the given line holds, if any.
// The three grammars are disjoint, so a comment is at most one of them.
func (d *Directives) readComment(line int, text string) {
	if names, ok := ParseGlobalIgnore(text); ok {
		for _, code := range names {
			d.global[code] = struct{}{}
		}

		return
	}

	if names, ok := ParseStageIgnore(text); ok {
		d.stage[line+1] = toSet(names)

		return
	}

	if names, ok := ParseIgnore(text); ok {
		d.line[line+1] = toSet(names)
	}
}

func toSet(names []rule.Code) codes {
	set := make(codes, len(names))

	for _, code := range names {
		set[code] = struct{}{}
	}

	return set
}

// ParseIgnore reads a line ignore pragma, "hadolint ignore=DL3006,DL3008".
// Ported from parseIgnorePragma.
func ParseIgnore(text string) ([]rule.Code, bool) {
	scan := scanner{text: text}

	if !scan.pragma() {
		return nil, false
	}

	return scan.ignoreList()
}

// ParseStageIgnore reads a stage ignore pragma, "hadolint stage
// ignore=DL3006". Ported from parseStageIgnorePragma.
func ParseStageIgnore(text string) ([]rule.Code, bool) {
	scan := scanner{text: text}

	if !scan.pragma() || !scan.keyword("stage") {
		return nil, false
	}

	return scan.ignoreList()
}

// ParseGlobalIgnore reads a global ignore pragma, "hadolint global
// ignore=DL3006". Ported from parseGlobalIgnorePragma.
func ParseGlobalIgnore(text string) ([]rule.Code, bool) {
	scan := scanner{text: text}

	if !scan.pragma() || !scan.keyword("global") {
		return nil, false
	}

	return scan.ignoreList()
}

// scanner walks a comment's text with hadolint's pragma grammar. Every
// method consumes what it matched and reports whether it did.
type scanner struct {
	text string
	pos  int
}

// isSpace is hadolint's `space`: a space or a tab, never a newline.
func isSpace(c byte) bool {
	return c == ' ' || c == '\t'
}

// isRuleChar is the alphabet of a rule name.
func isRuleChar(c byte) bool {
	return c == 'D' || c == 'L' || c == 'S' || c == 'C' || (c >= '0' && c <= '9')
}

// spaces skips spaces and returns how many there were.
func (s *scanner) spaces() int {
	start := s.pos

	for s.pos < len(s.text) && isSpace(s.text[s.pos]) {
		s.pos++
	}

	return s.pos - start
}

// literal consumes the exact string.
func (s *scanner) literal(word string) bool {
	if !strings.HasPrefix(s.text[s.pos:], word) {
		return false
	}

	s.pos += len(word)

	return true
}

// pragma is hadolintPragma: spaces, "hadolint", at least one space.
func (s *scanner) pragma() bool {
	s.spaces()

	return s.literal("hadolint") && s.spaces() > 0
}

// keyword is the "stage" and "global" productions: the word, then at least
// one space.
func (s *scanner) keyword(word string) bool {
	return s.literal(word) && s.spaces() > 0
}

// ignoreList is "ignore", "=", the rule list, and nothing else but an
// inline comment before the end of the text.
func (s *scanner) ignoreList() ([]rule.Code, bool) {
	if !s.literal("ignore") {
		return nil, false
	}

	s.spaces()

	if !s.literal("=") {
		return nil, false
	}

	s.spaces()

	names, ok := s.ruleList()
	if !ok || s.pos != len(s.text) {
		return nil, false
	}

	return names, true
}

// ruleList is one or more rule names separated by commas, spaces allowed
// around the commas. Each name may be followed by an inline comment, which
// runs to the end of the text.
func (s *scanner) ruleList() ([]rule.Code, bool) {
	var names []rule.Code

	for {
		name, ok := s.ruleName()
		if !ok {
			return nil, false
		}

		names = append(names, name)

		// A separator: spaces, a comma, spaces. Nothing after the last name
		// but what ruleName's inline comment already took.
		mark := s.pos

		s.spaces()

		if !s.literal(",") {
			s.pos = mark

			return names, true
		}

		s.spaces()
	}
}

// ruleName is a non-empty run of rule characters, then an optional inline
// comment.
func (s *scanner) ruleName() (rule.Code, bool) {
	start := s.pos

	for s.pos < len(s.text) && isRuleChar(s.text[s.pos]) {
		s.pos++
	}

	if s.pos == start {
		return "", false
	}

	name := rule.Code(s.text[start:s.pos])

	s.inlineComment()

	return name, true
}

// inlineComment is spaces and, if a "#" follows, the rest of the line.
func (s *scanner) inlineComment() {
	s.spaces()

	if s.pos < len(s.text) && s.text[s.pos] == '#' {
		if end := strings.IndexByte(s.text[s.pos:], '\n'); end >= 0 {
			s.pos += end
		} else {
			s.pos = len(s.text)
		}
	}
}
