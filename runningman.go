package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
	"syscall"
	"github.com/pborman/getopt/v2"
)

var (
	shell = "/bin/sh -c "
)

func runCommand(cmd string) {
	var (
		exitCode int
		startTime time.Time
		elapsedTime time.Duration
	)

	startTime = time.Now()
	out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	elapsedTime = time.Now().Sub(startTime)
	if exitError, ok := err.(*exec.ExitError); ok {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			exitCode = status.ExitStatus()
		}
	}
	fmt.Printf("[debug:output] %s\n", out)
	fmt.Printf("[debug] Command returned exit-code: %d\n", exitCode)
	fmt.Printf("[debug] Command execution time: %.1f seconds\n", elapsedTime.Seconds())
}

func getCommandLineArgs() string {
	var userCommand string
	getopt.FlagLong(&userCommand, "command",  'c', "Command to run")
	getopt.Parse()
	if !getopt.IsSet("command") {
		getopt.PrintUsage(os.Stdout)
		fmt.Println("\n[error] you need to specify command!")
		os.Exit(1)
	}
	return userCommand
}

func main() {
	userCommand := getCommandLineArgs()
	//cmd := shellescape.Quote(userCommand)
	runCommand(userCommand)
}
