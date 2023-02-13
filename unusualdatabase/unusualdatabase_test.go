package unusualdatabase_test

import (
	"fmt"
	"net"
	"protohackers/unusualdatabase"
	"testing"
	"time"
)

func TestInsertAndRetrieveDifferentKeys(t *testing.T) {
	server, err := unusualdatabase.Serve("localhost:")
	if err != nil {
		t.Fatalf("Failed to start server: %s\n", err)
	}
	defer server.Close()

	client, err := StartUDPClient(server.Addr().String())
	if err != nil {
		t.Fatalf("Error starting UDP client for testing: %s", err)
	}

	client.SendPacket([]byte("key=value"))
	client.SendPacket([]byte("foo=bar"))
	client.SendPacket([]byte("=empty"))

	var p []byte

	client.SendPacket([]byte("foo"))
	p = client.ReadPacket()
	assertPacket(t, "foo=bar", p)

	client.SendPacket([]byte("key"))
	p = client.ReadPacket()
	assertPacket(t, "key=value", p)

	client.SendPacket([]byte(""))
	p = client.ReadPacket()
	assertPacket(t, "=empty", p)
}

func TestKeysAreOverwritten(t *testing.T) {
	server, err := unusualdatabase.Serve("localhost:")
	if err != nil {
		t.Fatalf("Failed to start server: %s\n", err)
	}
	defer server.Close()

	client, err := StartUDPClient(server.Addr().String())
	if err != nil {
		t.Fatalf("Error starting UDP client for testing: %s", err)
	}

	client.SendPacket([]byte("foo=bar"))
	client.SendPacket([]byte("foo=bar=baz"))

	var p []byte
	client.SendPacket([]byte("foo"))
	p = client.ReadPacket()
	assertPacket(t, "foo=bar=baz", p)
}

func TestVersionAlreadyExistsAndCannotBeOverwritten(t *testing.T) {
	server, err := unusualdatabase.Serve("localhost:")
	if err != nil {
		t.Fatalf("Failed to start server: %s\n", err)
	}
	defer server.Close()

	client, err := StartUDPClient(server.Addr().String())
	if err != nil {
		t.Fatalf("Error starting UDP client for testing: %s", err)
	}

	client.SendPacket([]byte("version"))
	version := client.ReadPacket()

	var p []byte
	client.SendPacket([]byte("version=foo"))
	client.SendPacket([]byte("version"))
	p = client.ReadPacket()
	assertPacket(t, string(version), p)
	client.SendPacket([]byte("version=bar"))
	client.SendPacket([]byte("version"))
	p = client.ReadPacket()
	assertPacket(t, string(version), p)
}

func assertPacket(t *testing.T, expected string, packet []byte) {
	if string(packet) != expected {
		t.Fatalf("Expected `%s` but got `%s`", expected, string(packet))
	}
}

type UDPClient struct {
	net.PacketConn
	msgChan chan []byte
	address net.Addr
}

func StartUDPClient(address string) (*UDPClient, error) {
	conn, err := net.ListenPacket("udp", "localhost:")
	if err != nil {
		return nil, fmt.Errorf("Failed to listen: %s\n", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	err = conn.SetDeadline(deadline)
	if err != nil {
		return nil, fmt.Errorf("Failed to set deadline: %s\n", err)
	}

	ch := make(chan []byte, 10)

	go func() {
		for {
			p := make([]byte, 1000)
			n, _, err := conn.ReadFrom(p)
			if err != nil {
				continue
			}
			ch <- p[:n]
		}
	}()

	addr, _ := net.ResolveUDPAddr("udp", address)
	return &UDPClient{PacketConn: conn, msgChan: ch, address: addr}, nil
}

func (c *UDPClient) ReadPacket() []byte {
	select {
	case p := <-c.msgChan:
		return p
	case <-time.After(100 * time.Millisecond):
		return []byte{}
	}
}

func (c *UDPClient) SendPacket(p []byte) {
	c.PacketConn.WriteTo(p, c.address)
}
