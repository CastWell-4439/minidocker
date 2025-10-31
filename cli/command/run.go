package command

import (
	"docker/container"
	"fmt"

	"github.com/urfave/cli/v2"
)

var Run = &cli.Command{
	Name:  "run",
	Usage: "run container",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "container name",
		},
		&cli.StringFlag{
			Name:  "image",
			Usage: "container image",
			Value: "busybox",
		},
		&cli.BoolFlag{
			Name:  "it",
			Usage: "interactive mode",
		},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() == 0 {
			return fmt.Errorf("empty args")
		}
		cmd := ctx.Args().Slice()
		interactive := ctx.Bool("it")
		containers := ctx.String("name")
		image := ctx.String("image")

		_, ans := container.Run(containers, image, cmd, interactive)
		return ans
	},
}
