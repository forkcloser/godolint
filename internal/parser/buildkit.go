package parser

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/moby/buildkit/frontend/dockerfile/parser"

	"github.com/forkcloser/godolint/internal/syntax"
)

const (
	// onBuildKeyword is the keyword ONBUILD lines start with: buildkit hands
	// the converter the whole line, and what follows the keyword is the
	// instruction ONBUILD wraps.
	onBuildKeyword = "ONBUILD"
	// instructionBlanks are the characters buildkit splits an instruction
	// from its arguments by: a tab after ONBUILD is a blank as a space is.
	instructionBlanks = "\t\v\f\r "
)

// Static sentinel errors for instruction conversion failures, so callers can
// match them with errors.Is; detail is attached by wrapping where needed.
var (
	// ErrUnknownInstruction reports an instruction the converter does not know about.
	ErrUnknownInstruction = errors.New("unknown instruction")
	// ErrFromMissingImage reports a FROM without an image reference.
	ErrFromMissingImage = errors.New("FROM missing image")
	// ErrCopyMissingArgs reports a COPY without both a source and a destination.
	ErrCopyMissingArgs = errors.New("COPY requires at least source and destination")
	// ErrAddMissingArgs reports an ADD without both a source and a destination.
	ErrAddMissingArgs = errors.New("ADD requires at least source and destination")
	// ErrArgMissingName reports an ARG without a name.
	ErrArgMissingName = errors.New("ARG missing name")
	// ErrInvalidOnBuild reports an ONBUILD line that does not start with the keyword.
	ErrInvalidOnBuild = errors.New("invalid ONBUILD instruction")
	// ErrOnBuildMissingInstruction reports an ONBUILD without a wrapped instruction.
	ErrOnBuildMissingInstruction = errors.New("ONBUILD missing instruction")
	// ErrOnBuildNoInnerInstruction reports an ONBUILD whose body parses to nothing.
	ErrOnBuildNoInnerInstruction = errors.New("ONBUILD has no inner instruction")
)

// BuildkitParser implements Parser using moby/buildkit's Dockerfile parser.
type BuildkitParser struct{}

// NewBuildkitParser creates a new buildkit-based parser.
func NewBuildkitParser() *BuildkitParser {
	return &BuildkitParser{}
}

// Parse parses a Dockerfile using moby/buildkit and converts to our AST format.
func (*BuildkitParser) Parse(dockerfile []byte) ([]syntax.InstructionPos, error) {
	result, err := parser.Parse(bytes.NewReader(dockerfile))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Dockerfile: %w", err)
	}

	var instructions []syntax.InstructionPos

	comments := commentLines(dockerfile, result.AST.Children)

	for _, child := range result.AST.Children {
		// The comments that stand before the instruction, each on its line.
		for len(comments) > 0 && comments[0].LineNumber < child.StartLine {
			instructions = append(instructions, comments[0])
			comments = comments[1:]
		}

		// An instruction the converter refuses (one it does not know, or
		// one without what it requires, a FROM with no image) is no
		// Dockerfile: hadolint refuses it at parse, and so does a build.
		// Skipping it lint-passed what neither accepts.
		instr, err := convertNode(child, result.EscapeToken)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", child.StartLine, err)
		}

		if instr != nil {
			instructions = append(instructions, syntax.InstructionPos{
				Instruction: instr,
				LineNumber:  child.StartLine,
				EndLine:     child.EndLine,
			})
		}
	}

	// The comments after the last instruction are comments too.
	return append(instructions, comments...), nil
}

// commentLines reads the comments of a Dockerfile from its text, each with
// the line it stands on: a line whose first character that is no blank is a
// `#`, outside every instruction. buildkit keeps the comments above an
// instruction without their lines, forgets those above an empty `#` line and
// those after the last instruction, and an ignore pragma covers the line
// after its own and no other, so the line has to be the true one. A `#` line
// among the lines of an instruction, between two continued lines or in a
// heredoc, is that instruction's and no comment.
func commentLines(dockerfile []byte, nodes []*parser.Node) []syntax.InstructionPos {
	var comments []syntax.InstructionPos

	within := 0

	for idx, line := range strings.Split(string(dockerfile), "\n") {
		number := idx + 1

		for within < len(nodes) && nodes[within].EndLine < number {
			within++
		}

		if within < len(nodes) && nodes[within].StartLine <= number {
			continue
		}

		if text, isComment := strings.CutPrefix(strings.TrimSpace(line), "#"); isComment {
			comments = append(comments, syntax.InstructionPos{
				Instruction: &syntax.Comment{Text: strings.TrimSpace(text)},
				LineNumber:  number,
				EndLine:     number,
			})
		}
	}

	return comments
}

