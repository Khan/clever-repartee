package main

import (
	"os"

	"github.com/Khan/clever-repartee/cmd"

	"go.uber.org/zap"
)

const (
	// exitFail is the exit code if the program
	// fails.
	exitFail = 1
	// exitSuccess is the exit code if the program succeeds
	exitSuccess = 0
)

// https://pace.dev/blog/2020/02/12/why-you-shouldnt-use-func-main-in-golang-by-mat-ryer
func main() {

	logger, err := zap.NewDevelopment()

	if err != nil {
		panic(err)
	}

	// pass all arguments without the executable name
	if err := cmd.Run(os.Args[1:], logger); err != nil {
		logger.Error("%s\n", zap.Error(err))
		os.Exit(exitFail)
	} else {
		logger.Info("Successful completion")
		os.Exit(exitSuccess)
	}
}
