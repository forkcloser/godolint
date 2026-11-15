package rules

import (
	"strings"

	"github.com/forkcloser/godolint/internal/rule"
	"github.com/forkcloser/godolint/internal/syntax"
)

// DL3064 creates the rule that ARG and ENV do not carry sensitive data.
// Ported from Hadolint.Rule.DL3064; the engine applies it inside ONBUILD too,
// as hadolint's `onbuild dl3064` does.
func DL3064() rule.Rule {
	return rule.NewSimpleRule(
		DL3064Meta.Code,
		DL3064Meta.Severity,
		DL3064Meta.Message,
		checkDL3064,
	)
}

// checkDL3064 fails an ARG whose name, or an ENV one of whose keys, names a
// known secret or contains a suspicious word.
func checkDL3064(instruction syntax.Instruction) bool {
	switch instr := instruction.(type) {
	case *syntax.Arg:
		return !sensitiveName(instr.ArgName)
	case *syntax.Env:
		for _, pair := range instr.Pairs {
			if sensitiveName(pair.Key) {
				return false
			}
		}

		return true
	default:
		return true
	}
}

// sensitiveName is hadolint's test: the name upper-cased is a known sensitive
// variable, or the name lower-cased contains a suspicious substring. A name
// ending in _FILE is neither: it holds the path to a secret
// (POSTGRES_PASSWORD_FILE), not the secret.
func sensitiveName(name string) bool {
	upper := strings.ToUpper(name)
	if strings.HasSuffix(upper, "_FILE") {
		return false
	}

	if knownSensitiveName(upper) {
		return true
	}

	lower := strings.ToLower(name)
	for _, word := range []string{
		"api_key", "client_key", "password", "private_key", "privatekey", "secret", "token", "username",
	} {
		if strings.Contains(lower, word) {
			return true
		}
	}

	return false
}

// knownSensitiveName is hadolint's knownSensitiveNames list, matched on the
// upper-cased name.
func knownSensitiveName(upper string) bool {
	switch upper {
	case "ACCESS_TOKEN", "APPLICATION_KEY", "APP_SECRET", "AUTH_TOKEN",
		"AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY",
		"BITTREX_API_KEY", "BITTREX_API_SECRET",
		"CF_PASSWORD", "CF_USERNAME", "CIRCLE_TOKEN", "CI_DEPLOY_PASSWORD", "CI_DEPLOY_USER",
		"DOCKERHUB_PASSWORD", "DOCKER_EMAIL", "DOCKER_PASSWORD", "DOCKER_USERNAME",
		"FACEBOOK_ACCESS_TOKEN", "FACEBOOK_APP_SECRET",
		"FIREBASE_API_TOKEN", "FIREBASE_TOKEN", "FOSSA_API_KEY",
		"GH_ENTERPRISE_TOKEN", "GH_TOKEN", "GITHUB_ENTERPRISE_TOKEN", "GITHUB_TOKEN",
		"HEROKU_API_KEY", "HEROKU_API_USER",
		"NPM_AUTH_TOKEN", "NPM_TOKEN",
		"OKTA_CLIENT_TOKEN", "OKTA_OAUTH2_CLIENTSECRET",
		"OPENAI_API_KEY", "OS_PASSWORD", "OS_USERNAME", "POSTGRES_PASSWORD",
		"SLACK_TOKEN", "STRIPE_API_KEY",
		"VAULT_CLIENT_KEY", "VAULT_TOKEN":
		return true
	default:
		return false
	}
}
