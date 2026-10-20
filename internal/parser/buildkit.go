package parser

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"

	"github.com/forkcloser/godolint/internal/syntax"
)

// onBuildPrefix is the keyword ONBUILD lines start with, trailing space
// included: buildkit hands the converter the whole line, and what follows the
// keyword is the instruction ONBUILD wraps.
const onBuildPrefix = "ONBUILD "

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
	// Parse using buildkit
	result, err := parser.Parse(bytes.NewReader(dockerfile))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Dockerfile: %w", err)
	}

	// Convert buildkit AST to our AST format
	var instructions []syntax.InstructionPos

	for _, child := range result.AST.Children {
		// First, add any preceding comments as Comment instructions
		for i, commentText := range child.PrevComment {
			// Comments come before the instruction, calculate line number
			// PrevComment is in order, with the first comment being furthest back
			commentLine := child.StartLine - (len(child.PrevComment) - i)

			// Strip leading/trailing whitespace and # prefix
			text := strings.TrimSpace(commentText)
			text = strings.TrimPrefix(text, "#")
			text = strings.TrimSpace(text)

			instructions = append(instructions, syntax.InstructionPos{
				Instruction: &syntax.Comment{Text: text},
				LineNumber:  commentLine,
			})
		}

		// An instruction the converter refuses (one it does not know, or
		// one without what it requires, a FROM with no image) is no
		// Dockerfile: hadolint refuses it at parse, and so does a build.
		// Skipping it lint-passed what neither accepts.
		instr, err := convertNode(child)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", child.StartLine, err)
		}

		if instr != nil {
			instructions = append(instructions, syntax.InstructionPos{
				Instruction: instr,
				LineNumber:  child.StartLine,
			})
		}
	}

	return instructions, nil
}

// convertNode converts a buildkit AST node to our Instruction type.
func convertNode(node *parser.Node) (syntax.Instruction, error) {
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
		return convertEnv(node)
	case "label":
		return convertLabel(node)
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
		return convertOnBuild(node)
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
			// We're in the middle of a quoted string
			current += " " + val
			if strings.HasSuffix(val, "\"") {
				// End of quoted string
				values = append(values, current)
				current = ""
				inQuote = false
			}
		case strings.HasPrefix(val, "\"") && !strings.HasSuffix(val, "\""):
			// Start of a quoted string that spans multiple nodes
			current = val
			inQuote = true
		default:
			// Normal value
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

	baseImage := syntax.BaseImage{
		Image: values[0],
	}

	// Extract --platform flag if present
	for _, flag := range node.Flags {
		if after, ok := strings.CutPrefix(flag, "--platform="); ok {
			platform := after
			baseImage.Platform = &platform
		}
	}

	// Check for AS alias (last token after AS keyword)
	for i := 1; i < len(values); i++ {
		if values[i] == "AS" || values[i] == "as" {
			if i+1 < len(values) {
				alias := values[i+1]
				baseImage.Alias = &alias
			}

			break
		}
	}

	// Parse image reference (image:tag@digest)
	// The buildkit parser gives us the raw string, we need to parse it
	imagePart := values[0]

	// Check for digest (@sha256:...)
	if atIdx := indexOf(imagePart, '@'); atIdx != -1 {
		digest := imagePart[atIdx+1:]
		baseImage.Digest = &digest
		imagePart = imagePart[:atIdx]
	}

	// Check for tag (:tag)
	if colonIdx := indexOf(imagePart, ':'); colonIdx != -1 {
		tag := imagePart[colonIdx+1:]
		baseImage.Tag = &tag
		baseImage.Image = imagePart[:colonIdx]
	}

	return &syntax.From{
		Image: baseImage,
	}, nil
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
	// Collect all parts of the RUN command
	command := ""

	for n := node.Next; n != nil; n = n.Next {
		if command != "" {
			command += " "
		}

		command += n.Value
	}

	return &syntax.Run{
		Command: command,
		Flags:   node.Flags,
		IsJSON:  node.Attributes != nil && node.Attributes["json"],
	}, nil
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

	// Check for --from flag
	// Buildkit parser includes flags in node.Flags
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

//nolint:unparam // Uniform signature with other converters for consistent error handling
func convertEnv(node *parser.Node) (*syntax.Env, error) {
	values := collectValues(node)

	var pairs []syntax.EnvPair

	// ENV has two syntaxes:
	// 1. "ENV key=value key2=value2" - buildkit tokenizes as: key, value, =, key2, value2, =
	// 2. "ENV key value" - buildkit tokenizes as: key, value
	for idx := 0; idx < len(values); {
		switch {
		case idx+2 < len(values) && values[idx+2] == "=":
			// Pattern: key value = (key=value syntax)
			pairs = append(pairs, syntax.EnvPair{
				Key:   values[idx],
				Value: unquote(values[idx+1]),
			})
			idx += 3 // Skip key, value, =
		case idx+1 < len(values):
			// Whitespace syntax: key value
			pairs = append(pairs, syntax.EnvPair{
				Key:   values[idx],
				Value: unquote(values[idx+1]),
			})
			idx += 2
		default:
			idx++
		}
	}

	return &syntax.Env{
		Pairs: pairs,
	}, nil
}

// unquote removes surrounding double quotes from a string if present.
func unquote(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}

	return s
}