// convertNode converts a buildkit AST node to our Instruction type. escape
// is the Dockerfile's escape character, a backslash unless its first lines
// say otherwise (# escape=`).
func convertNode(node *parser.Node, escape rune) (syntax.Instruction, error) {
	switch strings.ToLower(node.Value) {
	case "from":
		return convertFrom(node)
	case "run":
		return convertRun(node)
	case "copy":
		return convertCopy(node)
	case "add":
		return convertAdd(node)
	case "env":
		return convertEnv(node, escape)
	case "label":
		return convertLabel(node, escape)
	case "workdir":
		return convertWorkdir(node)
	case "user":
		return convertUser(node)
	case "expose":
		return convertExpose(node)
	case "volume":
		return convertVolume(node)
	case "cmd":
		return convertCmd(node)
	case "entrypoint":
		return convertEntrypoint(node)
	case "healthcheck":
		return convertHealthcheck(node)
	case "maintainer":
		return convertMaintainer(node)
	case "arg":
		return convertArg(node)
	case "stopsignal":
		return convertStopSignal(node)
	case "shell":
		return convertShell(node)
	case "onbuild":
		return convertOnBuild(node, escape)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownInstruction, node.Value)
	}
}

// Helper to get next child value.
func nextValue(node *parser.Node) string {
	if node.Next != nil {
		return node.Next.Value
	}

	return ""
}

// Helper to collect all remaining values.
func collectValues(node *parser.Node) []string {
	var values []string

	var current string

	inQuote := false

	for n := node.Next; n != nil; n = n.Next {
		val := n.Value

		switch {
		case inQuote:
			current += " " + val
			if strings.HasSuffix(val, "\"") {
				values = append(values, current)
				current = ""
				inQuote = false
			}
		case strings.HasPrefix(val, "\"") && !strings.HasSuffix(val, "\""):
			current = val
			inQuote = true
		default:
			values = append(values, val)
		}
	}

	// Handle unclosed quote (shouldn't happen with valid Dockerfile)
	if inQuote && current != "" {
		values = append(values, current)
	}

	return values
}

func convertFrom(node *parser.Node) (*syntax.From, error) {
	if node.Next == nil {
		return nil, ErrFromMissingImage
	}

	// Collect all values to handle: image:tag@digest AS alias
	values := collectValues(node)
	if len(values) == 0 {
		return nil, ErrFromMissingImage
	}

	image, tag, digest := splitReference(values[0])
	baseImage := syntax.BaseImage{Image: image, Tag: tag, Digest: digest}

	for _, flag := range node.Flags {
		if after, ok := strings.CutPrefix(flag, "--platform="); ok {
			platform := after
			baseImage.Platform = &platform
		}
	}

	for i := 1; i < len(values); i++ {
		if values[i] == "AS" || values[i] == "as" {
			if i+1 < len(values) {
				// A build keeps a stage's name in lower case.
				alias := strings.ToLower(values[i+1])
				baseImage.Alias = &alias
			}

			break
		}
	}

	return &syntax.From{
		Image: baseImage,
	}, nil
}

