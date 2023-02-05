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

	server, err := smoketest.Serve("localhost:")
	if err != nil {
		t.Fatalf("Failed to start server: %s\n", err)
	}
	defer server.Close()

	conn := assertClient(t, "tcp", server.Addr().String(), message)

	conn.Close()
	server.Close()
}

func TestFiveSimultaneousConnections(t *testing.T) {
	server, err := smoketest.Serve("localhost:")
	if err != nil {
		t.Fatalf("Failed to start server: %s\n", err)
	}
	defer server.Close()

	// Set up 5 clients, assert that we get all echos correctly
	conns := make([]net.Conn, 0, 5)
	for i := 1; i <= 5; i++ {
		conns = append(conns, assertClient(t, "tcp", server.Addr().String(), fmt.Sprintf("Hello n%d\n", i)))
	}
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
