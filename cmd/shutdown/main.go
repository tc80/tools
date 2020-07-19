package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/cdnjs/tools/sentry"
	"github.com/cdnjs/tools/util"
)

var (
	// initialize standard debug logger
	logger = util.GetStandardLogger()
)

func main() {
	defer sentry.PanicHandler()
	flag.Parse()

	if util.IsDebug() {
		fmt.Println("Running in debug mode")
	}

	switch subcommand := flag.Arg(0); subcommand {
	case "kill":
		{
			kill(flag.Arg(1))
		}
	default:
		panic(fmt.Sprintf("unknown subcommand: `%s`", subcommand))
	}
}

func kill(processID string) {
	ctx := util.ContextWithEntries(util.GetCheckerEntries(processID, logger)...)

	util.Debugf(ctx, "Attempting to send graceful kill signal...\n")
	pid, err := strconv.Atoi(processID)
	util.Check(err)

	process, err := os.FindProcess(pid)
	util.Check(err)

	err = process.Signal(util.ShutdownSignal)
	util.Check(err)

	util.Debugf(ctx, "Sent graceful kill signal...\n")

	state, err := process.Wait()
	util.Check(err)

	fmt.Printf("state is: %v\n", state)

	util.Debugf(ctx, "Success...\n")
}
