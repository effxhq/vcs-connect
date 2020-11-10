package effx

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// Configuration encapsulates information needed for communicating with effx
type Configuration struct {
	APIHost string
	APIKey  string
}

// Validate ensures the configuration provided contains the required information.
func (c *Configuration) Validate() error {
	if c.APIHost == "" {
		return fmt.Errorf("an api host must be provided")
	} else if c.APIKey == "" {
		return fmt.Errorf("an api key must be provided")
	}
	return nil
}

// DefaultConfigWithFlags returns configuration and flags specific to effx
func DefaultConfigWithFlags() (*Configuration, []cli.Flag) {
	cfg := &Configuration{
		APIHost: "api.effx.io",
	}

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "effx-api-host",
			Usage:       "where the effx api is located",
			Destination: &(cfg.APIHost),
			Value:       cfg.APIHost,
			EnvVars:     []string{"EFFX_API_HOST"},
		},
		&cli.StringFlag{
			Name:        "effx-api-key",
			Usage:       "the key associated with your effx acount",
			Destination: &(cfg.APIKey),
			Value:       cfg.APIKey,
			EnvVars:     []string{"EFFX_API_KEY"},
		},
	}

	return cfg, flags
}