// splitReference splits the image of a FROM into its name (registry and port
// included), its tag and its digest. The digest follows the first @; the tag
// follows the last colon after the last slash, since a colon before a slash
// is a registry's port. A ${...} expansion is opaque: the colon of its :-, a
// slash or an @ in its default split nothing, and one left open runs to the
// end.
func splitReference(reference string) (name string, tag, digest *string) {
	digestAt, tagAt, depth := -1, -1, 0

	for i := 0; i < len(reference) && digestAt < 0; i++ {
		switch char := reference[i]; {
		case depth > 0 && char == '{':
			depth++
		case depth > 0 && char == '}':
			depth--
		case depth > 0:
		case char == '{' && i > 0 && reference[i-1] == '$':
			depth = 1
		case char == '@':
			digestAt = i
		case char == '/':
			tagAt = -1
		case char == ':':
			tagAt = i
		}
	}

	name = reference

	if digestAt >= 0 {
		rest := name[digestAt+1:]
		digest, name = &rest, name[:digestAt]
	}

	if tagAt >= 0 {
		rest := name[tagAt+1:]
		tag, name = &rest, name[:tagAt]
	}

	return name, tag, digest
}

// Helper to find index of character in string.
func indexOf(s string, ch rune) int {
	for i, c := range s {
		if c == ch {
			return i
		}
	}

	return -1
}

//nolint:unparam // Uniform signature with other converters for consistent error handling
func convertRun(node *parser.Node) (*syntax.Run, error) {
	command := ""

	for n := node.Next; n != nil; n = n.Next {
		if command != "" {
			command += " "
		}

		command += n.Value
	}

	offset := 0

	if len(node.Heredocs) > 0 {
		var below int

		command, below = heredocScript(command, node.Heredocs[0])

		// A body that is the script is the only heredoc of its instruction,
		// so it ends on the line above the instruction's last, the delimiter.
		if below >= 0 {
			offset = node.EndLine - strings.Count(node.Heredocs[0].Content, "\n") - node.StartLine + below
		}
	}

	return &syntax.Run{
		Command:      command,
		Flags:        node.Flags,
		IsJSON:       node.Attributes != nil && node.Attributes["json"],
		ScriptOffset: offset,
	}, nil
}

// shells is the interpreters whose scripts are shell scripts.
//
//nolint:gochecknoglobals // a constant table
var shells = map[string]bool{"sh": true, "bash": true, "ash": true, "dash": true, "ksh": true, "zsh": true}

// heredocScript is the shell script of a RUN written with a heredoc, the
// first one of its line. Where the heredoc is the whole command (RUN <<EOF),
// the build runs its body, and the body is the script. Where a command comes
// first (RUN cat <<EOF > /etc/motd), the body is that command's input and
// the script is what stands before the heredoc, as hadolint's parser has it.
// A body that names an interpreter of its own that is no shell (#!/usr/bin/env
// python3) is run by it, and is no shell script.
//
// With a script that is the body comes the number of blank lines the body
// opens with, which the script does not keep; with any other, -1.
func heredocScript(command string, heredoc parser.Heredoc) (script string, blank int) {
	operator := regexp.MustCompile(`\d*<<-?\s*["']?` + regexp.QuoteMeta(heredoc.Name) + `["']?`)

	span := operator.FindStringIndex(command)
	if span == nil {
		return command, -1
	}

	if strings.TrimSpace(command[:span[0]]+command[span[1]:]) != "" {
		return strings.TrimSpace(command[:span[0]]), -1
	}

	body := heredoc.Content
	if heredoc.Chomp {
		body = parser.ChompHeredocContent(body)
	}

	if interpreter, named := shebang(body); named && !shells[interpreter] {
		return "", -1
	}

	opening := body[:len(body)-len(strings.TrimLeftFunc(body, unicode.IsSpace))]

	return strings.TrimSpace(body), strings.Count(opening, "\n")
}

// shebang is the interpreter a script names on its first line, by its base
// name, with env and busybox looked through: #!/usr/bin/env -S python3 -u is
// python3, and #!/bin/busybox sh is sh.
func shebang(script string) (string, bool) {
	line, _, _ := strings.Cut(script, "\n")

	words, isShebang := strings.CutPrefix(line, "#!")
	if !isShebang {
		return "", false
	}

	for word := range strings.FieldsSeq(words) {
		name := word[strings.LastIndex(word, "/")+1:]
		if name == "env" || name == "busybox" || strings.HasPrefix(name, "-") {
			continue
		}

		return name, true
	}

	return "", true
}

