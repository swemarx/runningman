# runningman

runningmans purpose is to run jobs, capture the results and submit them (as JSON) to Logstash. We wanted to minimize the use of emails and eliminate the need to scan mailboxes for results of running important jobs. We also wanted to record other metrics such as execution-time, exitcode and output for any kind of shell-expression/script/binary and make it searchable using a standard ELK-setup.

## Getting started

```
git clone https://github.com/swemarx/runningman
cd runningman
make
```

## Prerequisites

golang (I've used 1.10.3)
make
Logstash setup somewhere accessible as en endpoint (the --endpoint option)

### Installing

```
copy runningman /wherever/
```

### Running

Getting help
```
$ ./runningman
runningman f519da9, built 2019-06-03T14:33:21Z
Usage: runningman [-d] [-c value] [-e value] [-h value] [-s value] [-t value] [parameters ...]
 -c, --command=value
                    Command to run
 -d, --debug        Debugmode (default: false)
 -e, --endpoint=value
                    Endpoint where to send report ("host:port")
 -h, --hostname=value
                    Hostname
 -s, --shell=value  Shell (default: "/bin/sh -c")
 -t, --timeout=value
                    Timeout (in seconds) for report submission

[error] you need to specify command and endpoint!
```

Note that -c/--command and -e/--endpoint are mandatory!

Running a shell-expression and submitting the result to a Logstash running at 172.20.16.20:1987, activating debug-output to see what happens
```
$ ./runningman --endpoint 172.20.16.20:1987 --command 'echo Writing to stderr >> /dev/stderr; sleep 1; echo Writing to stdout' --debug
[debug] commandline: echo Writing to stderr >> /dev/stderr; sleep 1; echo Writing to stdout
[debug] user: swemarx
[debug] hostname: tinydancer
[debug] start-time: 2019-06-04 09:36:16.642672092 +0200 CEST m=+0.001305031
[debug] elapsed-time: 1.013994
[debug] exitcode: 0
[debug] see below:
Writing to stderr
Writing to stdout
[debug] endpoint: 172.20.16.20:1987
[debug] timeout: 10
[debug] Sent data successfully!
```

