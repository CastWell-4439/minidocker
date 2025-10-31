package cli

import (
	"docker/cli/command"
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"
)

func NewCli() *cli.App {
	app := &cli.App{
		Name: "easydocker",
		Commands: []*cli.Command{
			command.Run,
			command.Ps,
			command.Stop,
			command.Exec,
			command.Init,
		},
		//命令执行前的钩子
		Before: func(ctx *cli.Context) error {
			handler := slog.NewJSONHandler(os.Stdout, nil)
			slog.SetDefault(slog.New(handler))
			return nil
		},
	}
	return app
}
