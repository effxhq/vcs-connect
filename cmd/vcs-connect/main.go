package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/effxhq/vcs-connect/internal/run"
	"github.com/effxhq/vcs-connect/internal/v"

	"github.com/urfave/cli/v2"
)

// variables set by build using -X ldflag
var version string
var commit string
var date string

type config struct {
	ScratchDir string
}

func main() {
	cfg := &config{
		ScratchDir: path.Join(os.TempDir(), "effx-vcs-connect"),
	}

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "scratch-dir",
			Usage:       "scratch space used to clone repositories for indexing",
			Destination: &(cfg.ScratchDir),
			Value:       cfg.ScratchDir,
			EnvVars:     []string{"SCRATCH_DIR"},
		},
	}

	app := &cli.App{
		Name:  "vcs-connect",
		Usage: "Index effx.yaml files in connected repositories.",
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Performs a one-time indexing of connected repositories",
				Flags: flags,
				Action: func(ctx *cli.Context) error {
					_ = &run.Consumer{
						ScratchDir: cfg.ScratchDir,
					}

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
