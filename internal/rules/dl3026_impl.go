package rules

import (
	"maps"
	"regexp"
	"strings"

	"github.com/forkcloser/godolint/internal/config"
	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

const (
	// An image with no registry in its name comes from Docker Hub, which is
	// allowed under either of these names.
	dockerHub     = "docker.io"
	dockerHubSite = "hub.docker.com"

	// The one host with neither a dot nor a port that is a registry.
	localhostName = "localhost"

	mountFlag      = "--mount="
	mountTypeBind  = "bind"
	mountTypeCache = "cache"
)

// dl3026State is the aliases of the stages so far: an image named as one of
// them is that stage, whatever the registries allowed.
type dl3026State struct {
	aliases map[string]bool
}

// DL3026Rule checks for allowed registries.
type DL3026Rule struct {
	rule.StatefulRuleBase

	allowedRegistries []string // Empty means all registries allowed
}

// DL3026 creates a rule that allows every registry.
func DL3026() rule.Rule {
	return DL3026WithConfig(config.Default())
}

// DL3026WithConfig creates the rule with the registries cfg allows.
func DL3026WithConfig(cfg *config.Config) rule.Rule {
	return &DL3026Rule{
		StatefulRuleBase:  rule.NewStatefulRuleBase(DL3026Meta),
		allowedRegistries: cfg.AllowedRegistries,
	}
}

// InitialState returns the initial state for this rule.
func (*DL3026Rule) InitialState() rule.State {
	return rule.EmptyState(dl3026State{
		aliases: make(map[string]bool),
	})
}

// Check reports an image from a registry that is not allowed: the image of a
// FROM, of a COPY --from, and of a bind or cache mount of a RUN.
// Ported from DL3026.hs.
func (r *DL3026Rule) Check(line int, state rule.State, instruction syntax.Instruction) rule.State {
	current := rule.Data[dl3026State](state)

	switch inst := instruction.(type) {
	case *syntax.From:
		// The stage's own alias joins after its image is judged: an image
		// named as its own alias is not an earlier stage.
		allowed := r.allows(inst.Image.Image, current.aliases)

		if inst.Image.Alias != nil {
			aliases := maps.Clone(current.aliases)
			aliases[*inst.Image.Alias] = true
			state = state.ReplaceData(dl3026State{aliases: aliases})
		}

		if allowed {
			return state
		}

	case *syntax.Copy:
		if inst.From == nil || r.allowsSource(*inst.From, current.aliases) {
			return state
		}

	case *syntax.Run:
		if r.allowsMounts(inst.Flags, current.aliases) {
			return state
		}

	default:
		return state
	}

	return state.AddFailure(rule.Fail(r, line))
}

// allowsMounts reports whether every bind and cache mount among flags takes
// its files from an allowed image. A mount with no from= takes them from
// the build, and the other types take none from an image.
func (r *DL3026Rule) allowsMounts(flags []string, aliases map[string]bool) bool {
	for _, flag := range flags {
		options, isMount := strings.CutPrefix(flag, mountFlag)
		if !isMount {
			continue
		}

		mountType, from := mountTypeBind, ""

		for option := range strings.SplitSeq(options, ",") {
			switch key, value, _ := strings.Cut(option, "="); key {
			case "type":
				mountType = value
			case "from":
				from = value
			}
		}

		if from == "" || (mountType != mountTypeBind && mountType != mountTypeCache) {
			continue
		}

		if !r.allowsSource(from, aliases) {
			return false
		}
	}

	return true
}

// allowsSource reports whether the source of a COPY --from or of a mount may
// be used: a stage by its number or by its name, or an image allows passes.
func (r *DL3026Rule) allowsSource(from string, aliases map[string]bool) bool {
	return isStageIndex(from) || aliases[stageNamed(from)] || r.allows(from, aliases)
}

// allows reports whether image may be used: an earlier stage, any image when
// no registry is named as allowed, scratch, or an image of an allowed
// registry.
func (r *DL3026Rule) allows(image string, aliases map[string]bool) bool {
	if aliases[image] || len(r.allowedRegistries) == 0 || image == scratchImage {
		return true
	}

	registry, named := imageRegistry(image)
	if named {
		return r.isRegistryAllowed(registry)
	}

	return r.isRegistryAllowed(dockerHub) || r.isRegistryAllowed(dockerHubSite)
}

// variableReference is a ${…} in an image's name.
var variableReference = regexp.MustCompile(`\$\{[^}]*\}`)

// imageRegistry is the registry in an image's name, and whether it has one:
// the part before the first slash when it is a host, which a dot, the colon
// of a port or the name localhost makes it. Without one the image comes from
// Docker Hub. The colon inside a ${NS:-default} is no port.
func imageRegistry(image string) (string, bool) {
	host, _, found := strings.Cut(image, "/")
	if !found {
		return "", false
	}

	port := strings.Contains(variableReference.ReplaceAllString(host, ""), ":")
	if port || strings.Contains(host, ".") || host == localhostName {
		return host, true
	}

	return "", false
}

// isStageIndex reports whether from names a stage by its number, as
// COPY --from=0 does: a stage of this build, of no registry.
func isStageIndex(from string) bool {
	if from == "" {
		return false
	}

	for _, char := range from {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

func (r *DL3026Rule) isRegistryAllowed(registry string) bool {
	for _, allowed := range r.allowedRegistries {
		if matchRegistry(allowed, registry) {
			return true
		}
	}

	return false
}

// matchRegistry reports whether registry matches an allowed pattern: a
// leading `*` is any registry that ends as the rest does (so `*` alone is
// any), a trailing `*` any that starts as the rest does, anything else the
// registry itself.
func matchRegistry(allowed, registry string) bool {
	if suffix, ok := strings.CutPrefix(allowed, "*"); ok {
		return strings.HasSuffix(registry, suffix)
	}

	if prefix, ok := strings.CutSuffix(allowed, "*"); ok {
		return strings.HasPrefix(registry, prefix)
	}

	return registry == allowed
}
