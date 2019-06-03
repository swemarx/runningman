package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"time"
	"strings"
	"syscall"
	"encoding/json"
	"github.com/pborman/getopt/v2"
)

var (
	userShell = "/bin/sh -c"
	userDebug = false
	userHostname string
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
			report.Hostname = "N/A"
		} else {
			report.Hostname = hostname
		}
	} else {
		report.Hostname = userHostname
	}

	username, usernameErr := user.Current()
	if usernameErr != nil {
		report.Username = "N/A"
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
	report.Output = string(output)

	if(userDebug) {
		fmt.Printf("[debug] commandline: %s\n", report.CommandLine)
		fmt.Printf("[debug] user: %s\n", report.Username)
		fmt.Printf("[debug] hostname: %s\n", report.Hostname)
		fmt.Printf("[debug] start-time: %s\n", report.StartTime)
		fmt.Printf("[debug] elapsed-time: %s\n", report.ElapsedTime)
		fmt.Printf("[debug] exitcode: %s\n", report.ExitCode)
		fmt.Printf("[debug] see below:\n%s", report.Output)
	}

	return &report
}

func getCommandLineArgs() string {
	var userCommand string
	getopt.FlagLong(&userShell,    "shell",    's', "Shell (default: \"/bin/sh -c\")")
	getopt.FlagLong(&userCommand,  "command",  'c', "Command to run")
	getopt.FlagLong(&userHostname, "hostname", 'h', "Hostname")
	getopt.FlagLong(&userDebug,    "debug",    'd', "Debugmode (default: false)")
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
	var envelope envelope
	envelope.Message = *(runCommand(userCommand))
	jsonOutput, err := json.Marshal(envelope)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(jsonOutput))
}
