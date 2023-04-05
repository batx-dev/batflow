package cmd

import (
	"github.com/urfave/cli/v2"
)

func GetRootCommand() *cli.App {
	app := cli.NewApp()
	app.Name = "batflow"
	app.Usage = "Manages batch workflow"

	app.Commands = []*cli.Command{
		getWorkerCommand(),
	}

	return app
}
