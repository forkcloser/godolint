package main

import (
	"errors"
	"fmt"
	"maps"
	"strconv"
	"strings"
	"unicode"
)

// The reader of hadolint's specs. A spec is Haskell, and this reads the part
// of Haskell the specs are written in: `it "name" $ …` blocks, the bindings
// in them, `?config`, and the assertions of test/Helpers.hs. Each block is
// read whole or refused: a shape this does not know is an error naming the
// block, never a case left out.

// ErrSpec is a spec this reader cannot read.
var ErrSpec = errors.New("unreadable spec")

type tokenKind int

const (
	tokenIdent tokenKind = iota
	tokenString
	tokenNumber
	tokenSymbol
)

// token is one lexeme, with where it stands: the layout of a spec is its
// indentation, so a block ends at the first token that starts a line no
// deeper than the block's own first token.
type token struct {
	kind       tokenKind
	text       string // an identifier, a symbol, a number, or a string's value
	line, col  int
	startsLine bool
}

func (t token) is(kind tokenKind, text string) bool {
	return t.kind == kind && t.text == text
}

// lex splits a spec into tokens. Comments are dropped.
func lex(src string) ([]token, error) {
	var tokens []token

	line, col, lineHasToken := 1, 1, false
	pos := 0

	advance := func(n int) {
		for range n {
			if src[pos] == '\n' {
				line++
				col = 1
				lineHasToken = false
			} else {
				col++
			}

			pos++
		}
	}

	emit := func(kind tokenKind, text string, startLine, startCol int) {
		tokens = append(
			tokens,
			token{kind: kind, text: text, line: startLine, col: startCol, startsLine: !lineHasToken},
		)
		lineHasToken = true
	}

	for pos < len(src) {
		char := src[pos]

		switch {
		case char == ' ' || char == '\t' || char == '\r' || char == '\n':
			advance(1)
		case strings.HasPrefix(src[pos:], "{-"):
			end := strings.Index(src[pos:], "-}")
			if end < 0 {
				return nil, fmt.Errorf("%w: line %d: a block comment that does not end", ErrSpec, line)
			}

			advance(end + len("-}"))
		case strings.HasPrefix(src[pos:], "--") && !isSymbolChar(at(src, pos+len("--"))):
			end := strings.IndexByte(src[pos:], '\n')
			if end < 0 {
				end = len(src) - pos
			}

			advance(end)
		case char == '"':
			startLine, startCol := line, col
			wasFirst := !lineHasToken

			value, length, err := lexString(src[pos:])
			if err != nil {
				return nil, fmt.Errorf("%w: line %d: %w", ErrSpec, line, err)
			}

			advance(length)
			// A string that runs over lines is still the first of the line it
			// began on.
			tokens = append(
				tokens,
				token{kind: tokenString, text: value, line: startLine, col: startCol, startsLine: wasFirst},
			)
			lineHasToken = true
		case isIdentStart(char):
			end := pos + 1
			for end < len(src) && isIdentChar(src[end]) {
				end++
			}

			// A qualified name ends on a letter, not on the dot of what follows.
			for src[end-1] == '.' {
				end--
			}

			emit(tokenIdent, src[pos:end], line, col)
			advance(end - pos)
		case char >= '0' && char <= '9':
			end := pos + 1
			for end < len(src) && src[end] >= '0' && src[end] <= '9' {
				end++
			}

			emit(tokenNumber, src[pos:end], line, col)
			advance(end - pos)
		case strings.IndexByte("()[]{},", char) >= 0:
			emit(tokenSymbol, string(char), line, col)
			advance(1)
		case isSymbolChar(char):
			end := pos + 1
			for end < len(src) && isSymbolChar(src[end]) {
				end++
			}

			emit(tokenSymbol, src[pos:end], line, col)
			advance(end - pos)
		default:
			return nil, fmt.Errorf("%w: line %d: character %q", ErrSpec, line, char)
		}
	}

	return tokens, nil
}

func at(src string, pos int) byte {
	if pos >= len(src) {
		return 0
	}

	return src[pos]
}

