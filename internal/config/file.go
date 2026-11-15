package config

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"go.yaml.in/yaml/v4"

	"github.com/forkcloser/godolint/internal/rule"
)

// The errors Read wraps, each with the file and what it found there.
var (
	// ErrUnknownKey is a key hadolint's configuration does not have.
	ErrUnknownKey = errors.New("unknown key")
	// ErrNotImplemented is a setting of hadolint's that godolint does not
	// carry out, refused so that a file is never half obeyed.
	ErrNotImplemented = errors.New("not implemented")
	// ErrInvalidValue is a value its key does not take.
	ErrInvalidValue = errors.New("invalid value")
)

// formatJSON is the one output format there is.
const formatJSON = "json"

// File is what a hadolint configuration file says, key by key, under
// hadolint's names. A nil pointer is a key the file does not have.
//
//nolint:tagliatelle // the keys are hadolint's, in the two cases it writes them in
type File struct {
	NoFail              *bool              `yaml:"no-fail"`
	NoColor             *bool              `yaml:"no-color"`
	Verbose             *bool              `yaml:"verbose"`
	Format              *string            `yaml:"format"`
	Formats             []string           `yaml:"formats"`
	Override            Override           `yaml:"override"`
	Ignored             []string           `yaml:"ignored"`
	TrustedRegistries   []string           `yaml:"trustedRegistries"`
	LabelSchema         map[string]*string `yaml:"label-schema"`
	StrictLabels        *bool              `yaml:"strict-labels"`
	DisableIgnorePragma *bool              `yaml:"disable-ignore-pragma"`
	FailureThreshold    *string            `yaml:"failure-threshold"`
}

// Override is the rule codes a file gives another severity, by severity.
type Override struct {
	Error   []string `yaml:"error"`
	Warning []string `yaml:"warning"`
	Info    []string `yaml:"info"`
	Style   []string `yaml:"style"`
}

