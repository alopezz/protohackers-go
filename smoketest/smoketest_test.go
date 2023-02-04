package smoketest_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"protohackers/smoketest"
)

func TestSingleEcho(t *testing.T) {
	message := "Hello World"

	listener, err := net.Listen("tcp", "localhost:")
	if err != nil {
		t.Fatalf("Failed to listen: %s\n", err)
	}
	defer listener.Close()

	go smoketest.SmokeTest(listener)

	conn := assertClient(t, "tcp", listener.Addr().String(), message)

	conn.Close()
	listener.Close()
}

func TestFiveSimultaneousConnections(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:9999")
	if err != nil {
		t.Fatalf("Failed to listen: %s\n", err)
	}
	defer listener.Close()

	go smoketest.SmokeTest(listener)

	// Set up 5 clients, assert that we get all echos correctly
	conns := make([]net.Conn, 0, 5)
	for i := 1; i <= 5; i++ {
		conns = append(conns, assertClient(t, "tcp", listener.Addr().String(), fmt.Sprintf("Hello n%d\n", i)))
	}

	// Only then we go and close connections
	for _, conn := range conns {
		conn.Close()
	}
	listener.Close()
}

func assertClient(t *testing.T, network string, address string, message string) net.Conn {
	conn, err := net.Dial(network, address)
	if err != nil {
		t.Fatalf("Error establishing a connection: %s\n", err)
	}
	defer conn.Close()

	deadline := time.Now().Add(5 * time.Second)
	err = conn.SetDeadline(deadline)
	if err != nil {
		t.Fatalf("Failed to set deadline: %s\n", err)
	}

	_, err = conn.Write([]byte(message))
	if err != nil {
		t.Fatalf("Error sending data: %s\n", err)
	}

	b := make([]byte, 1024)

	n, err := conn.Read(b)
	if err != nil {
		t.Fatalf("Error while reading data: %s\n", err)
	}

	result := string(b[:n])
	if result != message {
		t.Errorf("Expected %s, got %s", message, result)
	}

	return conn
}
