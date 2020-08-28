package cmd

import (
	"fmt"

	"github.com/Khan/clever-repartee/pkg/version"
	"go.uber.org/zap"
)

func VersionCommand(logger *zap.Logger) *Command {
	cmd := &Command{
		UsageLine: "version",
		Short:     "Shows Version",
		Long:      "Shows the version of this binary",
		Run:       Version,
		Logger:    logger,
	}

	return cmd
}

func Version(cmd *Command, args []string) error {
	cmd.Logger.Info(fmt.Sprintf("version %v", version.HumanVersion))
	return nil
}
