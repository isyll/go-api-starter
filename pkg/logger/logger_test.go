package logger_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/isyll/go-api-starter/pkg/logger"
)

func TestLogger_New(t *testing.T) {
	// Dev
	logDev := logger.New("development")
	assert.NotNil(t, logDev)
	logDev.Sync()

	// Prod
	logProd := logger.New("production")
	assert.NotNil(t, logProd)
	logProd.Sync()
}

// Call all simple methods just to ensure no panics

func TestLogger_ContextMethods(t *testing.T) {
	l := logger.New("development")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequestWithContext(
		context.Background(),
		"GET",
		"/",
		http.NoBody,
	)
	c.Set("request_id", "req-1234")

	l.InfoCtx(c, "info ctx", "k", "v")
	l.ErrorCtx(c, "error ctx", "k", "v")
	l.WarnCtx(c, "warn ctx", "k", "v")
	l.DebugCtx(c, "debug ctx", "k", "v")

	// Also should not panic or fail with an empty context
	cEmpty, _ := gin.CreateTestContext(w)
	cEmpty.Request, _ = http.NewRequestWithContext(
		context.Background(),
		"GET",
		"/",
		http.NoBody,
	)
	l.InfoCtx(cEmpty, "info empty ctx", "another", "val")
}