// Read reads the configuration file at path. A file with nothing in it but
// comments says nothing. A key hadolint does not have, a format other than
// json, a label type or a failure threshold that is not one, and a file that
// is not YAML are errors naming the file and what is wrong in it.
func Read(path string) (*File, error) {
	content, err := os.ReadFile(path) // #nosec G304 -- the configuration file is the caller's to name
	if err != nil {
		return nil, fmt.Errorf("reading the configuration file: %w", err)
	}

	file := &File{}
	if saysNothing(content) {
		return file, nil
	}

	var keys map[string]any
	if err := yaml.Load(content, &keys); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	if err := checkKeys(keys); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	if err := yaml.Load(content, file); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	if err := file.check(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	return file, nil
}

// saysNothing reports whether content holds only blank lines and comments,
// which YAML calls no document at all.
func saysNothing(content []byte) bool {
	for line := range strings.SplitSeq(string(content), "\n") {
		if text := strings.TrimSpace(line); text != "" && !strings.HasPrefix(text, "#") {
			return false
		}
	}

	return true
}

// checkKeys refuses a key that is not one of hadolint's, at the top or under
// override, by its name.
func checkKeys(keys map[string]any) error {
	known := []string{
		"no-fail", "no-color", "verbose", "format", "formats", "override", "ignored",
		"trustedRegistries", "label-schema", "strict-labels", "disable-ignore-pragma", "failure-threshold",
	}

	for _, key := range slices.Sorted(maps.Keys(keys)) {
		if !slices.Contains(known, key) {
			return fmt.Errorf("%w `%s`", ErrUnknownKey, key)
		}
	}

	// An override that is no mapping is refused when the file is read into
	// its shape.
	override, isMapping := keys["override"].(map[string]any)
	if !isMapping {
		return nil
	}

	for _, key := range slices.Sorted(maps.Keys(override)) {
		if _, ok := ParseSeverity(key); !ok || key == severityIgnore || key == severityNone {
			return fmt.Errorf("%w `override.%s`", ErrUnknownKey, key)
		}
	}

	return nil
}

// check refuses the values the keys do not take, and the formats godolint
// does not write.
func (f *File) check() error {
	formats := slices.Clone(f.Formats)
	if f.Format != nil {
		formats = append(formats, *f.Format)
	}

	for _, format := range formats {
		if format != formatJSON {
			return fmt.Errorf("%w: format `%s` (godolint writes json)", ErrNotImplemented, format)
		}
	}

	for _, label := range slices.Sorted(maps.Keys(f.LabelSchema)) {
		// hadolint takes a type written as a string, the empty one for text,
		// and not a label left with none.
		name := f.LabelSchema[label]
		if name == nil {
			return fmt.Errorf("%w: label-schema gives `%s` no type", ErrInvalidValue, label)
		}

		if _, ok := ParseLabelType(*name); !ok {
			return fmt.Errorf("%w: label-schema type `%s` of `%s`", ErrInvalidValue, *name, label)
		}
	}

	if f.FailureThreshold != nil {
		if _, ok := ParseSeverity(*f.FailureThreshold); !ok {
			return fmt.Errorf("%w: failure-threshold `%s`", ErrInvalidValue, *f.FailureThreshold)
		}
	}

	return nil
}

// The two names hadolint has for the severity below every other.
const (
	severityIgnore = "ignore"
	severityNone   = "none"
)

// ParseSeverity reads a severity under hadolint's names: error, warning,
// info, style, and ignore or none.
func ParseSeverity(name string) (rule.Severity, bool) {
	switch name {
	case "error":
		return rule.Error, true
	case "warning":
		return rule.Warning, true
	case "info":
		return rule.Info, true
	case "style":
		return rule.Style, true
	case severityIgnore, severityNone:
		return rule.Ignore, true
	default:
		return rule.Ignore, false
	}
}

// ParseLabelType reads a label type under hadolint's names: text, url,
// spdx, hash, rfc3339, semver and email. No name is text.
func ParseLabelType(name string) (LabelType, bool) {
	switch name {
	case "", "text":
		return LabelTypeRawText, true
	case "url":
		return LabelTypeURL, true
	case "spdx":
		return LabelTypeSPDX, true
	case "hash":
		return LabelTypeGitHash, true
	case "rfc3339":
		return LabelTypeRFC3339, true
	case "semver":
		return LabelTypeSemVer, true
	case "email":
		return LabelTypeEmail, true
	default:
		return "", false
	}
}

// Find is the configuration file hadolint would use when none is named, and
// whether there is one: the first that exists of .hadolint.yaml and
// .hadolint.yml in the working directory; hadolint.yaml and hadolint.yml in
// the user's configuration directory; those two, config.yaml and config.yml
// in hadolint's own directory; .hadolint.yaml and .hadolint.yml in the
// user's home. A directory that cannot be told (no home) holds none.
func Find() (string, bool) {
	hidden := []string{".hadolint.yaml", ".hadolint.yml"}
	visible := []string{"hadolint.yaml", "hadolint.yml"}
	own := []string{"hadolint.yaml", "hadolint.yml", "config.yaml", "config.yml"}

	var candidates []string

	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, under(cwd, hidden)...)
	}

	configDir, ownDir, home := userDirectories()
	candidates = append(candidates, under(configDir, visible)...)
	candidates = append(candidates, under(ownDir, own)...)
	candidates = append(candidates, under(home, hidden)...)

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
			return candidate, true
		}
	}

	return "", false
}

// under is names in dir, none when there is no dir.
func under(dir string, names []string) []string {
	if dir == "" {
		return nil
	}

	paths := make([]string, 0, len(names))
	for _, name := range names {
		paths = append(paths, filepath.Join(dir, name))
	}

	return paths
}

// userDirectories is where hadolint looks beyond the working directory: the
// user's configuration directory, hadolint's own, and the home it looks for
// a hidden file in. On Windows these are %APPDATA%, %APPDATA%\hadolint and
// the Documents folder under the profile; elsewhere $XDG_CONFIG_HOME (or
// ~/.config), ~/.hadolint and the home. An empty one is one there is not.
func userDirectories() (configDir, ownDir, home string) {
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			configDir, ownDir = appData, filepath.Join(appData, "hadolint")
		}

		if profile := os.Getenv("USERPROFILE"); profile != "" {
			home = filepath.Join(profile, "Documents")
		}

		return configDir, ownDir, home
	}

	home = os.Getenv("HOME")

	// An XDG directory that is not absolute is not one (the XDG specification).
	configDir = os.Getenv("XDG_CONFIG_HOME")
	if !filepath.IsAbs(configDir) {
		configDir = ""

		if home != "" {
			configDir = filepath.Join(home, ".config")
		}
	}

	if home != "" {
		ownDir = filepath.Join(home, ".hadolint")
	}

	return configDir, ownDir, home
}
