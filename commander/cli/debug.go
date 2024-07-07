package cli

import (
	"github.com/urfave/cli/v2"
)

type debug struct {
}

func (e debug) command() *cli.Command {
	return &cli.Command{
		Name:        "debug",
		Aliases:     []string{"deb"},
		Subcommands: []*cli.Command{},
		Usage:       "Debug actions",
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}
