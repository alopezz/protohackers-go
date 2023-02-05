package smoketest

import (
	"fmt"
	"io"
	"net"
	"protohackers/protos"
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

	if _, err := io.Copy(conn, conn); err != nil {
		fmt.Printf("Failed to read: %s", err)
	}
}
