package smoketest

import (
	"fmt"
	"io"
	"net"
	"protohackers/protos"
	"time"
)

func SmokeTest(listener net.Listener) {
	protos.Serve(listener, smokeTestHandler)
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
