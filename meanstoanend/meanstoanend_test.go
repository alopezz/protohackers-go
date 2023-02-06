package meanstoanend_test

import (
	"bytes"
	"io"
	"net"
	"protohackers/meanstoanend"
	"testing"
	"time"
)

func TestRespondsZeroWhenQueriedAfterZeroSamplesGiven(t *testing.T) {
	server, err := meanstoanend.Serve("localhost:")
	if err != nil {
		t.Fatalf("Failed to start server: %s\n", err)
	}
	defer server.Close()

	conn, err := net.Dial("tcp", server.Addr().String())
	if err != nil {
		t.Fatalf("Error establishing a connection: %s", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	err = conn.SetDeadline(deadline)
	if err != nil {
		t.Fatalf("Failed to set deadline: %s\n", err)
	}

	_, err = conn.Write([]byte{'Q', 0, 0, 0x03, 0xe8, 0, 0x01, 0x86, 0xa0})
	if err != nil {
		t.Fatalf("Error sending data: %s\n", err)
	}

	buf := getResponse(t, conn)

	assertResponse(t, buf, []byte{0, 0, 0, 0})
}

func TestCorrectlyReturnsMeanForSingleValue(t *testing.T) {
	server, err := meanstoanend.Serve("localhost:")
	if err != nil {
		t.Fatalf("Failed to start server: %s\n", err)
	}
	defer server.Close()

	conn, err := net.Dial("tcp", server.Addr().String())
	if err != nil {
		t.Fatalf("Error establishing a connection: %s", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	err = conn.SetDeadline(deadline)
	if err != nil {
		t.Fatalf("Failed to set deadline: %s\n", err)
	}

	_, err = conn.Write([]byte{'I', 0, 0, 0x05, 0xe8, 0, 0x01, 0x86, 0xa0})
	if err != nil {
		t.Fatalf("Error sending data: %s\n", err)
	}

	_, err = conn.Write([]byte{'Q', 0, 0, 0x03, 0xe8, 0, 0x01, 0x86, 0xa0})
	if err != nil {
		t.Fatalf("Error sending data: %s\n", err)
	}

	buf := getResponse(t, conn)

	assertResponse(t, buf, []byte{0, 0x01, 0x86, 0xa0})
}

func TestHandlesExampleSessionCorrectly(t *testing.T) {
	server, err := meanstoanend.Serve("localhost:")
	if err != nil {
		t.Fatalf("Failed to start server: %s\n", err)
	}
	defer server.Close()

	conn, err := net.Dial("tcp", server.Addr().String())
	if err != nil {
		t.Fatalf("Error establishing a connection: %s", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	err = conn.SetDeadline(deadline)
	if err != nil {
		t.Fatalf("Failed to set deadline: %s\n", err)
	}

	messages := [][]byte{
		{0x49, 0x00, 0x00, 0x30, 0x39, 0x00, 0x00, 0x00, 0x65},
		{0x49, 0x00, 0x00, 0x30, 0x3a, 0x00, 0x00, 0x00, 0x66},
		{0x49, 0x00, 0x00, 0x30, 0x3b, 0x00, 0x00, 0x00, 0x64},
		{0x49, 0x00, 0x00, 0xa0, 0x00, 0x00, 0x00, 0x00, 0x05},
		{0x51, 0x00, 0x00, 0x30, 0x00, 0x00, 0x00, 0x40, 0x00},
	}

	for _, m := range messages {
		_, err = conn.Write(m)
		if err != nil {
			t.Fatalf("Error sending data: %s\n", err)
		}
	}

	buf := getResponse(t, conn)

	assertResponse(t, buf, []byte{0x00, 0x00, 0x00, 0x65})
}

func getResponse(t *testing.T, conn io.Reader) []byte {
	buf := make([]byte, 4)
	n, err := io.ReadFull(conn, buf)
	if err != nil {
		t.Fatalf("Unable to get the full response: %s", err)
	}
	if n != 4 {
		t.Fatalf("Expected to get 4 bytes in response, got %d", 4)
	}
	return buf
}

func assertResponse(t *testing.T, got []byte, expected []byte) {
	if !bytes.Equal(got, expected) {
		t.Errorf("Error asserting response, got %v, expected %v", got, expected)
	}
}
