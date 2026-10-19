// Package shell provides shell script parsing and shellcheck integration for
// analyzing and validating the commands of RUN instructions.
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

	// Parse the script
	r := strings.NewReader(fullScript)

	file, err := syntax.NewParser().Parse(r, "")
	if err != nil {
		return nil, fmt.Errorf("failed to parse shell script: %w", err)
	}

	// Extract commands
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
		// Look for CallExpr nodes (simple commands)
		if call, ok := node.(*syntax.CallExpr); ok {
			if cmd := extractCommand(call); cmd != nil {
				commands = append(commands, *cmd)
			}
		}

		return true
	})

	return commands
}

// extractCommand extracts a Command from a CallExpr node.
func extractCommand(call *syntax.CallExpr) *Command {
	if len(call.Args) == 0 {
		return nil
	}

	// Get command name from first argument
	nameWord := call.Args[0]

	name := wordToString(nameWord)
	if name == "" {
		return nil
	}

	// Extract all arguments with IDs
	var allArgs []CmdPart

	for i, arg := range call.Args[1:] {
		argStr := wordToString(arg)
		allArgs = append(allArgs, CmdPart{
			Arg: argStr,
			ID:  i,
		})
	}

	// Extract flags from arguments
	flags := extractFlags(allArgs)

	return &Command{
		Name:      name,
		Arguments: allArgs,
		Flags:     flags,
	}
}

// maskedExpansion stands in for any dynamic word part (variables, command
// substitutions, arithmetic) that oversimplification cannot resolve.
const maskedExpansion = "${VAR}"

// wordToString converts a Word node to a string.
// Similar to ShellCheck.ASTLib.oversimplify.
func wordToString(word *syntax.Word) string {
	var build strings.Builder

	for _, part := range word.Parts {
		switch typart := part.(type) {
		case *syntax.Lit:
			_, _ = build.WriteString(typart.Value)
		case *syntax.SglQuoted:
			_, _ = build.WriteString(typart.Value)
		case *syntax.DblQuoted:
			// Recursively process quoted parts
			for _, qp := range typart.Parts {
				switch quoted := qp.(type) {
				case *syntax.Lit:
					_, _ = build.WriteString(quoted.Value)
				case *syntax.ParamExp:
					_, _ = build.WriteString(paramString(quoted))
				default:
					_, _ = build.WriteString(maskedExpansion)
				}
			}
		case *syntax.ParamExp:
			_, _ = build.WriteString(paramString(typart))
		case *syntax.CmdSubst:
			_, _ = build.WriteString(maskedExpansion)
		case *syntax.ArithmExp:
			_, _ = build.WriteString(maskedExpansion)
		default:
			// Other expansions simplified
			_, _ = build.WriteString(maskedExpansion)
		}
	}

	return build.String()
}

// paramString is a plain variable reference, $name or ${name}, as ${name},
// so that a rule can tell which variable a word uses, as hadolint's simplify
// lets it; a reference with an operator (${name:-x}, ${#name}, ${name[0]})
// is masked like any other expansion.
func paramString(param *syntax.ParamExp) string {
	plain := param.Param != nil && !param.Excl && !param.Length && !param.Width &&
		param.Index == nil && param.Slice == nil && param.Repl == nil && param.Exp == nil && param.Names == 0
	if !plain {
		return maskedExpansion
	}

	return "${" + param.Param.Value + "}"
}

// extractFlags extracts flag arguments from a list of arguments.
// Ported from Hadolint.Shell.getAllFlags.
func extractFlags(args []CmdPart) []CmdPart {
	var flags []CmdPart

	for _, arg := range args {
		// Skip special cases
		if arg.Arg == "--" || arg.Arg == "-" {
			continue
		}

		// Long flags: --flag or --flag=value
		if after, ok := strings.CutPrefix(arg.Arg, "--"); ok {
			flagName := after
			// Remove =value part if present
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

	// Check if any of the expected args are present
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
	// Get IDs of all flags
	flagIDs := make(map[int]bool)
	for _, flag := range cmd.Flags {
		flagIDs[flag.ID] = true
	}

	// Return arguments whose IDs are not in flag IDs
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

// GetFlagArg returns the argument values for a specific flag.
// Ported from Hadolint.Shell.getFlagArg.
func GetFlagArg(flag string, cmd Command) []string {
	var values []string

	args := cmd.Arguments

	for i, arg := range args {
		// Check for -flag value or --flag value
		if arg.Arg == "-"+flag || arg.Arg == "--"+flag {
			// Next argument is the value
			if i+1 < len(args) {
				values = append(values, args[i+1].Arg)
			}

			continue
		}

		// Check for --flag=value
		if after, ok := strings.CutPrefix(arg.Arg, "--"+flag+"="); ok {
			value := after
			values = append(values, value)
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

	// Walk the AST looking for pipe nodes
	hasPipe := false

	syntax.Walk(file, func(node syntax.Node) bool {
		// Check for BinaryCmd with pipe operator
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

// isPipName reports whether a name is one pip goes by, as a command or as a
// python module. Matched exactly, so that pipenv — a different tool — is not.
func isPipName(name string) bool {
	switch name {
	case "pip", "pip2", "pip3":
		return true
	default:
		return false
	}
}

// IsPipInstall checks if a command is a pip install command.
// Ported from Hadolint.Shell.isPipInstall.
func IsPipInstall(cmd Command) bool {
	// pip, pip3, pipx and whatever else pip goes by; pipenv is a different
	// tool. install may follow flags.
	if strings.HasPrefix(cmd.Name, "pip") && !strings.HasPrefix(cmd.Name, "pipenv") {
		return slices.Contains(GetArgs(cmd), "install")
	}

	// Check for: python -m pip install
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
		if args[idx] != "-m" || !isPipName(args[idx+1]) {
			continue
		}

		if slices.Contains(args[idx+2:], "install") {
			return true
		}
	}

	return false
}
