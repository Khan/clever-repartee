package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
)

type Command struct {
	// Run runs the command. The args are the arguments after the command
	// name.
	Run func(cmd *Command, args []string) error

	// UsageLine is the one-line usage message.
	UsageLine string

	// Short is the short description shown in the 'help' output.
	Short string

	// Long is the long message shown in the 'go help <this-command>' output.
	Long string

	// Logger
	Logger *zap.Logger
}

// Arguments without the executable name
func Run(args []string, logger *zap.Logger) error {
	commands := []*Command{
		VersionCommand(logger),
		DiffCommand(logger),
	}

	var m = make(map[string]*Command)
	for i := range commands {
		cmd := commands[i]
		m[cmd.UsageLine] = cmd
	}

	// if they don't pass a command or only pass "help"
	if len(args) == 0 || (len(args) == 1 && args[0] == "help") {
		logger.Info(
			"clever - Tool for interacting with the Clever API. Available Commands:",
		)
		for i := range commands {
			cmd := *commands[i]
			logger.Info(fmt.Sprintf("%s - %s", cmd.UsageLine, cmd.Short))
		}
		return nil
	}

	arg := args[0]

	var cmd = m[arg]
	if cmd == nil {
		return errors.New(arg + ": invalid command")
	}
	// pass arguments without the executable and without the command itself
	return cmd.Run(cmd, args[1:])
}

func getFlags() []string {
	var args []string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			args = append(args, arg)
		}
	}
	return args
}
