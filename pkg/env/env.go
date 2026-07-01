package env

import (
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// Production is the production environment identifier.
	Production = "production"
	// Testing is the test environment identifier.
	Testing = "testing"
	// Development is the development environment identifier.
	Development = "development"

	prodShort = "prod"
	testShort = "test"
	devShort  = "dev"
)

var (
	envMultiSpace  = regexp.MustCompile(`\s+`)
	envNonAlphaNum = regexp.MustCompile(`[^a-z0-9\s]`)
)

// sanitize lowercases and strips non-alphanumeric characters from
// env-var values. Used internally for APP_ENV / APP_DEBUG / similar
// "tag-shaped" variables where we want loose matching. Inlined
// here rather than imported from pkg/utils.SanitizeString so this
// package has no upward dependency on the legacy utils grab-bag.
func sanitize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = envNonAlphaNum.ReplaceAllString(s, "")
	s = envMultiSpace.ReplaceAllString(s, " ")
	return s
}

// GetOrDefault returns the value of the environment variable key,
// or defaultVal when the variable is unset or empty. The result is
// sanitized (lowercased + non-alphanumerics stripped), which is
// the historical contract — appropriate for the "tag-shaped"
// variables this helper exists to read (APP_ENV, APP_DEBUG, …).
// Do NOT use it for secrets or URLs whose value must survive
// punctuation untouched.
func GetOrDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		val = defaultVal
	}
	return sanitize(val)
}

// InitApp reads APP_ENV / GO_ENV, sets the matching Gin mode, and
// returns the canonical environment name (one of Production /
// Testing / Development). Called from every binary's main during
// the first lines of boot.
func InitApp() string {
	appEnv := GetOrDefault("APP_ENV", os.Getenv("GO_ENV"))
	appDebug := GetOrDefault("APP_DEBUG", "")

	var env string
	switch appEnv {
	case Production, prodShort:
		env = Production
		gin.SetMode(gin.ReleaseMode)
	case Testing, testShort:
		env = Testing
		gin.SetMode(gin.TestMode)
	default:
		env = Development
		gin.SetMode(gin.DebugMode)
	}

	if appDebug != "true" {
		gin.DisableConsoleColor()
		gin.SetMode(gin.ReleaseMode)
		log.SetOutput(os.Stdout)
	}

	return env
}

// IsDev reports whether the current environment is development or
// testing. Used by callers that want to relax certain checks in
// dev (e.g. allow weak default passwords).
func IsDev() bool {
	appEnv := strings.ToLower(
		GetOrDefault("APP_ENV", os.Getenv("GO_ENV")),
	)
	return appEnv == Development || appEnv == devShort ||
		appEnv == Testing ||
		appEnv == testShort
}