func isIdentStart(char byte) bool {
	return char == '_' || char == '?' || (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}

func isIdentChar(char byte) bool {
	return isIdentStart(char) && char != '?' || char == '\'' || char == '.' || (char >= '0' && char <= '9')
}

func isSymbolChar(char byte) bool {
	return char != 0 && strings.IndexByte("!#$%&*+./<=>@\\^|-~:", char) >= 0
}

// lexString reads the string literal src starts with: its value, escapes
// resolved and gaps closed, and how many bytes of src it takes.
func lexString(src string) (string, int, error) {
	var value strings.Builder

	pos := 1
	for pos < len(src) {
		char := src[pos]

		switch {
		case char == '"':
			return value.String(), pos + 1, nil
		case char != '\\':
			value.WriteByte(char)

			pos++
		case pos+1 >= len(src):
			return "", 0, errUnterminated
		case isSpace(src[pos+1]):
			// A gap: backslash, white space, backslash. Nothing of it is in
			// the string.
			end := pos + 1
			for end < len(src) && isSpace(src[end]) {
				end++
			}

			if at(src, end) != '\\' {
				return "", 0, errGap
			}

			pos = end + 1
		case src[pos+1] >= '0' && src[pos+1] <= '9':
			end := pos + 1
			for end < len(src) && src[end] >= '0' && src[end] <= '9' {
				end++
			}

			code, err := strconv.ParseInt(src[pos+1:end], 10, 32)
			if err != nil || code > unicode.MaxRune {
				return "", 0, fmt.Errorf("%w %s", errEscape, src[pos:end])
			}

			value.WriteRune(rune(code))

			pos = end
		default:
			escaped, known := escapes[src[pos+1]]
			if !known {
				return "", 0, fmt.Errorf("%w \\%c", errEscape, src[pos+1])
			}

			value.WriteString(escaped)

			pos += len(`\n`)
		}
	}

	return "", 0, errUnterminated
}

var (
	errUnterminated = errors.New("a string that does not end")
	errGap          = errors.New("a string gap that does not close")
	errEscape       = errors.New("unknown escape")

	// escapes is Haskell's single-character escapes; \& is the empty string.
	//
	//nolint:gochecknoglobals // a constant table
	escapes = map[byte]string{
		'0': "\x00", 'a': "\x07", 'b': "\x08", 'f': "\x0c", 'n': "\n", 'r': "\r", 't': "\t", 'v': "\x0b",
		'"': `"`, '\'': "'", '\\': `\`, '&': "",
	}
)

func isSpace(char byte) bool {
	return char == ' ' || char == '\t' || char == '\r' || char == '\n'
}

// value is what an expression of a spec is worth: a text, or a list of
// texts.
type value struct {
	text   string
	list   []string
	isList bool
}

// The assertions of test/Helpers.hs a block is read for.
const (
	assertCatches           = "ruleCatches"
	assertCatchesNot        = "ruleCatchesNot"
	assertOnBuildCatches    = "onBuildRuleCatches"
	assertOnBuildCatchesNot = "onBuildRuleCatchesNot"
	assertChecks            = "assertChecks"
	assertOnBuildChecks     = "assertOnBuildChecks"
)

// block reads one `it`: its tokens, what its names are bound to, and the
// configuration in force.
type block struct {
	name   string
	tokens []token
	pos    int
	env    map[string]value
	config *HadolintConfig
	// configured is whether the block bound `?config` itself.
	configured bool
	cases      []TestCase
}

func (b *block) peek() token {
	if b.pos >= len(b.tokens) {
		return token{kind: tokenSymbol, text: "<end>"}
	}

	return b.tokens[b.pos]
}

func (b *block) next() token {
	tok := b.peek()
	b.pos++

	return tok
}

func (b *block) errorf(format string, args ...any) error {
	tok := b.peek()
	if b.pos > 0 && b.pos >= len(b.tokens) {
		tok = b.tokens[len(b.tokens)-1]
	}

	return fmt.Errorf("%w: it %q, line %d: %s", ErrSpec, b.name, tok.line, fmt.Sprintf(format, args...))
}

func (b *block) expect(kind tokenKind, text string) error {
	if tok := b.next(); !tok.is(kind, text) {
		b.pos--

		return b.errorf("%q where %q is expected", tok.text, text)
	}

	return nil
}

// read reads the statements of the block: bindings and assertions.
func (b *block) read() error {
	for b.pos < len(b.tokens) {
		tok := b.peek()

		switch {
		case tok.is(tokenIdent, "do"), tok.is(tokenIdent, "in"):
			b.pos++
		case tok.is(tokenIdent, "let"):
			b.pos++

			if err := b.bindings(); err != nil {
				return err
			}
		case tok.kind == tokenIdent:
			if err := b.assertion(); err != nil {
				return err
			}
		default:
			return b.errorf("%q where a binding or an assertion is expected", tok.text)
		}
	}

	return nil
}

// bindings reads what follows a `let`: names bound to texts and lists, the
// configuration, and the type signatures written above a binding.
func (b *block) bindings() error {
	for b.pos+1 < len(b.tokens) && b.peek().kind == tokenIdent {
		name, after := b.peek(), b.tokens[b.pos+1]

		switch {
		case after.is(tokenSymbol, "::"):
			// The signature is the rest of its line.
			b.pos++
			for b.pos < len(b.tokens) && !b.tokens[b.pos].startsLine {
				b.pos++
			}
		case !after.is(tokenSymbol, "="):
			return nil
		case name.text == "?config":
			b.pos += len("?=")

			config, err := b.configuration()
			if err != nil {
				return err
			}

			b.config, b.configured = config, true
		default:
			b.pos += len("n=")

			bound, err := b.expression()
			if err != nil {
				return err
			}

			b.env[name.text] = bound
		}
	}

	return nil
}

// expression reads a text or a list of texts: terms joined by `<>`.
func (b *block) expression() (value, error) {
	left, err := b.term()
	if err != nil {
		return value{}, err
	}

	for b.peek().is(tokenSymbol, "<>") {
		b.pos++

		right, err := b.term()
		if err != nil {
			return value{}, err
		}

		if left.isList || right.isList {
			return value{}, b.errorf("`<>` between lists")
		}

		left.text += right.text
	}

	return left, nil
}

// term reads a string, a list, a name, a parenthesised expression, what
// follows a `$`, or `Text.unlines` of a list.
func (b *block) term() (value, error) {
	tok := b.next()

	switch {
	case tok.kind == tokenString:
		return value{text: tok.text}, nil
	case tok.is(tokenSymbol, "$"):
		return b.expression()
	case tok.is(tokenSymbol, "("):
		inner, err := b.expression()
		if err != nil {
			return value{}, err
		}

		return inner, b.expect(tokenSymbol, ")")
	case tok.is(tokenSymbol, "["):
		return b.list()
	case tok.is(tokenIdent, "Text.unlines"), tok.is(tokenIdent, "unlines"):
		lines, err := b.term()
		if err != nil {
			return value{}, err
		}

		if !lines.isList {
			return value{}, b.errorf("%s of what is not a list", tok.text)
		}

		return value{text: strings.Join(lines.list, "\n")}, nil
	case tok.kind == tokenIdent:
		bound, ok := b.env[tok.text]
		if !ok {
			b.pos--

			return value{}, b.errorf("%q is not bound to a text", tok.text)
		}

		return bound, nil
	default:
		b.pos--

		return value{}, b.errorf("%q where a text is expected", tok.text)
	}
}

// list reads the elements of a list whose `[` is read.
func (b *block) list() (value, error) {
	items := []string{}

	for !b.peek().is(tokenSymbol, "]") {
		item, err := b.expression()
		if err != nil {
			return value{}, err
		}

		if item.isList {
			return value{}, b.errorf("a list in a list")
		}

		items = append(items, item.text)

		if b.peek().is(tokenSymbol, ",") {
			b.pos++
		} else if !b.peek().is(tokenSymbol, "]") {
			return value{}, b.errorf("%q in a list", b.peek().text)
		}
	}

	b.pos++

	return value{list: items, isList: true}, nil
}

// text reads an expression that must be a text.
func (b *block) text() (string, error) {
	read, err := b.expression()
	if err != nil {
		return "", err
	}

	if read.isList {
		return "", b.errorf("a list where a Dockerfile is expected")
	}

	return read.text, nil
}

// code reads a rule code: a string, or a name bound to one.
func (b *block) code() (string, error) {
	tok := b.next()

	switch {
	case tok.kind == tokenString:
		return tok.text, nil
	case tok.kind == tokenIdent && !b.env[tok.text].isList && b.env[tok.text].text != "":
		return b.env[tok.text].text, nil
	default:
		b.pos--

		return "", b.errorf("%q where a rule code is expected", tok.text)
	}
}

// assertion reads one assertion and adds its cases.
func (b *block) assertion() error {
	tok := b.next()

	switch tok.text {
	case assertCatches, assertCatchesNot, assertOnBuildCatches, assertOnBuildCatchesNot:
		code, err := b.code()
		if err != nil {
			return err
		}

		dockerfile, err := b.text()
		if err != nil {
			return err
		}

		b.add(tok.text, code, dockerfile, 0)

		return nil
	case "mapM_":
		return b.mapped()
	case assertChecks, assertOnBuildChecks:
		return b.counted(tok.text)
	default:
		b.pos--

		return b.errorf("%q is not an assertion this reads", tok.text)
	}
}

// mapped reads `mapM_ (assertion code) list`: the assertion on each element.
func (b *block) mapped() error {
	if err := b.expect(tokenSymbol, "("); err != nil {
		return err
	}

	assertion := b.next()
	if !isRuleAssertion(assertion.text) {
		b.pos--

		return b.errorf("mapM_ of %q", assertion.text)
	}

	code, err := b.code()
	if err != nil {
		return err
	}

	if err = b.expect(tokenSymbol, ")"); err != nil {
		return err
	}

	dockerfiles, err := b.expression()
	if err != nil {
		return err
	}

	if !dockerfiles.isList {
		return b.errorf("mapM_ over what is not a list")
	}

	for _, dockerfile := range dockerfiles.list {
		b.add(assertion.text, code, dockerfile, 0)
	}

	return nil
}

// counted reads `assertChecks dockerfile (failsWith N code)`, with or
// without the parentheses and the `$`.
func (b *block) counted(assertion string) error {
	read, err := b.term()
	if err != nil {
		return err
	}

	if read.isList {
		return b.errorf("a list where a Dockerfile is expected")
	}

	if b.peek().is(tokenSymbol, "$") {
		b.pos++
	}

	parenthesised := b.peek().is(tokenSymbol, "(")
	if parenthesised {
		b.pos++
	}

	if tok := b.next(); !tok.is(tokenIdent, "failsWith") {
		b.pos--

		return b.errorf("%s with %q, which is not failsWith", assertion, tok.text)
	}

	times := b.next()
	if times.kind != tokenNumber {
		b.pos--

		return b.errorf("failsWith %q", times.text)
	}

	count, err := strconv.Atoi(times.text)
	if err != nil {
		return b.errorf("failsWith %s: %v", times.text, err)
	}

	code, err := b.code()
	if err != nil {
		return err
	}

	if parenthesised {
		if err := b.expect(tokenSymbol, ")"); err != nil {
			return err
		}
	}

	b.add(assertion, code, read.text, count)

	return nil
}

func isRuleAssertion(name string) bool {
	return name == assertCatches || name == assertCatchesNot ||
		name == assertOnBuildCatches || name == assertOnBuildCatchesNot
}

// add adds the case an assertion makes, as Helpers.hs defines the assertion:
// ruleCatches is at least one failure, each on a line; onBuildRuleCatches
// exactly one, on a line, with every RUN wrapped in ONBUILD; the two Not are
// none; failsWith is the count given.
func (b *block) add(assertion, code, dockerfile string, count int) {
	testCase := TestCase{
		Name:       b.name,
		RuleCode:   code,
		Dockerfile: dockerfile,
		Config:     b.config,
		OnBuild: assertion == assertOnBuildCatches || assertion == assertOnBuildCatchesNot ||
			assertion == assertOnBuildChecks,
	}

	switch assertion {
	case assertCatches:
		testCase.ShouldFail = true
	case assertOnBuildCatches:
		testCase.Exactly, testCase.Count, testCase.OnLines = true, 1, true
	case assertChecks, assertOnBuildChecks:
		testCase.Exactly, testCase.Count = true, count
	}

	b.cases = append(b.cases, testCase)
}

// configuration reads what `?config` is bound to: `def`, or `def { field =
// value, … }` with the fields the rules read.
func (b *block) configuration() (*HadolintConfig, error) {
	if err := b.expect(tokenIdent, "def"); err != nil {
		return nil, err
	}

	if !b.peek().is(tokenSymbol, "{") {
		return nil, nil //nolint:nilnil // def is no configuration, which is not an error
	}

	b.pos++

	config := &HadolintConfig{LabelSchema: map[string]string{}}

	for !b.peek().is(tokenSymbol, "}") {
		field := b.next()

		if err := b.expect(tokenSymbol, "="); err != nil {
			return nil, err
		}

		switch field.text {
		case "labelSchema":
			if err := b.labelSchema(config); err != nil {
				return nil, err
			}
		case "strictLabels":
			switch tok := b.next(); tok.text {
			case "True":
				config.StrictLabels = true
			case "False":
			default:
				return nil, b.errorf("strictLabels = %q", tok.text)
			}
		case "allowedRegistries":
			registries, err := b.term()
			if err != nil {
				return nil, err
			}

			if !registries.isList {
				return nil, b.errorf("allowedRegistries is not a list")
			}

			config.AllowedRegistries = registries.list
		default:
			return nil, b.errorf("configuration field %q", field.text)
		}

		if b.peek().is(tokenSymbol, ",") {
			b.pos++
		}
	}

	b.pos++

	return config, nil
}

// labelSchema reads `Map.fromList [("label", Type), …]`, the type with or
// without its module.
func (b *block) labelSchema(config *HadolintConfig) error {
	if err := b.expect(tokenIdent, "Map.fromList"); err != nil {
		return err
	}

	if err := b.expect(tokenSymbol, "["); err != nil {
		return err
	}

	for !b.peek().is(tokenSymbol, "]") {
		if err := b.expect(tokenSymbol, "("); err != nil {
			return err
		}

		label := b.next()
		if label.kind != tokenString {
			return b.errorf("label %q", label.text)
		}

		if err := b.expect(tokenSymbol, ","); err != nil {
			return err
		}

		labelType := b.next()
		if labelType.kind != tokenIdent {
			return b.errorf("label type %q", labelType.text)
		}

		config.LabelSchema[label.text] = strings.TrimPrefix(labelType.text, "Rule.")

		if err := b.expect(tokenSymbol, ")"); err != nil {
			return err
		}

		if b.peek().is(tokenSymbol, ",") {
			b.pos++
		}
	}

	b.pos++

	return nil
}

// readSpec reads every `it` of a spec: its cases, in the order written. A
// `let` outside the blocks binds for the blocks below it: a name for the
// rest of the spec, and `?config` for the blocks at its depth or deeper,
// until the indentation comes back above it.
func readSpec(src string) ([]TestCase, error) {
	tokens, err := lex(src)
	if err != nil {
		return nil, err
	}

	type scoped struct {
		col    int
		config *HadolintConfig
	}

	var (
		cases  []TestCase
		scopes []scoped
	)

	names := map[string]value{}
	read, written := 0, 0

	for pos, tok := range tokens {
		if tok.is(tokenIdent, "it") && pos+1 < len(tokens) && tokens[pos+1].kind == tokenString {
			written++
		}
	}

	for pos := 0; pos < len(tokens); pos++ {
		tok := tokens[pos]
		if !tok.startsLine {
			continue
		}

		for len(scopes) > 0 && scopes[len(scopes)-1].col > tok.col {
			scopes = scopes[:len(scopes)-1]
		}

		isLet := tok.is(tokenIdent, "let")
		isIt := tok.is(tokenIdent, "it") && pos+1 < len(tokens) && tokens[pos+1].kind == tokenString

		if !isLet && !isIt {
			continue
		}

		end := pos + 1
		for end < len(tokens) && (!tokens[end].startsLine || tokens[end].col > tok.col) {
			end++
		}

		within := &block{env: names}
		if len(scopes) > 0 {
			within.config = scopes[len(scopes)-1].config
		}

		if isLet {
			within.name, within.tokens = "let", tokens[pos:end]
			if err := within.read(); err != nil {
				return nil, err
			}

			if within.configured {
				for len(scopes) > 0 && scopes[len(scopes)-1].col >= tok.col {
					scopes = scopes[:len(scopes)-1]
				}

				scopes = append(scopes, scoped{col: tok.col, config: within.config})
			}
		} else {
			// What a block binds is its own.
			within.env = maps.Clone(names)
			within.name, within.tokens = tokens[pos+1].text, tokens[pos+len("it"):end]

			if err := within.expect(tokenSymbol, "$"); err != nil {
				return nil, err
			}

			if err := within.read(); err != nil {
				return nil, err
			}

			if len(within.cases) == 0 {
				return nil, fmt.Errorf("%w: it %q, line %d: no assertion", ErrSpec, within.name, tok.line)
			}

			cases = append(cases, within.cases...)
			read++
		}

		pos = end - 1
	}

	// An `it` that does not start its line is inside something this walk
	// does not enter.
	if read != written {
		return nil, fmt.Errorf("%w: %d `it` blocks written, %d read", ErrSpec, written, read)
	}

	return cases, nil
}
