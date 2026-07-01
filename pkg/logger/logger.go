// Package logger wraps zap with the conventions the App backend
// uses: development mode produces colored console output, production
// mode produces JSON with a buffered write syncer for throughput.
// The Logger type is the single accepted logging primitive across
// every service, repository, middleware, and worker.
package logger

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// requestIDGinKey mirrors httpx.KeyRequestID. Duplicated as a
// string constant so logger can run without importing httpx (httpx
// imports logger, so the reverse would form an import cycle).
const requestIDGinKey = "request_id"

// Logger is the project-wide structured logger. Methods come in two
// flavors: plain (Info/Error/Debug/Warn/Fatal/Printf) and Gin-aware
// (InfoCtx/ErrorCtx/DebugCtx/WarnCtx) which automatically attach the
// current request_id field. Construct with New; call Sync once at
// shutdown to flush the buffered production sink.
type Logger struct {
	appLogger *zap.SugaredLogger
	// buf is non-nil only in production mode. The buffer is flushed
	// on Sync and on a 1-second ticker; in dev mode logs go to
	// stdout unbuffered.
	buf *zapcore.BufferedWriteSyncer
}

// New constructs a Logger configured for the given environment.
// "production" enables JSON encoding plus a 256 KiB buffered write
// syncer (flushed every second); anything else uses a colored
// console encoder at debug level. AddCallerSkip(1) is applied so
// the call-site file:line refers to the caller of Info/Error/etc.
// rather than this wrapper.
func New(env string) *Logger {
	var core zapcore.Core
	var buf *zapcore.BufferedWriteSyncer

	if env == "production" {
		encoderCfg := zap.NewProductionEncoderConfig()
		encoder := zapcore.NewJSONEncoder(encoderCfg)

		buf = &zapcore.BufferedWriteSyncer{
			WS:            zapcore.AddSync(os.Stdout),
			Size:          256 * 1024,
			FlushInterval: time.Second,
		}

		core = zapcore.NewCore(encoder, buf, zapcore.InfoLevel)
	} else {
		encoderCfg := zap.NewDevelopmentEncoderConfig()
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout(
			"2006-01-02 15:04:05")
		encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder
		encoderCfg.ConsoleSeparator = " "
		encoder := zapcore.NewConsoleEncoder(encoderCfg)

		core = zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			zapcore.DebugLevel,
		)
	}

	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)

	return &Logger{appLogger: logger.Sugar(), buf: buf}
}

// Info logs at the info level with structured key/value fields.
// Pass alternating key, value pairs as keysAndValues.
func (l *Logger) Info(msg string, keysAndValues ...any) {
	l.appLogger.Infow(msg, keysAndValues...)
}

// Error logs at the error level with structured key/value fields.
func (l *Logger) Error(msg string, keysAndValues ...any) {
	l.appLogger.Errorw(msg, keysAndValues...)
}

// Debug logs at the debug level. Debug entries are dropped in
// production mode (info threshold).
func (l *Logger) Debug(msg string, keysAndValues ...any) {
	l.appLogger.Debugw(msg, keysAndValues...)
}

// Warn logs at the warn level with structured key/value fields.
func (l *Logger) Warn(msg string, keysAndValues ...any) {
	l.appLogger.Warnw(msg, keysAndValues...)
}

// Fatal logs at the fatal level and then calls os.Exit(1). Use only
// from cmd/* main functions when startup cannot proceed.
func (l *Logger) Fatal(msg string, keysAndValues ...any) {
	l.appLogger.Fatalw(msg, keysAndValues...)
}

// Printf logs an info-level message using fmt-style formatting.
// Provided so third-party libraries that expect a Printf interface
// (Asynq, GORM logger adapters) can use Logger directly.
func (l *Logger) Printf(format string, args ...any) {
	l.appLogger.Infof(format, args...)
}

// Sync flushes the buffered production write syncer and the
// underlying zap logger. Call once during graceful shutdown.
func (l *Logger) Sync() {
	if l.buf != nil {
		_ = l.buf.Sync()
	}
	if l.appLogger != nil {
		_ = l.appLogger.Sync()
	}
}

// HTTP logs a single access-log line with the canonical App
// fields (method, path, status, duration) plus any extra structured
// fields supplied by the access-log middleware.
func (l *Logger) HTTP(
	method, path, status string,
	duration any,
	fields ...any,
) {
	allFields := []any{
		"method", method,
		"path", path,
		"status", status,
		"duration", duration,
	}
	allFields = append(allFields, fields...)
	l.appLogger.Infow("HTTP Request", allFields...)
}

// InfoCtx logs at info level with the current request_id from the
// Gin context automatically prepended to keysAndValues.
func (l *Logger) InfoCtx(
	c *gin.Context,
	msg string,
	keysAndValues ...any,
) {
	args := append(
		[]any{requestIDGinKey, c.GetString(requestIDGinKey)},
		keysAndValues...)
	l.appLogger.Infow(msg, args...)
}

// ErrorCtx logs at error level with the current request_id from the
// Gin context automatically prepended to keysAndValues.
func (l *Logger) ErrorCtx(
	c *gin.Context,
	msg string,
	keysAndValues ...any,
) {
	args := append(
		[]any{requestIDGinKey, c.GetString(requestIDGinKey)},
		keysAndValues...)
	l.appLogger.Errorw(msg, args...)
}

// DebugCtx logs at debug level with the current request_id from the
// Gin context automatically prepended to keysAndValues.
func (l *Logger) DebugCtx(
	c *gin.Context,
	msg string,
	keysAndValues ...any,
) {
	args := append(
		[]any{requestIDGinKey, c.GetString(requestIDGinKey)},
		keysAndValues...)
	l.appLogger.Debugw(msg, args...)
}

// WarnCtx logs at warn level with the current request_id from the
// Gin context automatically prepended to keysAndValues.
func (l *Logger) WarnCtx(
	c *gin.Context,
	msg string,
	keysAndValues ...any,
) {
	args := append(
		[]any{requestIDGinKey, c.GetString(requestIDGinKey)},
		keysAndValues...)
	l.appLogger.Warnw(msg, args...)
}
