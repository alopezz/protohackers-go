package smoketest

import (
	"fmt"
	"io"
	"net"
	"protohackers/protos"
	"time"
)

func Serve(address string) (protos.Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("Failed to listen: %s\n", err)
	}
	go protos.Serve(listener, handler)

	return &server{Listener: listener}, nil
}

type server struct {
	net.Listener
}

func handler(conn net.Conn) {
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