func convertCopy(node *parser.Node) (*syntax.Copy, error) {
	values := collectValues(node)
	if len(values) < 2 {
		return nil, ErrCopyMissingArgs
	}

	copyInstr := &syntax.Copy{
		Source:      values[:len(values)-1],
		Destination: values[len(values)-1],
	}

	for _, flag := range node.Flags {
		if after, ok := strings.CutPrefix(flag, "--from="); ok {
			fromValue := after
			copyInstr.From = &fromValue
		}
	}

	return copyInstr, nil
}

func convertAdd(node *parser.Node) (*syntax.Add, error) {
	values := collectValues(node)
	if len(values) < 2 {
		return nil, ErrAddMissingArgs
	}

	return &syntax.Add{
		Source:      values[:len(values)-1],
		Destination: values[len(values)-1],
	}, nil
}

// keyValue is one pair of an ENV or a LABEL.
type keyValue struct {
	key, value string
}

// pairs reads the pairs of an ENV or a LABEL. buildkit hands them over three
// nodes a pair: the key and the value as they are written, quotes included,
// and what stood between them, `=` or nothing for the old `key value` form,
// whose value is the rest of the line.
func pairs(node *parser.Node, escape rune) []keyValue {
	var found []keyValue

	for key := node.Next; key != nil && key.Next != nil; {
		value := key.Next
		found = append(found, keyValue{key: pairText(key.Value, escape), value: pairText(value.Value, escape)})

		if value.Next == nil {
			break
		}

		key = value.Next.Next
	}

	return found
}

// pairText is a key or a value as it is meant: the pieces written in double
// quotes, in single quotes and in none, joined, the quotes gone. A quote
// escaped inside its own kind of quotes, and a blank escaped outside any,
// stand for themselves; any other backslash is kept. So `"/opt/bin":$PATH`
// is `/opt/bin:$PATH`, and a value is never cut at a quote that closes
// before its end. The backslash is the Dockerfile's escape character,
// whichever it is.
func pairText(written string, escape rune) string {
	var (
		text  strings.Builder
		quote rune
	)

	runes := []rune(written)

	for idx := 0; idx < len(runes); idx++ {
		char := runes[idx]

		escapes := quote
		if quote == 0 {
			escapes = ' '
		}

		switch {
		case char == escape && idx+1 < len(runes) && (runes[idx+1] == escapes || quote == 0 && runes[idx+1] == '\t'):
			idx++
			text.WriteRune(runes[idx])
		case quote == 0 && (char == '"' || char == '\''):
			quote = char
		case char == quote:
			quote = 0
		default:
			text.WriteRune(char)
		}
	}

	return text.String()
}

//nolint:unparam // Uniform signature with other converters for consistent error handling
func convertEnv(node *parser.Node, escape rune) (*syntax.Env, error) {
	env := &syntax.Env{}

	for _, pair := range pairs(node, escape) {
		env.Pairs = append(env.Pairs, syntax.EnvPair{Key: pair.key, Value: pair.value})
	}

	return env, nil
}

//nolint:unparam // Uniform signature with other converters for consistent error handling
func convertLabel(node *parser.Node, escape rune) (*syntax.Label, error) {
	label := &syntax.Label{}

	for _, pair := range pairs(node, escape) {
		label.Pairs = append(label.Pairs, syntax.LabelPair{Key: pair.key, Value: pair.value})
	}

	return label, nil
}

//nolint:unparam // Uniform signature with other converters for consistent error handling
func convertWorkdir(node *parser.Node) (*syntax.Workdir, error) {
	return &syntax.Workdir{
		Directory: nextValue(node),
	}, nil
}

//nolint:unparam // Uniform signature with other converters for consistent error handling
func convertUser(node *parser.Node) (*syntax.User, error) {
	return &syntax.User{
		User: nextValue(node),
	}, nil
}

//nolint:unparam // Uniform signature with other converters for consistent error handling
func convertExpose(node *parser.Node) (*syntax.Expose, error) {
	return &syntax.Expose{
		Ports: collectValues(node),
	}, nil
}

//nolint:unparam // Uniform signature with other converters for consistent error handling
func convertVolume(node *parser.Node) (*syntax.Volume, error) {
	return &syntax.Volume{
		Volumes: collectValues(node),
	}, nil
}

