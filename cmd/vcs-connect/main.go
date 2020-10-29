package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"

	"github.com/effxhq/vcs-connect/internal/integrations"
	"github.com/effxhq/vcs-connect/internal/integrations/github"
	"github.com/effxhq/vcs-connect/internal/integrations/gitlab"
	"github.com/effxhq/vcs-connect/internal/logger"
	"github.com/effxhq/vcs-connect/internal/model"
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

type config struct {
	EffxAPIKey string
	ScratchDir string
	Workers    int
}

func runIntegration(parent context.Context, cfg *config, authMethod transport.AuthMethod, integration integrations.Runner) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	defer os.RemoveAll(cfg.ScratchDir)

	ctx = logger.AttachToContext(ctx, logger.MustSetup())

	consumer := &run.Consumer{
		ScratchDir: cfg.ScratchDir,
		AuthMethod: authMethod,
	}

	data := make(chan *model.Repository)

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		defer signal.Stop(signals)
		<-signals
		cancel()
	}()

	for i := 0; i < cfg.Workers; i++ {
		go consumer.Run(ctx, data)
	}

	// Run the integration until completion
	return integration.Run(ctx, data)
}

func main() {
	githubConfig, githubFlags := github.DefaultConfigWithFlags()
	gitlabConfig, gitlabFlags := gitlab.DefaultConfigWithFlags()

	cfg := &config{
		EffxAPIKey: "",
		ScratchDir: path.Join(os.TempDir(), "effx-vcs-connect"),
		Workers:    1,
	}

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "effx-api-key",
			Usage:       "api key used to authenticate with the effx API",
			Destination: &(cfg.EffxAPIKey),
			Value:       cfg.EffxAPIKey,
			EnvVars:     []string{"EFFX_API_KEY"},
		},
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

	app := &cli.App{
		Name:  "vcs-connect",
		Usage: "Index effx.yaml files in connected version control systems.",
		Commands: []*cli.Command{
			{
				Name:  "github",
				Usage: "Index repositories connected via GitHub",
				Flags: append(flags, githubFlags...),
				Action: func(ctx *cli.Context) error {
					integration, err := github.NewIntegration(ctx.Context, githubConfig)
					if err != nil {
						return errors.Wrap(err, "failed to setup GitHub integration")
					}

					authMethod := &http.BasicAuth{
						Username: githubConfig.UserName,
						Password: githubConfig.PersonalAccessToken,
					}

					return runIntegration(ctx.Context, cfg, authMethod, integration)
				},
			},
			{
				Name:  "gitlab",
				Usage: "Index repositories connected via GitLab",
				Flags: append(flags, gitlabFlags...),
				Action: func(ctx *cli.Context) error {
					integration, err := gitlab.NewIntegration(ctx.Context, gitlabConfig)
					if err != nil {
						return errors.Wrap(err, "failed to setup GitLab integration")
					}

					authMethod := &http.BasicAuth{
						Username: gitlabConfig.UserName,
						Password: gitlabConfig.PersonalAccessToken,
					}

					return runIntegration(ctx.Context, cfg, authMethod, integration)
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
