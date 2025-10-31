package command

import (
	"docker/container"

	"github.com/urfave/cli/v2"
)

var Ps = &cli.Command{
	Name:  "ps",
	Usage: "list all containers",
	Action: func(ctx *cli.Context) error {
		return container.ListContainers()
	},
}
