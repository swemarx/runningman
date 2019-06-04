package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"time"
	"strings"
	"strconv"
	"syscall"
	"encoding/json"
	"github.com/pborman/getopt/v2"
)

var (
	buildVersion string			// populated during build-process, see Makefile
	buildTime string			// populated during build-process, see Makefile
	userShell = "/bin/sh -c"
	userDebug = false
	userCommand string
	userHostname string
	userEndpoint string
	userTimeout = 10
)

type report struct {
	CommandLine string `json:"commandline"`
	Username    string `json:"username"`
	Hostname    string `json:"hostname"`
	StartTime   string `json:"starttime"`
	ElapsedTime string `json:"elapsedtime"`
	ExitCode    string `json:"exitcode"`
	Output      string `json:"output"`
}

type envelope struct {
	Message report `json:"message"`
}

func runCommand(cmd string) *report {
	var report report

	report.CommandLine = cmd
	if !getopt.IsSet("hostname") {
		hostname, hostnameErr := os.Hostname()
		if hostnameErr != nil {
			report.Hostname = "<Error>"
		} else {
			report.Hostname = hostname
		}
	} else {
		report.Hostname = userHostname
	}

	username, usernameErr := user.Current()
	if usernameErr != nil {
		report.Username = "<Error>"
	} else {
		report.Username = username.Username
	}

	shellParts := strings.Fields(userShell + " " + cmd)

	startTime := time.Now()
	report.StartTime = startTime.String()
	output, err := exec.Command(shellParts[0], shellParts[1], strings.Join(shellParts[2:], " ")).CombinedOutput()
	if exitError, ok := err.(*exec.ExitError); ok {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			exitCode := status.ExitStatus()
			report.ExitCode = fmt.Sprintf("%d", exitCode)
		} else {
			report.ExitCode = "-1"
		}
	} else {
		report.ExitCode = "0"
	}

	elapsedTime := time.Now().Sub(startTime).Seconds()
	report.ElapsedTime = fmt.Sprintf("%f", elapsedTime)

	if len(string(output)) == 0 {
		report.Output = "<None>\n"
	} else {
		report.Output = string(output)
	}

	if(userDebug) {
		fmt.Printf("[debug] commandline: %s\n", report.CommandLine)
		fmt.Printf("[debug] user: %s\n", report.Username)
		fmt.Printf("[debug] hostname: %s\n", report.Hostname)
		fmt.Printf("[debug] start-time: %s\n", report.StartTime)
		fmt.Printf("[debug] elapsed-time: %s\n", report.ElapsedTime)
		fmt.Printf("[debug] exitcode: %s\n", report.ExitCode)
		fmt.Printf("[debug] see below:\n%s", report.Output)
		fmt.Printf("[debug] endpoint: %s\n", userEndpoint)
		fmt.Printf("[debug] timeout: %d\n", userTimeout)
	}

	return &report
}

func getAndValidateArgs() {
	var (
		timeout string
	)

	getopt.FlagLong(&timeout,      "timeout",  't', "Timeout (in seconds) for report submission")
	getopt.FlagLong(&userCommand,  "command",  'c', "Command to run")
	getopt.FlagLong(&userShell,    "shell",    's', "Shell (default: \"/bin/sh -c\")")
	getopt.FlagLong(&userHostname, "hostname", 'h', "Hostname")
	getopt.FlagLong(&userEndpoint, "endpoint", 'e', "Endpoint where to send report (\"host:port\")")
	getopt.FlagLong(&userDebug,    "debug",    'd', "Debugmode (default: false)")
	getopt.Parse()

	// Check mandatory arguments
	if !getopt.IsSet("command") || !getopt.IsSet("endpoint") {
		fmt.Printf("runningman %s, built %s\n", buildVersion, buildTime)
		getopt.PrintUsage(os.Stdout)
		fmt.Println("\n[error] you need to specify command and endpoint!")
		os.Exit(1)
	}

	// Validate arguments that need it
	if getopt.IsSet("timeout") {
		timeoutAsInt, err := strconv.Atoi(timeout)
		if err != nil || userTimeout < 1 {
			fmt.Println("[error] timeout must be a positive integer")
			os.Exit(1)
		}
		userTimeout = timeoutAsInt
	}
}

func main() {
	getAndValidateArgs()
	var envelope envelope
	envelope.Message = *(runCommand(userCommand))
	jsonOutput, err := json.Marshal(envelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println(string(jsonOutput))
	tcpSend(userEndpoint, jsonOutput, time.Duration(userTimeout) * time.Second)
}
