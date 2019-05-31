package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"time"
	"syscall"
	"strconv"
	"encoding/json"
	"github.com/pborman/getopt/v2"
)

var (
	shell = "/bin/sh -c "
	debugMode = false
)

type report struct {
	CommandLine string `json:"commandline"`
	Username string `json:"username"`
	Hostname string `json:"hostname"`
	StartTime string `json:"starttime"`
	ElapsedTime string `json:"elapsedtime"`
	ExitCode string `json:"exitcode"`
	Output string `json:"output"`
}

type envelope struct {
	Message report `json:"message"`
}

func runCommand(cmd string) *report {
	var report report

	report.CommandLine = cmd
	hostname, err := os.Hostname()
	if err != nil {
		report.Hostname = "N/A"
	} else {
		report.Hostname = hostname
	}
	currentUser, err := user.Current()
	if err != nil {
		report.Username = "N/A"
	} else {
		report.Username = currentUser.Username
	}
	startTime := time.Now()
	report.StartTime = startTime.String()
	output, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
	elapsedTime := time.Now().Sub(startTime).Seconds()
	report.ElapsedTime = fmt.Sprintf("%f", elapsedTime)
	report.Output = string(output)
	if exitError, ok := err.(*exec.ExitError); ok {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			//fmt.Printf("EXITCODE: %d\n", status.ExitStatus())
			report.ExitCode = strconv.Itoa(status.ExitStatus())
			//report.ExitCode = fmt.Sprintf("%d", status.ExitStatus())
		} else {
			report.ExitCode = "-1"
		}
	} else {
		//fmt.Println("THIRD ONE")
	}

	if(debugMode) {
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
	getopt.FlagLong(&userCommand, "command",  'c', "Command to run")
	getopt.FlagLong(&debugMode,   "debug",    'd', "Debug-mode")
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
