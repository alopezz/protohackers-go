package primetime_test

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"protohackers/primetime"
)

func TestRespondsToMalformedRequestWithMalformedResponseAndDisconnects(t *testing.T) {
	server, err := primetime.Serve("localhost:")
	if err != nil {
		t.Fatalf("Failed to start server: %s\n", err)
	}
	defer server.Close()

	testCases := map[string]string{
		"Malformed JSON":         "{malformed;json.",
		"Missing field":          `{"method":"isPrime"}`,
		"Wrong method name":      `{"method":"isBanana","number":123}`,
		"Non-string method name": `{"method":42,"number":123}`,
		"Invalid number field":   `{"method":"isPrime","number":"not a number"}`,
	}

	for tc, message := range testCases {
		t.Run(tc, func(t *testing.T) {
			conn, err := net.Dial("tcp", server.Addr().String())
			if err != nil {
				t.Fatalf("Error establishing a connection: %s\n", err)
			}
			defer conn.Close()

			deadline := time.Now().Add(5 * time.Second)
			err = conn.SetDeadline(deadline)
			if err != nil {
				t.Fatalf("Failed to set deadline: %s\n", err)
			}

			_, err = conn.Write([]byte(message + "\n"))
			if err != nil {
				t.Fatalf("Error sending data: %s\n", err)
			}

			s := bufio.NewScanner(conn)
			s.Split(bufio.ScanLines)

			if !s.Scan() {
				t.Fatalf("Error while reading data: %s\n", s.Err())
			}

			result := s.Bytes()
			if isWellFormedResponse(result) {
				t.Errorf("Got `%s`, which is well formed, when expecting a malformed response.\n", result)
			}

			// Assert disconnect
			conn.SetDeadline(time.Now().Add(10 * time.Millisecond))
			_, err = conn.Read(make([]byte, 1))
			if !errors.Is(err, io.EOF) {
				t.Errorf("Expected to be disconnected, but was not")
			}
		})
	}
}

func TestRespondsWithCorrectIsPrimeAnswer(t *testing.T) {
	server, err := primetime.Serve("localhost:")
	if err != nil {
		t.Fatalf("Failed to start server: %s\n", err)
	}
	defer server.Close()

	testCases := []struct {
		input    int
		expected bool
	}{
		{input: -1, expected: false},
		{input: 0, expected: false},
		{input: 1, expected: false},
		{input: 2, expected: true},
		{input: 3, expected: true},
		{input: 4, expected: false},
		{input: 5, expected: true},
		{input: 6, expected: false},
		{input: 17, expected: true},
		{input: 97, expected: true},
	}

	conn, err := net.Dial("tcp", server.Addr().String())
	if err != nil {
		t.Fatalf("Error establishing a connection: %s\n", err)
	}

	defer conn.Close()
	deadline := time.Now().Add(5 * time.Second)
	err = conn.SetDeadline(deadline)
	if err != nil {
		t.Fatalf("Failed to set deadline: %s\n", err)
	}

	for _, tc := range testCases {
		_, err = io.WriteString(conn, fmt.Sprintf(`{"method":"isPrime","number":%d}`+"\n", tc.input))
		if err != nil {
			t.Fatalf("Error sending data: %s\n", err)
		}
	}

	s := bufio.NewScanner(conn)
	s.Split(bufio.ScanLines)

	for _, tc := range testCases {
		if !s.Scan() {
			t.Fatalf("Error reading data: %s\n", s.Err())
		}

		result := s.Bytes()
		fmt.Printf("result: %s\n", result)
		if !isWellFormedResponse(result) {
			t.Fatalf("Got `%s`, a malformed response.\n", result)
		}

		response := &struct {
			Method string
			Prime  bool
		}{}
		err := json.Unmarshal(result, &response)
		if err != nil {
			t.Fatalf("Failed to read response JSON, %s\n", err)
		}

		if response.Prime != tc.expected {
			if tc.expected {
				t.Errorf("%d not identified as prime", tc.input)
			} else {
				t.Errorf("%d identified incorrectly as prime", tc.input)
			}
		}
	}
}

func isWellFormedResponse(data []byte) bool {
	response := make(map[string]interface{})
	err := json.Unmarshal(data, &response)
	if err != nil {
		return false
	}

	method, ok := response["method"]
	if !ok || method.(string) != "isPrime" {
		return false
	}

	isPrime, ok := response["prime"]
	if !ok {
		return false
	}
	if _, ok := isPrime.(bool); !ok {
		return false
	}

	return true
}
