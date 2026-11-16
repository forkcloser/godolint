// Package shell provides shell script parsing and shellcheck integration for
// analyzing and validating the commands of RUN instructions.
//
// A comment that says "Ported from" names the hadolint function the code is
// compared with, call for call: it is the map for that comparison, and is
// kept true.
package shell

import (
	"fmt"
	"slices"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// CmdPart represents a part of a command (argument or flag).
// Ported from Hadolint.Shell.CmdPart.
type CmdPart struct {
	Arg string // The argument text
	ID  int    // Position/ID for tracking which args belong to which flags
}

// Command represents a parsed shell command.
// Ported from Hadolint.Shell.Command.
type Command struct {
	Name      string    // Command name (e.g., "apt-get")
	Arguments []CmdPart // All arguments including flags
	Flags     []CmdPart // Extracted flags only
}

// ParsedShell represents a parsed shell script.
// Ported from Hadolint.Shell.ParsedShell.
type ParsedShell struct {
	Original        string    // Original script text
	PresentCommands []Command // Extracted commands
}

// ParseShell parses a shell script and extracts commands.
// Ported from Hadolint.Shell.parseShell.
func ParseShell(script string) (*ParsedShell, error) {
	// Add shebang to help parser
	fullScript := "#!/bin/bash\n" + script

	r := strings.NewReader(fullScript)

	file, err := syntax.NewParser().Parse(r, "")
	if err != nil {
		return nil, fmt.Errorf("failed to parse shell script: %w", err)
	}

	commands := extractCommands(file)

	return &ParsedShell{
		Original:        script,
		PresentCommands: commands,
	}, nil
}

// extractCommands walks the AST and extracts all commands.
func extractCommands(file *syntax.File) []Command {
	var commands []Command

	syntax.Walk(file, func(node syntax.Node) bool {
		switch typed := node.(type) {
		case *syntax.CallExpr:
			if cmd := extractCommand(typed); cmd != nil {
				commands = append(commands, *cmd)
			}
		case *syntax.DeclClause:
			// export, declare, local, readonly and typeset are commands as
			// any other; the parser gives them a node of their own.
			commands = append(commands, declaration(typed))
		}

		return true
	})

	return commands
}

// declaration is the command a declaration is: its name, and each thing it
// declares as an argument.
func declaration(decl *syntax.DeclClause) Command {
	args := make([]CmdPart, len(decl.Args))

	for idx, assign := range decl.Args {
		var text strings.Builder

		if assign.Name != nil {
			text.WriteString(assign.Name.Value)
		}

		if !assign.Naked {
			text.WriteString("=")
		}

		if assign.Value != nil {
			text.WriteString(wordToString(assign.Value))
		}

		args[idx] = CmdPart{Arg: text.String(), ID: idx}
	}

	return Command{Name: decl.Variant.Value, Arguments: args, Flags: extractFlags(args)}
}

// isWrapper reports whether name runs the command its first argument names:
// busybox wget is wget.
func isWrapper(name string) bool {
	switch name {
	case "busybox", "builtin", "command", "exec", "run":
		return true
	default:
		return false
	}
}

// extractCommand extracts a Command from a CallExpr node.
func extractCommand(call *syntax.CallExpr) *Command {
	if len(call.Args) == 0 {
		return nil
	}

	name := wordToString(call.Args[0])
	if name == "" {
		return nil
	}

	words := call.Args[1:]

	// A wrapper followed by an option is the wrapper's own business
	// (command -v pip, busybox --list), and the command is the wrapper.
	if isWrapper(name) && len(words) > 0 {
		if wrapped := wordToString(words[0]); wrapped != "" && !strings.HasPrefix(wrapped, "-") {
			name, words = wrapped, words[1:]
		}
	}

	var allArgs []CmdPart

	for i, arg := range words {
		argStr := wordToString(arg)
		allArgs = append(allArgs, CmdPart{
			Arg: argStr,
			ID:  i,
		})
	}

	flags := extractFlags(allArgs)

	return &Command{
		Name:      name,
		Arguments: allArgs,
		Flags:     flags,
	}
}

// maskedExpansion stands in for what a word expands to when nothing of it
// can be read: an arithmetic expansion, a command in backquotes.
const maskedExpansion = "${VAR}"

// wordToString is the text of a word as hadolint's simplify reads it, which
// is what the rules look for names, versions and paths in.
//
// A quoted piece is its content. A backslash outside quotes gives the
// character after it; inside double quotes it does so before a backslash, a
// double quote, a dollar or a backquote, and stays otherwise. A variable is
// ${name}, and with an operator the name, the operator and its word:
// ${VERSION:-1.0}. A $(command) is ${ and the words of its command run
// together and }: $(cat /tmp/list) is ${cat/tmp/list}. A brace expansion,
// an extended glob and a process substitution are nothing: pkg-{a,b}-1.0
// is pkg--1.0.
func wordToString(word *syntax.Word) string {
	return wordText(word, "")
}

// wordText is wordToString with the characters a backslash escapes outside
// quotes: any, when escapable is empty, as in a command's word; those of
// escapable alone in the word of a ${name<op>word}.
func wordText(word *syntax.Word, escapable string) string {
	// The parser leaves a brace expansion as the literal text it is written
	// in. SplitBraces splits out what the shell expands: a list, and a
	// sequence of numbers or of letters. {foo..bar} is no sequence and stays
	// as it is written, where ShellCheck takes any two dots for one. It
	// rewrites the word it is given: it gets a copy, since the word belongs
	// to a tree that is being walked, and the walk does not know a brace
	// expansion.
	split := &syntax.Word{Parts: make([]syntax.WordPart, len(word.Parts))}

	for idx, part := range word.Parts {
		if lit, isLit := part.(*syntax.Lit); isLit {
			part = &syntax.Lit{Value: lit.Value}
		}

		split.Parts[idx] = part
	}

	syntax.SplitBraces(split)

	var build strings.Builder

	for _, part := range split.Parts {
		build.WriteString(partText(part, escapable))
	}

	return build.String()
}

func partText(part syntax.WordPart, escapable string) string {
	switch typart := part.(type) {
	case *syntax.Lit:
		return unescaped(typart.Value, escapable)
	case *syntax.SglQuoted:
		return typart.Value
	case *syntax.DblQuoted:
		if typart.Dollar {
			return ""
		}

		var build strings.Builder

		for _, quoted := range typart.Parts {
			if lit, isLit := quoted.(*syntax.Lit); isLit {
				build.WriteString(unescaped(lit.Value, "\\\"$`"))
			} else {
				build.WriteString(partText(quoted, escapable))
			}
		}

		return build.String()
	case *syntax.ParamExp:
		return paramString(typart)
	case *syntax.CmdSubst:
		return substitutionText(typart)
	case *syntax.ArithmExp:
		// $((…)) is not read; the old $[…] is nothing to hadolint.
		if typart.Bracket {
			return ""
		}

		return maskedExpansion
	default:
		// A brace expansion, an extended glob, a process substitution.
		return ""
	}
}

// unescaped is text without the backslashes that escape: each one before a
// character of escapable, or before any character when escapable is empty,
// goes and leaves the character. A backslash that ends a line is gone
// before a word is read: the parser joins the lines.
func unescaped(text, escapable string) string {
	if !strings.Contains(text, "\\") {
		return text
	}

	var build strings.Builder

	for idx := 0; idx < len(text); idx++ {
		if text[idx] != '\\' || idx+1 == len(text) {
			build.WriteByte(text[idx])

			continue
		}

		next := text[idx+1]

		switch {
		case escapable == "" || strings.IndexByte(escapable, next) >= 0:
			idx++

			build.WriteByte(next)
		default:
			build.WriteByte(text[idx])
		}
	}

	return build.String()
}

// paramString is a variable reference: $name and ${name} are ${name}, and
// ${name<op>word} is the name, the operator and the word's own text. Any
// other form (${#name}, ${name[0]}, ${name/a/b}) is as it is written.
func paramString(param *syntax.ParamExp) string {
	plain := param.Param != nil && !param.Excl && !param.Length && !param.Width &&
		param.Index == nil && param.Slice == nil && param.Repl == nil && param.Names == 0

	switch {
	case plain && param.Exp == nil:
		return "${" + param.Param.Value + "}"
	case plain && param.Exp.Word == nil:
		return "${" + param.Param.Value + param.Exp.Op.String() + "}"
	case plain:
		// Inside the braces a backslash escapes these and no other.
		return "${" + param.Param.Value + param.Exp.Op.String() + wordText(param.Exp.Word, "}\"$`'") + "}"
	}

	var written strings.Builder
	if err := syntax.NewPrinter().Print(&written, param); err != nil {
		return maskedExpansion
	}

	return written.String()
}

// substitutionText is a $(…) as hadolint reads it: ${, the words of each
// simple command it holds run together, and }. A pipeline or a list joined
// by && or || gives no words. A command in backquotes is not read.
func substitutionText(subst *syntax.CmdSubst) string {
	if subst.Backquotes {
		return maskedExpansion
	}

	var build strings.Builder

	build.WriteString("${")

	for _, stmt := range subst.Stmts {
		// A command negated or sent to the background gives no words either.
		if stmt.Negated || stmt.Background {
			continue
		}

		if call, isCall := stmt.Cmd.(*syntax.CallExpr); isCall {
			for _, arg := range call.Args {
				build.WriteString(wordToString(arg))
			}
		}
	}

	build.WriteString("}")

	return build.String()
}

// extractFlags extracts flag arguments from a list of arguments.
// Ported from Hadolint.Shell.getAllFlags.
func extractFlags(args []CmdPart) []CmdPart {
	var flags []CmdPart

	for _, arg := range args {
		if arg.Arg == "--" || arg.Arg == "-" {
			continue
		}

		// Long flags: --flag or --flag=value
		if after, ok := strings.CutPrefix(arg.Arg, "--"); ok {
			flagName := after
			if idx := strings.IndexByte(flagName, '='); idx != -1 {
				flagName = flagName[:idx]
			}

			flags = append(flags, CmdPart{
				Arg: flagName,
				ID:  arg.ID,
			})

			continue
		}

		// Short flags: -abc becomes three flags: a, b, c
		if after, ok := strings.CutPrefix(arg.Arg, "-"); ok {
			flagChars := after
			for _, ch := range flagChars {
				flags = append(flags, CmdPart{
					Arg: string(ch),
					ID:  arg.ID,
				})
			}
		}
	}

	return flags
}

// FindCommandNames returns all command names in the parsed shell.
// Ported from Hadolint.Shell.findCommandNames.
func FindCommandNames(ps *ParsedShell) []string {
	names := make([]string, len(ps.PresentCommands))
	for i, cmd := range ps.PresentCommands {
		names[i] = cmd.Name
	}

	return names
}

// CmdHasArgs checks if a command has a specific name and contains specific arguments.
// Ported from Hadolint.Shell.cmdHasArgs.
func CmdHasArgs(expectedName string, expectedArgs []string, cmd Command) bool {
	if cmd.Name != expectedName {
		return false
	}

	for _, expected := range expectedArgs {
		for _, arg := range cmd.Arguments {
			if arg.Arg == expected {
				return true
			}
		}
	}

	return false
}

// CmdHasAllArgs reports whether cmd is expectedName and carries every one of
// expectedArgs, where CmdHasArgs asks for any of them.
// Ported from Hadolint.Shell.cmdHasAllArgs.
func CmdHasAllArgs(expectedName string, expectedArgs []string, cmd Command) bool {
	if cmd.Name != expectedName {
		return false
	}

	for _, expected := range expectedArgs {
		if !slices.ContainsFunc(cmd.Arguments, func(arg CmdPart) bool { return arg.Arg == expected }) {
			return false
		}
	}

	return true
}

// IsRecursiveRemoval reports whether cmd is an rm that recursively removes
// one of paths, however its flags are written (-rf, -fr, -r -f, -R,
// --recursive). A path is matched as written: a glob is not expanded.
// Ported from Hadolint.Shell.isRecursiveRemoval.
func IsRecursiveRemoval(paths []string, cmd Command) bool {
	if cmd.Name != "rm" || !HasAnyFlag([]string{"r", "R", "recursive"}, cmd) {
		return false
	}

	return slices.ContainsFunc(GetArgsNoFlags(cmd), func(arg string) bool {
		return slices.Contains(paths, arg)
	})
}

// Basename is the last path segment of a shell's name, so /usr/bin/bash,
// /bin/bash and bash are one shell.
// Ported from Hadolint.Shell.shellBasename.
func Basename(name string) string {
	return name[strings.LastIndex(name, "/")+1:]
}

// DropFlagArg returns cmd without the value arguments of the given flags: the
// argument after each such flag, unless the flag carries its value inline
// (--flag=value). The flags themselves stay.
// Ported from Hadolint.Shell.dropFlagArg.
func DropFlagArg(flagsToDrop []string, cmd Command) Command {
	drop := make(map[int]bool)

	for _, flag := range cmd.Flags {
		if !slices.Contains(flagsToDrop, flag.Arg) || hasInlineValue(cmd, flag.ID) {
			continue
		}

		if id, ok := valueID(cmd, flag.ID); ok {
			drop[id] = true
		}
	}

	kept := make([]CmdPart, 0, len(cmd.Arguments))

	for _, arg := range cmd.Arguments {
		if !drop[arg.ID] {
			kept = append(kept, arg)
		}
	}

	return Command{Name: cmd.Name, Arguments: kept, Flags: cmd.Flags}
}

// hasInlineValue reports whether the argument with id carries an "=".
func hasInlineValue(cmd Command, id int) bool {
	for _, arg := range cmd.Arguments {
		if arg.ID == id && strings.Contains(arg.Arg, "=") {
			return true
		}
	}

	return false
}

// valueID is the id of the argument following the one with flagID, the
// smallest id above it, as Hadolint.Shell.getValueId finds it.
func valueID(cmd Command, flagID int) (int, bool) {
	best, found := 0, false

	for _, arg := range cmd.Arguments {
		if arg.ID > flagID && (!found || arg.ID < best) {
			best, found = arg.ID, true
		}
	}

	return best, found
}

// HasFlag checks if a command has a specific flag.
// Ported from Hadolint.Shell.hasFlag.
func HasFlag(flag string, cmd Command) bool {
	for _, f := range cmd.Flags {
		if f.Arg == flag {
			return true
		}
	}

	return false
}

// GetArgs returns all argument strings (including flags).
// Ported from Hadolint.Shell.getArgs.
func GetArgs(cmd Command) []string {
	args := make([]string, len(cmd.Arguments))
	for i, arg := range cmd.Arguments {
		args[i] = arg.Arg
	}

	return args
}

// GetArgsNoFlags returns arguments that are not flags or flag values.
// Ported from Hadolint.Shell.getArgsNoFlags.
func GetArgsNoFlags(cmd Command) []string {
	flagIDs := make(map[int]bool)
	for _, flag := range cmd.Flags {
		flagIDs[flag.ID] = true
	}

	var result []string

	for _, arg := range cmd.Arguments {
		if !flagIDs[arg.ID] {
			result = append(result, arg.Arg)
		}
	}

	return result
}

// CountCommands returns the number of commands in a parsed shell script.
// Used by DL3059 to detect chained commands.
func CountCommands(ps *ParsedShell) int {
	return len(ps.PresentCommands)
}

// HasAnyFlag checks if a command has any of the specified flags.
// Ported from Hadolint.Shell.hasAnyFlag.
func HasAnyFlag(flags []string, cmd Command) bool {
	for _, flag := range flags {
		if HasFlag(flag, cmd) {
			return true
		}
	}

	return false
}

// CountFlag counts how many times a specific flag appears.
// Ported from Hadolint.Shell.countFlag.
func CountFlag(flag string, cmd Command) int {
	count := 0

	for _, f := range cmd.Flags {
		if f.Arg == flag {
			count++
		}
	}

	return count
}

// HasArg checks if a command has a specific argument.
// Ported from Hadolint.Shell.hasArg.
func HasArg(arg string, cmd Command) bool {
	for _, a := range cmd.Arguments {
		if a.Arg == arg {
			return true
		}
	}

	return false
}

// UsingProgram checks if any command uses a specific program.
// Ported from Hadolint.Shell.usingProgram.
func UsingProgram(program string, ps *ParsedShell) bool {
	for _, cmd := range ps.PresentCommands {
		if cmd.Name == program {
			return true
		}
	}

	return false
}

// GetFlagArg is each value flag is given: the argument after --flag, or
// after -f alone or last in a cluster (-xf value), and what follows the =
// of --flag=value. A value glued to a short flag (-fvalue) is not read:
// nothing tells it from a cluster.
func GetFlagArg(flag string, cmd Command) []string {
	var values []string

	for idx, arg := range cmd.Arguments {
		text := arg.Arg

		if len(flag) > 1 {
			if value, glued := strings.CutPrefix(text, "--"+flag+"="); glued {
				values = append(values, value)

				continue
			}
		}

		named := text == "--"+flag
		if len(flag) == 1 {
			named = strings.HasPrefix(text, "-") && !strings.HasPrefix(text, "--") && strings.HasSuffix(text[1:], flag)
		}

		if named && idx+1 < len(cmd.Arguments) {
			values = append(values, cmd.Arguments[idx+1].Arg)
		}
	}

	return values
}

// HasPipes checks if a parsed shell script contains pipe operators.
// Ported from Hadolint.Shell.hasPipes / findPipes.
func HasPipes(ps *ParsedShell) bool {
	// Re-parse to access the full AST (not just extracted commands)
	fullScript := "#!/bin/bash\n" + ps.Original
	r := strings.NewReader(fullScript)

	file, err := syntax.NewParser().Parse(r, "")
	if err != nil {
		return false
	}

	hasPipe := false

	syntax.Walk(file, func(node syntax.Node) bool {
		if bin, ok := node.(*syntax.BinaryCmd); ok {
			// Op_Pipe is |, Op_PipeAll is |&
			if bin.Op == syntax.Pipe || bin.Op == syntax.PipeAll {
				hasPipe = true

				return false // Stop walking
			}
		}

		return true
	})

	return hasPipe
}

// IsPipInstall checks if a command is a pip install command.
// Ported from Hadolint.Shell.isPipInstall.
func IsPipInstall(cmd Command) bool {
	// pip, pip3, pipx and whatever else pip goes by; pipenv is a different
	// tool. install may follow flags.
	if strings.HasPrefix(cmd.Name, "pip") && !strings.HasPrefix(cmd.Name, "pipenv") {
		return slices.Contains(GetArgs(cmd), "install")
	}

	if strings.HasPrefix(cmd.Name, "python") {
		return runsPipModuleInstall(GetArgs(cmd))
	}

	return false
}

// runsPipModuleInstall reports whether a python command runs pip's install
// through -m. Anything may sit between the module and the subcommand:
// python -m pip --quiet install is still an install.
func runsPipModuleInstall(args []string) bool {
	for idx := range len(args) - 1 {
		if args[idx] != "-m" || args[idx+1] != "pip" {
			continue
		}

		if slices.Contains(args[idx+2:], "install") {
			return true
		}
	}

	return false
}
