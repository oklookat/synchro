package cli

import (
	"os"

	"github.com/urfave/cli/v2"
)

func Boot() error {
	acc := account{}
	tr := transfer{}
	dest := destruct{}

	app := &cli.App{
		Name:  "synchro",
		Usage: "Music streaming utils.",
		Commands: []*cli.Command{
			acc.command(),
			tr.command(),
			dest.command(),
		},
	}

	return app.Run(os.Args)
}
