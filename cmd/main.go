package main

import (
	"fmt"
	"os"
	"strconv"

	"protohackers/meanstoanend"
	"protohackers/primetime"
	"protohackers/protos"
	"protohackers/smoketest"
)

var servers = map[int]func(string) (protos.Server, error){
	0: smoketest.Serve,
	1: primetime.Serve,
	2: meanstoanend.Serve,
}

func main() {
	address := os.Getenv("PROTO_ADDRESS")
	if address == "" {
		address = "0.0.0.0"
	}

	problem, _ := strconv.Atoi(os.Getenv("PROTOHACKERS_PROBLEM"))

	server, err := servers[problem](address + ":8080")
	if err != nil {
		fmt.Printf("Failed to start server: %s\n", err)
		os.Exit(1)
	}
	defer server.Close()

	// Wait forever
	select {}
}