//nolint:unparam // Uniform signature with other converters for consistent error handling
func convertCmd(node *parser.Node) (*syntax.Cmd, error) {
	isJSON := node.Attributes != nil && node.Attributes["json"]

	return &syntax.Cmd{
		Arguments: collectValues(node),
		IsJSON:    isJSON,
	}, nil
}

//nolint:unparam // Uniform signature with other converters for consistent error handling
func convertEntrypoint(node *parser.Node) (*syntax.Entrypoint, error) {
	isJSON := node.Attributes != nil && node.Attributes["json"]

	return &syntax.Entrypoint{
		Arguments: collectValues(node),
		IsJSON:    isJSON,
	}, nil
}

//nolint:unparam // Uniform signature with other converters for consistent error handling
func convertHealthcheck(node *parser.Node) (*syntax.Healthcheck, error) {
	command := ""

	for n := node.Next; n != nil; n = n.Next {
		if command != "" {
			command += " "
		}

		command += n.Value
	}

	return &syntax.Healthcheck{
		Command: command,
		None:    strings.EqualFold(nextValue(node), "NONE"),
		IsJSON:  node.Attributes != nil && node.Attributes["json"],
	}, nil
}

//nolint:unparam // Uniform signature with other converters for consistent error handling
func convertMaintainer(node *parser.Node) (*syntax.Maintainer, error) {
	return &syntax.Maintainer{
		MaintainerName: nextValue(node),
	}, nil
}

func convertArg(node *parser.Node) (*syntax.Arg, error) {
	value := nextValue(node)
	if value == "" {
		return nil, ErrArgMissingName
	}

	arg := &syntax.Arg{}

	if eqIdx := indexOf(value, '='); eqIdx != -1 {
		arg.ArgName = value[:eqIdx]
		defaultVal := value[eqIdx+1:]
		arg.Value = &defaultVal
	} else {
		arg.ArgName = value
	}

	return arg, nil
}

//nolint:unparam // Uniform signature with other converters for consistent error handling
func convertStopSignal(node *parser.Node) (*syntax.StopSignal, error) {
	return &syntax.StopSignal{
		Signal: nextValue(node),
	}, nil
}

//nolint:unparam // Uniform signature with other converters for consistent error handling
func convertShell(node *parser.Node) (*syntax.Shell, error) {
	return &syntax.Shell{
		Arguments: collectValues(node),
	}, nil
}

func convertOnBuild(node *parser.Node, escape rune) (*syntax.OnBuild, error) {
	// ONBUILD wraps another instruction
	// buildkit stores the full instruction in Original field (e.g., "ONBUILD FROM debian")
	// We need to parse the inner instruction from the Original string
	original := strings.TrimSpace(node.Original)

	if len(original) < len(onBuildKeyword) || !strings.EqualFold(original[:len(onBuildKeyword)], onBuildKeyword) {
		return nil, ErrInvalidOnBuild
	}

	rest := original[len(onBuildKeyword):]
	if rest != "" && !strings.ContainsRune(instructionBlanks, rune(rest[0])) {
		return nil, ErrInvalidOnBuild
	}

	innerText := strings.TrimSpace(rest)
	if innerText == "" {
		return nil, ErrOnBuildMissingInstruction
	}

	// The bodies of the instruction's heredocs follow its line, each up to
	// its delimiter; Original is the line alone.
	var withBodies strings.Builder

	withBodies.WriteString(innerText)

	for _, heredoc := range node.Heredocs {
		withBodies.WriteString("\n" + heredoc.Content + heredoc.Name)
	}

	// Parse the inner instruction as a complete Dockerfile, under the escape
	// character of the one it comes from: buildkit splits the words of an
	// ENV or a LABEL by it.
	innerResult, err := parser.Parse(strings.NewReader("# escape=" + string(escape) + "\n" + withBodies.String()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ONBUILD inner instruction: %w", err)
	}

	if len(innerResult.AST.Children) == 0 {
		return nil, ErrOnBuildNoInnerInstruction
	}

	// Convert the first (and only) child instruction
	inner, err := convertNode(innerResult.AST.Children[0], escape)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ONBUILD inner instruction: %w", err)
	}

	return &syntax.OnBuild{
		Inner: inner,
	}, nil
}
