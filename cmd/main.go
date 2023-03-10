package main

import (
	"fmt"
	"os"
	"strconv"

	"protohackers/budgetchat"
	"protohackers/meanstoanend"
	"protohackers/mobinthemiddle"
	"protohackers/primetime"
	"protohackers/protos"
	"protohackers/smoketest"
	"protohackers/unusualdatabase"
)

const UPSTREAM_BUDGETCHAT_ADDRESS = "chat.protohackers.com:16963"

var servers = map[int]func(string) (protos.Server, error){
	0: smoketest.Serve,
	1: primetime.Serve,
	2: meanstoanend.Serve,
	3: budgetchat.Serve,
	4: unusualdatabase.Serve,
	5: func(addr string) (protos.Server, error) {
		return mobinthemiddle.Serve(addr, UPSTREAM_BUDGETCHAT_ADDRESS)
	},
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
