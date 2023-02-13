package unusualdatabase

import (
	"fmt"
	"net"
	"protohackers/protos"
	"strings"
)

func Serve(address string) (protos.Server, error) {
	conn, err := net.ListenPacket("udp", address)
	if err != nil {
		return nil, fmt.Errorf("Failed to listen: %s\n", err)
	}

	go func() {
		kvStore := make(map[string]string)
		p := make([]byte, 1000)

		for {
			n, addr, err := conn.ReadFrom(p)
			if err != nil {
				fmt.Printf("Failed to read: %s\n", err)
			}

			msg := string(p[:n])

			// Insert
			if strings.Contains(msg, "=") {
				kvPair := strings.SplitN(msg, "=", 2)
				key, value := kvPair[0], kvPair[1]
				fmt.Printf("Storing value `%s` under key `%s`\n", value, key)
				kvStore[key] = value
				continue
			}
			// Retrieve
			var response []byte

			// Version
			if msg == "version" {
				response = []byte("version=v1.0")
			} else {
				response = []byte(msg + "=" + kvStore[msg])
			}

			fmt.Printf("Requesting value for key `%s`, responding with `%s`\n", msg, response)
			conn.WriteTo(response, addr)
		}
	}()

	return &server{PacketConn: conn}, nil
}

type server struct {
	net.PacketConn
}

func (s *server) Addr() net.Addr {
	return s.PacketConn.LocalAddr()
}
