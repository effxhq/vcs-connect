package logger

import (
	"context"
	"log"

	"go.uber.org/zap"
)

const loggerKey = "EFFX_LOGGER"

func MustGetFromContext(ctx context.Context) *zap.Logger {
	v := ctx.Value(loggerKey)
	if v == nil {
		log.Fatal("logger not set on context!")
	}
	return v.(*zap.Logger)
}

func AttachContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
