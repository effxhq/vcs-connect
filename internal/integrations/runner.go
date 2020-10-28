package integrations

import (
	"context"

	"github.com/effxhq/vcs-connect/internal/model"
)

// Runner provides a generic interface that supports shutdown by cancelling
// a context and exchanging data through a chan.
type Runner interface {
	Run(ctx context.Context, data chan *model.Repository) error
}
