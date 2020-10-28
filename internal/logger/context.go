package logger

import (
	"context"
	"log"

	"go.uber.org/zap"
)

// StringKey is used to represent string keys in Context
type StringKey string

const loggerKey = StringKey("EFFX_LOGGER")

// MustGetFromContext returns the logger attached to the context or fails if it's absent.
func MustGetFromContext(ctx context.Context) *zap.Logger {
	v := ctx.Value(loggerKey)
	if v == nil {
		log.Fatal("logger not set on context!")
	}
	return v.(*zap.Logger)
}

// AttachToContext injects the logger into the context for later access.
func AttachToContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
