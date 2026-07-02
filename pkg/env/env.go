package env

import (
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	Production  = "production"
	Testing     = "testing"
	Development = "development"

	prodShort = "prod"
	testShort = "test"
	devShort  = "dev"
)

var (
	envMultiSpace  = regexp.MustCompile(`\s+`)
	envNonAlphaNum = regexp.MustCompile(`[^a-z0-9\s]`)
)

func sanitize(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = envNonAlphaNum.ReplaceAllString(s, "")
	s = envMultiSpace.ReplaceAllString(s, " ")
	return s
}

func GetOrDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		val = defaultVal
	}
	return sanitize(val)
}

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

func IsDev() bool {
	appEnv := strings.ToLower(
		GetOrDefault("APP_ENV", os.Getenv("GO_ENV")),
	)
	return appEnv == Development || appEnv == devShort ||
		appEnv == Testing ||
		appEnv == testShort
}
