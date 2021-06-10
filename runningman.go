package main

import (
	"fmt"
	"net"
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
	userJobName string
	userCommand string
	userHostname string
	userLogfile string
	userEndpoint string
	userLogOutput = false
	userTimeout = 10
	userDebug = false
	userLogScreen = false
)

type report struct {
	JobName		string `json:"jobname"`
	CommandLine string `json:"commandline"`
	Username    string `json:"username"`
	Hostname    string `json:"hostname"`
	StartTime   string `json:"starttime"`
	ElapsedSecs string `json:"elapsedsecs"`
	ExitCode    string `json:"exitcode"`
	LogOutput	string `json:"output"`
	Output      string `json:"output"`
}

type envelope struct {
	result report `json:"result"`
}

func runCommand(cmd string) *report {
	var report report

	report.JobName = userJobName
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
	//report.StartTime = startTime.Format(time.RFC3339)
	report.StartTime = startTime.Format("2006-01-02 15:04:05")
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

	elapsedSeconds := time.Now().Sub(startTime).Seconds()
	report.ElapsedSecs = fmt.Sprintf("%f", elapsedSeconds)

	report.LogOutput = strconv.FormatBool(userLogOutput)
	if userLogOutput {
		if len(string(output)) == 0 {
			report.Output = "<None>\n"
		} else {
			report.Output = string(output)
		}
	}

	if(userDebug) {
		fmt.Printf("[debug] jobname: %s\n", report.JobName)
		fmt.Printf("[debug] commandline: %s\n", report.CommandLine)
		fmt.Printf("[debug] user: %s\n", report.Username)
		fmt.Printf("[debug] hostname: %s\n", report.Hostname)
		fmt.Printf("[debug] start-time: %s\n", report.StartTime)
		fmt.Printf("[debug] elapsed-seconds: %s\n", report.ElapsedSecs)
		fmt.Printf("[debug] exitcode: %s\n", report.ExitCode)
		fmt.Printf("[debug] log-output: %s\n", report.LogOutput)
		fmt.Printf("[debug] output below:\n%s", report.Output)
		if len(userEndpoint) > 0 {
			fmt.Printf("[debug] endpoint: %s\n", userEndpoint)
		} else {
			fmt.Printf("[debug] logfile: %s\n", userLogfile)
		}
		fmt.Printf("[debug] timeout: %d\n", userTimeout)
	}

	return &report
}

func getAndValidateArgs() {
	var (
		timeout string
	)

	getopt.FlagLong(&userJobName,  "name",           'n', "Name of job")
	getopt.FlagLong(&userShell,    "shell",          'S', "Shell (default: \"/bin/sh -c\")")
	getopt.FlagLong(&userLogfile,  "logfile",        'f', "Log to local file")
	getopt.FlagLong(&userCommand,  "command",        'c', "Command to run")
	getopt.FlagLong(&userHostname, "hostname",       'h', "Hostname")
	getopt.FlagLong(&userEndpoint, "endpoint",       'e', "Endpoint where to send report (\"host:port\")")
	getopt.FlagLong(&timeout,      "timeout",        't', "Timeout (in seconds) for endpoint submission")
	getopt.FlagLong(&userDebug,    "debug",          'd', "Debugmode (default: false)")
	getopt.FlagLong(&userLogScreen,"logscreen",      's', "Log to stdout")
	getopt.FlagLong(&userLogOutput,"capture-output", 'o', "Include output in log-file (default: false)")
	getopt.Parse()

	// Check mandatory arguments
	if !getopt.IsSet("command") || !getopt.IsSet("name") || (!getopt.IsSet("endpoint") && (!getopt.IsSet("logfile"))) {
		fmt.Printf("runningman %s, built %s\n", buildVersion, buildTime)
		getopt.PrintUsage(os.Stdout)
		fmt.Println("\n[error] you must specify command, name and endpoint and/or logfile!")
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

func tcpSend(socket string, data []byte, timeout time.Duration) {
	conn, err := net.DialTimeout("tcp", socket, timeout)
	if err != nil {
		fmt.Printf("[error] %s\n", err.Error())
		os.Exit(1)
	}

	_, err = conn.Write([]byte(data))
	if err != nil {
		fmt.Printf("[error] could not send data: %s\n", err.Error())
		os.Exit(1)
	}

	if userDebug {
		fmt.Print("[debug] Sent data to endpoint successfully!\n")
	}

	conn.Close()
}

func prepareReportForOutput(r report) string {
	s := "# " + r.StartTime + " " +
		"name:" + r.JobName + " " +
		"host:" + r.Hostname + " " +
		"user:" + r.Username + " " +
		"elapsed:" + r.ElapsedSecs + " " +
		"exitcode:" + r.ExitCode + " " +
		"log-output:" + r.LogOutput + "\n"

	if userLogOutput {
		s += r.Output
	}

	return s
}

// Writes a report to file as a single-line. Omits the output.
func writeLog(path string, output string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("[error] could not open logfile %s: %s\n", path, err.Error()) 
	}
	defer f.Close()
	if _, err := f.WriteString(output); err != nil {
		fmt.Printf("[error] could not write to logfile %s: %s\n", path, err.Error())
	}
}

func main() {
	getAndValidateArgs()

	var envelope envelope
	envelope.result = *(runCommand(userCommand))

	// If sending over the network, turn it into json.
	if getopt.IsSet("endpoint") {
		jsonOutput, err := json.Marshal(envelope)
		if err != nil {
			fmt.Println(err)
			return
		}
		//fmt.Println(string(jsonOutput))
		tcpSend(userEndpoint, jsonOutput, time.Duration(userTimeout) * time.Second)
	}

	// If logging to file or screen, format it appropriately.
	if getopt.IsSet("logfile") || getopt.IsSet("logscreen") {
		s := prepareReportForOutput(envelope.result)

		if getopt.IsSet("logfile") {
			// open file and output
			writeLog(userLogfile, s)
		}

		if getopt.IsSet("logfile") {
			fmt.Printf("%s", s)
		}

		if getopt.IsSet("logscreen") {
		}
	}
}