//nolint:unparam // Uniform signature with other converters for consistent error handling
func convertLabel(node *parser.Node) (*syntax.Label, error) {
	values := collectValues(node)

	var pairs []syntax.LabelPair

	// Buildkit returns tokens as: key, value, "=", key2, value2, "=", ...
	for i := 0; i+2 < len(values); i += 3 {
		key := values[i]
		value := stripQuotes(values[i+1])
		// values[i+2] is "=", skip it

		pairs = append(pairs, syntax.LabelPair{
			Key:   key,
			Value: value,
		})
	}

	return &syntax.Label{
		Pairs: pairs,
	}, nil
}

// stripQuotes removes surrounding quotes from a string if present.
func stripQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}

	return s
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
	// Check if using JSON notation (exec form)
	isJSON := node.Attributes != nil && node.Attributes["json"]

	return &syntax.Cmd{
		Arguments: collectValues(node),
		IsJSON:    isJSON,
	}, nil
}

//nolint:unparam // Uniform signature with other converters for consistent error handling
func convertEntrypoint(node *parser.Node) (*syntax.Entrypoint, error) {
	// Check if using JSON notation (exec form)
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

	// Parse ARG name or ARG name=value
	arg := &syntax.Arg{}

	// Check for = separator
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

func convertOnBuild(node *parser.Node) (*syntax.OnBuild, error) {
	// ONBUILD wraps another instruction
	// buildkit stores the full instruction in Original field (e.g., "ONBUILD FROM debian")
	// We need to parse the inner instruction from the Original string
	original := strings.TrimSpace(node.Original)

	// Remove the ONBUILD keyword, leaving the instruction it wraps.
	if !strings.HasPrefix(strings.ToUpper(original), onBuildPrefix) {
		return nil, ErrInvalidOnBuild
	}

	innerText := strings.TrimSpace(original[len(onBuildPrefix):])
	if innerText == "" {
		return nil, ErrOnBuildMissingInstruction
	}

	// Parse the inner instruction as a complete Dockerfile
	innerResult, err := parser.Parse(bytes.NewReader([]byte(innerText)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ONBUILD inner instruction: %w", err)
	}

	if len(innerResult.AST.Children) == 0 {
		return nil, ErrOnBuildNoInnerInstruction
	}

	// Convert the first (and only) child instruction
	inner, err := convertNode(innerResult.AST.Children[0])
	if err != nil {
		return nil, fmt.Errorf("failed to convert ONBUILD inner instruction: %w", err)
	}

	return &syntax.OnBuild{
		Inner: inner,
	}, nil
}
