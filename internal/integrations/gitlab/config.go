package gitlab

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

type Configuration struct {
	BaseURL             string
	UserName            string
	PersonalAccessToken string
	Groups              *cli.StringSlice
}

func (c *Configuration) Validate() error {
	if c.UserName == "" {
		return fmt.Errorf("a username must be provided")
	} else if c.PersonalAccessToken == "" {
		return fmt.Errorf("a personal access token must be provided")
	}
	return nil
}

func DefaultConfigWithFlags() (*Configuration, []cli.Flag) {
	cfg := &Configuration{
		Groups: cli.NewStringSlice(),
	}

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "gitlab-base-url",
			Usage:       "url to the GitHub instance",
			Destination: &(cfg.BaseURL),
			Value:       cfg.BaseURL,
			EnvVars:     []string{"GITLAB_BASE_URL"},
		},
		&cli.StringFlag{
			Name:        "gitlab-username",
			Usage:       "the user associated with the personal access token",
			Destination: &(cfg.UserName),
			Value:       cfg.UserName,
			EnvVars:     []string{"GITLAB_USERNAME"},
		},
		&cli.StringFlag{
			Name:        "gitlab-access-token",
			Usage:       "used to read data from the GitLab API and clone repositories",
			Destination: &(cfg.PersonalAccessToken),
			Value:       cfg.PersonalAccessToken,
			EnvVars:     []string{"GITLAB_ACCESS_TOKEN"},
		},
		&cli.StringSliceFlag{
			Name:        "gitlab-groups",
			Usage:       "restricts operations to listed GitLab groups",
			Destination: cfg.Groups,
			Value:       cfg.Groups,
			EnvVars:     []string{"GITLAB_GROUPS"},
		},
	}

	return cfg, flags
}
