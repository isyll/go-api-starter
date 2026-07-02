// Package logger wraps zap with the backend logging conventions.
package logger

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const requestIDGinKey = "request_id"

type Logger struct {
	appLogger *zap.SugaredLogger
	buf       *zapcore.BufferedWriteSyncer
}

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

func (l *Logger) Info(msg string, keysAndValues ...any) {
	l.appLogger.Infow(msg, keysAndValues...)
}

func (l *Logger) Error(msg string, keysAndValues ...any) {
	l.appLogger.Errorw(msg, keysAndValues...)
}

func (l *Logger) Debug(msg string, keysAndValues ...any) {
	l.appLogger.Debugw(msg, keysAndValues...)
}

func (l *Logger) Warn(msg string, keysAndValues ...any) {
	l.appLogger.Warnw(msg, keysAndValues...)
}

func (l *Logger) Fatal(msg string, keysAndValues ...any) {
	l.appLogger.Fatalw(msg, keysAndValues...)
}

func (l *Logger) Printf(format string, args ...any) {
	l.appLogger.Infof(format, args...)
}

func (l *Logger) Sync() {
	if l.buf != nil {
		_ = l.buf.Sync()
	}
	if l.appLogger != nil {
		_ = l.appLogger.Sync()
	}
}

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
