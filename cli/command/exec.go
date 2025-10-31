package command

import (
	"docker/container"
	"fmt"

	"github.com/urfave/cli/v2"
)

var Exec = &cli.Command{
	Name:  "exec",
	Usage: "execute command",
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() < 2 {
			return fmt.Errorf("empty id or command")
		}
		id := ctx.Args().First()
		cmd := ctx.Args().Slice()[1:]
		return container.Exec(id, cmd)
	},
}
