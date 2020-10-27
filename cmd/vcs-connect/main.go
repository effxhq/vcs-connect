package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/effxhq/vcs-connect/internal/v"

	"github.com/urfave/cli/v2"
)

// variables set by build using -X ldflag
var version string
var commit string
var date string

type config struct {
	placeholder string
}

func main() {
	cfg := &config{
		placeholder: "default value",
	}

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "placeholder",
			Usage:       "description of the flag",
			Destination: &(cfg.placeholder),
			Value:       cfg.placeholder,
			EnvVars:     []string{"PLACEHOLDER"},
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
					// TODO
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
