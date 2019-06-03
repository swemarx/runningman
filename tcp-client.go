package main

import (
	"os"
	"net"
	"fmt"
	"time"
)

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
		fmt.Print("[debug] Sent data successfully!\n")
	}

	conn.Close()
}
