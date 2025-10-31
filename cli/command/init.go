package command

import (
	"docker/isolation"
	"docker/storage"
	"fmt"

	"github.com/urfave/cli/v2"
)

var Init = &cli.Command{
	Name:  "init",
	Usage: "init container process",
	Action: func(ctx *cli.Context) error {
		if err := isolation.InitContainer(storage.Root); err != nil {
			return fmt.Errorf("fail to init container,%v", err)
		}
		return nil
	},
}
