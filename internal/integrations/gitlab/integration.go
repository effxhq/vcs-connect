package gitlab

import (
	"context"

	"github.com/effxhq/vcs-connect/internal/logger"
	"github.com/effxhq/vcs-connect/internal/model"

	"github.com/pkg/errors"

	"github.com/xanzy/go-gitlab"

	"go.uber.org/zap"
)

// NewIntegration returns the Integration responsible for communicating with GitLab.
// Before construction, the Configuration is validated to ensure it contains the
// proper information.
func NewIntegration(ctx context.Context, config *Configuration) (*Integration, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	options := make([]gitlab.ClientOptionFunc, 0)
	if config.BaseURL != "" {
		options = append(options, gitlab.WithBaseURL(config.BaseURL))
	}

	client, err := gitlab.NewClient(config.PersonalAccessToken, options...)
	if err != nil {
		return nil, err
	}

	return &Integration{
		client: client,
		config: config,
	}, nil
}

// Integration encapsulates the logic for integrating data from GitLab.
type Integration struct {
	client *gitlab.Client
	config *Configuration
}

func (i *Integration) discoverGroups(ctx context.Context) ([]string, error) {
	configured := i.config.Groups.Value()
	if len(configured) > 0 {
		return configured, nil
	}

	groups := make([]string, 0)
	page := 1

	for page > 0 {
		grps, resp, err := i.client.Groups.ListGroups(&gitlab.ListGroupsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		})
		if err != nil {
			return nil, err
		}

		results := make([]string, len(groups))
		for i, grp := range grps {
			results[i] = grp.FullPath
		}

		groups = append(groups, results...)
		page = resp.NextPage
	}

	return groups, nil
}

func (i *Integration) discoverRepositories(ctx context.Context, group string) ([]*model.Repository, error) {
	repositories := make([]*model.Repository, 0)

	page := 1
	for page > 0 {
		repos, resp, err := i.client.Groups.ListGroupProjects(group, &gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		})
		if err != nil {
			return nil, err
		}

		results := make([]*model.Repository, len(repos))
		for i, repo := range repos {
			results[i] = &model.Repository{
				CloneURL:    repo.HTTPURLToRepo,
				Tags:        map[string]string{},
				Annotations: map[string]string{},
			}
		}

		repositories = append(repositories, results...)
		page = resp.NextPage
	}

	return repositories, nil
}

// Run feeds the data channel with results it discovers from GitLab
func (i *Integration) Run(ctx context.Context, data chan *model.Repository) error {
	log := logger.MustGetFromContext(ctx)

	groups, err := i.discoverGroups(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to discover groups from GitLab")
	}

	for _, group := range groups {
		log.Info("discovering repositories",
			zap.String("group", group))

		repositories, err := i.discoverRepositories(ctx, group)
		if err != nil {
			log.Error("failed to discover repositories",
				zap.String("group", group),
				zap.Error(err))
			continue
		}

		// push to consumers or stop if cancelled
		for _, repository := range repositories {
			log.Info("processing repository",
				zap.String("repository", repository.CloneURL))

			select {
			case <-ctx.Done():
				return nil
			case data <- repository:
				continue
			}
		}
	}

	return nil
}
