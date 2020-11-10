package controller

import (
	"fmt"
	"os"
	"path"

	"github.com/urfave/cli/v2"
)

// Configuration encapsulates information used by the control loop.
type Configuration struct {
	ScratchDir string
	Workers    int
}

// Validate ensures the configuration provided contains the required information.
func (c *Configuration) Validate() error {
	if c.ScratchDir != "" {
		return fmt.Errorf("a scratch dir must be provided")
	} else if c.Workers <= 0 {
		return fmt.Errorf("at least one worker must be configured")
	}
	return nil
}

// DefaultConfigWithFlags returns configuration and flags specific to the control loop.
func DefaultConfigWithFlags() (*Configuration, []cli.Flag) {
	cfg := &Configuration{
		ScratchDir: path.Join(os.TempDir(), "effx-vcs-connect"),
		Workers:    1,
	}

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "scratch-dir",
			Usage:       "scratch space used to clone repositories for indexing",
			Destination: &(cfg.ScratchDir),
			Value:       cfg.ScratchDir,
			EnvVars:     []string{"SCRATCH_DIR"},
		},
		&cli.IntFlag{
			Name:        "workers",
			Usage:       "the number of concurrent repositories cloned at once",
			Destination: &(cfg.Workers),
			Value:       cfg.Workers,
			EnvVars:     []string{"WORKERS"},
		},
	}

	return cfg, flags
}
