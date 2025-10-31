package command

import (
	"docker/container"
	"errors"

	"github.com/urfave/cli/v2"
)

var Stop = &cli.Command{
	Name:  "stop",
	Usage: "stop container",
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() == 0 {
			return errors.New("empty command")
		}
		ID := ctx.Args().First()
		return container.Stop(ID)
	},
}
