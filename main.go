package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		fmt.Printf("Failed to listen: %s\n", err)
		return
	}
	defer listener.Close()

	SmokeTest(listener)
}

func SmokeTest(listener net.Listener) {
	fmt.Println("Waiting for client...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				fmt.Printf("Listener closed, exiting.\n")
				break
			}
			fmt.Printf("Failed to accept: %s\n", err)
		}
		go smokeTestHandler(conn)
	}
}

func smokeTestHandler(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Client connected")

	for {
		deadline := time.Now().Add(time.Minute)
		conn.SetDeadline(deadline)
		b := make([]byte, 1024)
		n, err := conn.Read(b)
		if n == 0 || err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("Failed to read, stop reading: %s", err)
			break
		}
		fmt.Printf("Received: %s\n", b[:n])
		conn.Write(b[:n])
	}
}
