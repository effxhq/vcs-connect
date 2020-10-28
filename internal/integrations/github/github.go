package github

import (
	"context"
	"github.com/effxhq/vcs-connect/internal/logger"
	"github.com/effxhq/vcs-connect/internal/model"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/google/go-github/v20/github"

	"golang.org/x/oauth2"
)

func NewIntegration(ctx context.Context, config *Configuration) (*Integration, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: config.PersonalAccessToken,
	})
	httpClient := oauth2.NewClient(ctx, tokenSource)

	var client *github.Client
	var err error

	if config.BaseURL != "" && config.UploadURL != "" {
		client, err = github.NewEnterpriseClient(config.BaseURL, config.UploadURL, httpClient)
	} else {
		client = github.NewClient(httpClient)
	}

	if err != nil {
		return nil, err
	}

	return &Integration{
		client: client,
		config: config,
	}, nil
}

type Integration struct {
	client *github.Client
	config *Configuration
}

func (i *Integration) discoverOrganizations(ctx context.Context) ([]string, error) {
	configured := i.config.Organizations.Value()
	if len(configured) > 0 {
		return configured, nil
	}

	organizations := make([]string, 0)
	page := 1

	for page > 0 {
		orgs, resp, err := i.client.Organizations.List(ctx, i.config.UserName, &github.ListOptions{
			Page: 1,
			PerPage: 100,
		})
		if err != nil {
			return nil, err
		}

		results := make([]string, len(orgs))
		for i, org := range orgs {
			results[i] = org.GetLogin()
		}

		organizations = append(organizations, results...)
		page = resp.NextPage
	}

	return organizations, nil
}

func (i *Integration) discoverRepositories(ctx context.Context, organization string) ([]*model.Repository, error) {
	repositories := make([]*model.Repository, 0)

	page := 1
	for page > 0 {
		repos, resp, err := i.client.Repositories.ListByOrg(ctx, organization, &github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{
				Page: page,
				PerPage: 100,
			},
		})
		if err != nil {
			return nil, err
		}

		results := make([]*model.Repository, len(repos))
		for i, repo := range repos {
			results[i] = &model.Repository{
				CloneURL: repo.GetCloneURL(),
				Tags: map[string]string{},
				Annotations: map[string]string{},
			}
		}

		repositories = append(repositories, results...)
		page = resp.NextPage
	}

	return repositories, nil
}

func (i *Integration) Run(ctx context.Context, data chan *model.Repository) error {
	log := logger.MustGetFromContext(ctx)

	organizations, err := i.discoverOrganizations(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to discover organizations from GitHub")
	}

	for _, organization := range organizations {
		log.Info("discovering repositories",
			zap.String("organization", organization))

		repositories, err := i.discoverRepositories(ctx, organization)
		if err != nil {
			log.Error("failed to discover repositories",
				zap.String("organization", organization),
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
