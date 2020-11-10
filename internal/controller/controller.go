package controller

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/effxhq/vcs-connect/internal/integrations"
	"github.com/effxhq/vcs-connect/internal/logger"
	"github.com/effxhq/vcs-connect/internal/model"
	"github.com/effxhq/vcs-connect/internal/run"
)

// New returns a new controller that manages the pipeline between the integration and the consumers
func New(
	cfg *Configuration,
	integration integrations.Runner,
	consumer *run.Consumer,
) (*Controller, error) {

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Controller{
		integration: integration,
		consumer:    consumer,
		workers:     cfg.Workers,
	}, nil
}

// Controller encapsulates the logic of spinning up multiple workers to feed from
// a common integration
type Controller struct {
	integration integrations.Runner
	consumer    *run.Consumer
	workers     int
}

// Run performs a single pass over the data.
func (c *Controller) Run(parent context.Context) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	ctx = logger.AttachToContext(ctx, logger.MustSetup())

	data := make(chan *model.Repository)

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		defer signal.Stop(signals)
		<-signals
		cancel()
	}()

	for i := 0; i < c.workers; i++ {
		go c.consumer.Run(ctx, data)
	}

	// Run the integration until completion
	return c.integration.Run(ctx, data)
}
