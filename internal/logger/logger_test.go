package logger_test

import (
	"context"
	"testing"

	"github.com/effxhq/vcs-connect/internal/logger"

	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	log, err := logger.Setup()
	require.NoError(t, err)

	ctx := context.Background()
	ctx = logger.AttachToContext(ctx, log)

	log2 := logger.MustGetFromContext(ctx)
	require.Equal(t, log, log2)
}
