package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/effxhq/vcs-connect/internal/controller"
	"github.com/effxhq/vcs-connect/internal/effx"
	"github.com/effxhq/vcs-connect/internal/integrations/github"
	"github.com/effxhq/vcs-connect/internal/integrations/gitlab"
	"github.com/effxhq/vcs-connect/internal/run"
	"github.com/effxhq/vcs-connect/internal/v"

	"github.com/pkg/errors"

	"github.com/urfave/cli/v2"

	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

// variables set by build using -X ldflag
var version string
var commit string
var date string

func initAuthForGitHub(cfg *github.Configuration) transport.AuthMethod {
	return &http.BasicAuth{
		Username: cfg.UserName,
		Password: cfg.PersonalAccessToken,
	}
}

func initAuthForGitLab(cfg *gitlab.Configuration) transport.AuthMethod {
	return &http.BasicAuth{
		Username: cfg.UserName,
		Password: cfg.PersonalAccessToken,
	}
}

func main() {
	clientConfig, clientFlags := effx.DefaultConfigWithFlags()
	githubConfig, githubFlags := github.DefaultConfigWithFlags()
	gitlabConfig, gitlabFlags := gitlab.DefaultConfigWithFlags()
	controllerConfig, controllerFlags := controller.DefaultConfigWithFlags()

	flags := append(controllerFlags, clientFlags...)

	app := &cli.App{
		Name:  "vcs-connect",
		Usage: "Index effx.yaml files in connected version control systems.",
		Commands: []*cli.Command{
			{
				Name:  "github",
				Usage: "Index repositories connected via GitHub",
				Flags: append(flags, githubFlags...),
				Action: func(ctx *cli.Context) error {
					effxClient, err := effx.New(clientConfig)
					if err != nil {
						return errors.Wrapf(err, "failed to setup effx client")
					}

					integration, err := github.NewIntegration(ctx.Context, githubConfig)
					if err != nil {
						return errors.Wrap(err, "failed to setup GitHub integration")
					}

					consumer := &run.Consumer{
						EffxClient: effxClient,
						ScratchDir: controllerConfig.ScratchDir,
						AuthMethod: initAuthForGitHub(githubConfig),
					}

					control, err := controller.New(controllerConfig, integration, consumer)
					if err != nil {
						return errors.Wrapf(err, "failed to setup controller")
					}

					return control.Run(ctx.Context)
				},
			},
			{
				Name:  "gitlab",
				Usage: "Index repositories connected via GitLab",
				Flags: append(flags, gitlabFlags...),
				Action: func(ctx *cli.Context) error {
					effxClient, err := effx.New(clientConfig)
					if err != nil {
						return errors.Wrapf(err, "failed to setup effx client")
					}

					integration, err := gitlab.NewIntegration(ctx.Context, gitlabConfig)
					if err != nil {
						return errors.Wrap(err, "failed to setup GitLab integration")
					}

					consumer := &run.Consumer{
						EffxClient: effxClient,
						ScratchDir: controllerConfig.ScratchDir,
						AuthMethod: initAuthForGitHub(githubConfig),
					}

					control, err := controller.New(controllerConfig, integration, consumer)
					if err != nil {
						return errors.Wrapf(err, "failed to setup controller")
					}

					return control.Run(ctx.Context)
				},
			},
			{
				Name:  "version",
				Usage: "Outputs information about the binary",
				Action: func(ctx *cli.Context) error {
					fmt.Println(fmt.Sprintf("vcs-connect %#v", v.Info{
						Version: version,
						Commit:  commit,
						Date:    date,
						OS:      runtime.GOOS,
						Arch:    runtime.GOARCH,
					}))
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err.Error())
	}
}
