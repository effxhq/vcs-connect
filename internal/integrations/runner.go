package integrations

import (
	"context"

	"github.com/effxhq/vcs-connect/internal/model"
)

type Runner interface {
	Run(ctx context.Context, data chan *model.Repository) error
}
