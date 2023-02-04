package protos

import (
	"errors"
	"fmt"
	"net"
)

type ConnHandler func(conn net.Conn)

func Serve(listener net.Listener, handle ConnHandler) {
	fmt.Println("Waiting for client...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				fmt.Printf("Listener closed, exiting.\n")
				break
			}
			fmt.Printf("Failed to accept: %s\n", err)
		}
		go handle(conn)
	}
}
