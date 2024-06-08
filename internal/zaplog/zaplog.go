package zaplog

import (
	"context"
	"net/http"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ConfigureZapLogger() *zap.Logger {
	atom := zap.NewAtomicLevelAt(zap.DebugLevel)
	consoleErrors := zapcore.Lock(os.Stderr)
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapcore.NewCore(consoleEncoder, consoleErrors, atom)
	zl := zap.New(core)
	zap.ReplaceGlobals(zl)
	http.Handle("/loglevel", atom) // on default handler at 4200

	return zl
}

type contextKey struct{}

func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

func FromContext(ctx context.Context) *zap.Logger {
	logger := zap.L() // default to global logger
	if l, ok := ctx.Value(contextKey{}).(*zap.Logger); ok {
		logger = l
	}
	return logger
}
