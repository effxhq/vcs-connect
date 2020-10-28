package main

import (
	"context"
	"fmt"
	"github.com/effxhq/vcs-connect/internal/logger"
	"log"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"

	"github.com/effxhq/vcs-connect/internal/integrations"
	"github.com/effxhq/vcs-connect/internal/integrations/github"
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
	Provider   string
	Workers    int
}

func main() {
	githubConfig, githubFlags := github.DefaultConfigWithFlags()
	scratchDir := path.Join(os.TempDir(), "effx-vcs-connect")

	cfg := &config{
		EffxAPIKey: "",
		ScratchDir: "",
		Provider:   "",
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
		&cli.StringFlag{
			Name:        "provider",
			Usage:       "which provider to enable",
			Destination: &(cfg.Provider),
			Value:       cfg.Provider,
			EnvVars:     []string{"PROVIDER"},
		},
		&cli.IntFlag{
			Name:        "workers",
			Usage:       "the number of concurrent repositories cloned at once",
			Destination: &(cfg.Workers),
			Value:       cfg.Workers,
			EnvVars:     []string{"WORKERS"},
		},
	}

	flags = append(flags, githubFlags...)

	app := &cli.App{
		Name:  "vcs-connect",
		Usage: "Index effx.yaml files in connected repositories.",
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Performs a one-time indexing of connected repositories",
				Flags: flags,
				Action: func(clictx *cli.Context) error {
					ctx, cancel := context.WithCancel(clictx.Context)
					defer cancel()

					ctx = logger.AttachContext(ctx, logger.MustSetup())

					var authMethod transport.AuthMethod
					var integration integrations.Runner
					var err error

					switch cfg.Provider {
					case "github":
						authMethod = &http.BasicAuth{
							Username: githubConfig.UserName,
							Password: githubConfig.PersonalAccessToken,
						}

						integration, err = github.NewIntegration(ctx, githubConfig)
						if err != nil {
							return errors.Wrap(err, "failed to setup GitHub integration")
						}
						break

					case "":
						return fmt.Errorf("provider was not specified")
					default:
						return fmt.Errorf("unrecognized provider: %s", cfg.Provider)
					}

					if cfg.ScratchDir == "" {
						cfg.ScratchDir = scratchDir
					}
					defer os.RemoveAll(cfg.ScratchDir)

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
					integration.Run(ctx, data)

					return nil
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
