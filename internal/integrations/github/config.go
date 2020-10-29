package github

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// Configuration encapsulates information needed for communicating with a
// GitHub API instance (Enterprise / Non-enterprise)
type Configuration struct {
	BaseURL             string
	UploadURL           string
	UserName            string
	PersonalAccessToken string
	Organizations       *cli.StringSlice
}

// Validate ensures the configuration provided contains the required information.
func (c *Configuration) Validate() error {
	if c.UserName == "" {
		return fmt.Errorf("a username must be provided")
	} else if c.PersonalAccessToken == "" {
		return fmt.Errorf("a personal access token must be provided")
	}
	return nil
}

// DefaultConfigWithFlags returns configuration and flags specific to GitHub
func DefaultConfigWithFlags() (*Configuration, []cli.Flag) {
	cfg := &Configuration{
		Organizations: cli.NewStringSlice(),
	}

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "github-base-url",
			Usage:       "url to the GitHub Enterprise instance",
			Destination: &(cfg.BaseURL),
			Value:       cfg.BaseURL,
			EnvVars:     []string{"GITHUB_BASE_URL"},
		},
		&cli.StringFlag{
			Name:        "github-upload-url",
			Usage:       "url to the upload endpoint of the GitHub Enterprise instance",
			Destination: &(cfg.UploadURL),
			Value:       cfg.UploadURL,
			EnvVars:     []string{"GITHUB_UPLOAD_URL"},
		},
		&cli.StringFlag{
			Name:        "github-username",
			Usage:       "the user associated with the personal access token",
			Destination: &(cfg.UserName),
			Value:       cfg.UserName,
			EnvVars:     []string{"GITHUB_USERNAME"},
		},
		&cli.StringFlag{
			Name:        "github-access-token",
			Usage:       "used to read data from the GitHub API and clone repositories",
			Destination: &(cfg.PersonalAccessToken),
			Value:       cfg.PersonalAccessToken,
			EnvVars:     []string{"GITHUB_ACCESS_TOKEN"},
		},
		&cli.StringSliceFlag{
			Name:        "github-organizations",
			Usage:       "restricts operations to listed GitHub organizations",
			Destination: cfg.Organizations,
			Value:       cfg.Organizations,
			EnvVars:     []string{"GITHUB_ORGANIZATIONS"},
		},
	}

	return cfg, flags
}
