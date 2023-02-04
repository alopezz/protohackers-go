package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"protohackers"
)

type protocolImplementation func(net.Listener)

var protos = map[int]protocolImplementation{
	0: protohackers.SmokeTest,
	1: protohackers.PrimeTime,
}

func main() {
	address := os.Getenv("PROTO_ADDRESS")
	if address == "" {
		address = "0.0.0.0"
	}
	
	listener, err := net.Listen("tcp", address + ":8080")
	if err != nil {
		fmt.Printf("Failed to listen: %s\n", err)
		return
	}
	defer listener.Close()

	problem, _ := strconv.Atoi(os.Getenv("PROTOHACKERS_PROBLEM"))

	protos[problem](listener)
}
