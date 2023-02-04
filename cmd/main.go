package main

import (
	"fmt"
	"net"

	"protohackers"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		fmt.Printf("Failed to listen: %s\n", err)
		return
	}
	defer listener.Close()

	protohackers.SmokeTest(listener)
}
