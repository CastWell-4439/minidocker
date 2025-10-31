package easydocker

import (
	"docker/cli"
	"log/slog"
	"os"
)

func main() {
	app := cli.NewCli()
	if err := app.Run(os.Args); err != nil {
		slog.Error("fail to run cli", "error", err)
		os.Exit(1)
	}
}
