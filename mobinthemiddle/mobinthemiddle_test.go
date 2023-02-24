package mobinthemiddle_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"protohackers/budgetchat"
	"protohackers/mobinthemiddle"
	"strings"
	"testing"
	"time"
)

func TestTamper(t *testing.T) {
	src := &closableBuffer{Buffer: *bytes.NewBuffer([]byte(
		`please send payment to 7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX
nothing to see here
7adNeSwJkMakpEcln9HEtthSRtxdmEHOT8T is my address
This is a product ID, not a Boguscoin: 79pHTh5sAWLiWsHTrmLlrqb9gE-7Mip5ZmCzZphpkdQbpmeW9EtcFr2lsoC-1234
Please pay the ticket price of 15 Boguscoins to one of these addresses: 7YWHMfk9JZe0LM0g1ZauHuiSxhI 7X5WTgxTYA5lugCy7FrhI1uA51
this message doesn't end in a new line, so it shouldn't be relayed`))}

	dst := &closableBuffer{}

	mobinthemiddle.Tamper(src, dst)

	if !dst.wasClosed {
		t.Errorf("The destination was not closed")
	}

	expected := `please send payment to 7YWHMfk9JZe0LM0g1ZauHuiSxhI
nothing to see here
7YWHMfk9JZe0LM0g1ZauHuiSxhI is my address
This is a product ID, not a Boguscoin: 79pHTh5sAWLiWsHTrmLlrqb9gE-7Mip5ZmCzZphpkdQbpmeW9EtcFr2lsoC-1234
Please pay the ticket price of 15 Boguscoins to one of these addresses: 7YWHMfk9JZe0LM0g1ZauHuiSxhI 7YWHMfk9JZe0LM0g1ZauHuiSxhI
`
	if dst.Buffer.String() != expected {
		t.Errorf("Expected `%s` but got `%s`", expected, dst.Buffer.String())
	}
}

type closableBuffer struct {
	bytes.Buffer
	wasClosed bool
}

func (b *closableBuffer) Close() error {
	b.wasClosed = true
	return nil
}

func TestExampleSession(t *testing.T) {
	budgetchatServer, err := budgetchat.Serve("localhost:")
	if err != nil {
		t.Fatalf("Failed to start server: %s\n", err)
	}
	defer budgetchatServer.Close()

	proxy, err := mobinthemiddle.Serve("localhost:", budgetchatServer.Addr().String())
	if err != nil {
		t.Fatalf("Failed to start proxy: %s\n", err)
	}
	defer proxy.Close()

	// Create clients
	clientNames := []string{"alice", "bob"}
	clients := make(map[string]*client, len(clientNames))

	// alice is connected to the real server
	clients["alice"], err = makeClient(budgetchatServer.Addr().String())
	if err != nil {
		t.Fatalf("Error constructing client: %s", err)
	}
	defer clients["alice"].Close()

	// alice identifies herself with the server
	_, err = clients["alice"].Recv()
	if err != nil {
		t.Fatal(err)
	}

	err = clients["alice"].Send("alice")
	if err != nil {
		t.Fatal(err)
	}

	// alice receives the message with the empty user list
	_, err = clients["alice"].Recv()
	if err != nil {
		t.Fatal(err)
	}

	// bob connects to the proxy
	clients["bob"], err = makeClient(proxy.Addr().String())
	if err != nil {
		t.Fatalf("Error constructing client: %s", err)
	}
	defer clients["bob"].Close()

	_, err = clients["bob"].Recv()
	if err != nil {
		t.Fatal(err)
	}

	err = clients["bob"].Send("bob")
	if err != nil {
		t.Fatal(err)
	}

	// Receive message with existing users (just alice)
	msg, err := clients["bob"].Recv()
	if err != nil {
		t.Fatal(err)
	}
	assertServerMessage(t, msg, "alice")

	// alice receives the message of bob joining
	msg, err = clients["alice"].Recv()
	if err != nil {
		t.Fatal(err)
	}
	assertServerMessage(t, msg, "bob")

	// bob asks for payment
	err = clients["bob"].Send("Hi alice, please send payment to 7iKDZEwPZSqIvDnHvVN2r0hUWXD5rHX")
	if err != nil {
		t.Fatal(err)
	}

	// alice receives the message with the address replaced by Tony's
	msg, err = clients["alice"].Recv()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(msg, "Hi alice, please send payment to 7YWHMfk9JZe0LM0g1ZauHuiSxhI") {
		t.Fatalf("Expected `%s` to contain the message with the modified address", msg)
	}
}

// A chat client
type client struct {
	net.Conn
	*bufio.Scanner
}

func (c *client) Send(message string) error {
	return writeLine(c.Conn, message)
}

func (c *client) Recv() (string, error) {
	if c.Scanner == nil {
		s := bufio.NewScanner(c.Conn)
		s.Split(bufio.ScanLines)
		c.Scanner = s
	}
	if !c.Scanner.Scan() {
		return "", c.Scanner.Err()
	}
	return c.Scanner.Text(), nil
}

func writeLine(w io.Writer, s string) error {
	_, err := io.WriteString(w, s+"\n")
	return err
}

func makeClient(addr string) (*client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("Error establishing a connection: %s", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	err = conn.SetDeadline(deadline)
	if err != nil {
		return nil, fmt.Errorf("Failed to set deadline: %s\n", err)
	}

	return &client{Conn: conn}, nil
}

// Assert that the message is a message sent server (starts with `*`) and contains the
// expected strings
func assertServerMessage(t *testing.T, m string, expected ...string) {
	if !strings.HasPrefix(m, "*") {
		t.Fatalf("%s does not start with character `*`", m)
	}

	for _, s := range expected {
		if !strings.Contains(m, s) {
			t.Fatalf("%s does not contain expected substring %s", m, s)
		}
	}
}
