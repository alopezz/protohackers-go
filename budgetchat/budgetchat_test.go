package budgetchat_test

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"protohackers/budgetchat"
	"strings"
	"testing"
	"time"
)

func TestNewClientIsAskedForTestNewClientIsAskedForItsName(t *testing.T) {
	server, err := budgetchat.Serve("localhost:")
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

	s := bufio.NewScanner(conn)
	s.Split(bufio.ScanLines)

	if !s.Scan() || len(s.Bytes()) == 0 {
		t.Fatalf("Did not receive prompt from server")
	}
}

func TestProvidingIllegalNameDisconnectsClient(t *testing.T) {
	server, err := budgetchat.Serve("localhost:")
	if err != nil {
		t.Fatalf("Failed to start server: %s\n", err)
	}
	defer server.Close()

	// Test cases represent different illegal user names
	testCases := []string{
		"",
		"under_scored",
		"name with spaces",
		"中文名字",
		"a.b,c;d?",
	}

	for _, username := range testCases {
		t.Run(fmt.Sprintf("illegal name `%s`", username), func(t *testing.T) {
			client, err := makeClient(server.Addr().String())
			if err != nil {
				t.Fatalf("Error constructing client: %s", err)
			}
			defer client.Close()

			_, err = client.Recv()
			if err != nil {
				t.Fatal(err)
			}

			client.Send(username)

			// Allow for optional disconnect message
			_, err = client.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				t.Fatal(err)
			}

			// Assert disconnect
			client.SetDeadline(time.Now().Add(10 * time.Millisecond))
			_, err = client.Read(make([]byte, 1))
			if !errors.Is(err, io.EOF) {
				t.Errorf("Expected to be disconnected, but was not")
			}
		})
	}
}

func TestWhenClientProperlyJoinsItReceivesListOfPresentUsers(t *testing.T) {
	server, err := budgetchat.Serve("localhost:")
	if err != nil {
		t.Fatalf("Failed to start server: %s\n", err)
	}
	defer server.Close()

	// Create clients
	clientNames := []string{"alice", "bob", "charlie"}
	clients := make(map[string]*client, len(clientNames))

	for _, name := range clientNames {
		client, err := makeClient(server.Addr().String())
		if err != nil {
			t.Fatalf("Error constructing client: %s", err)
		}
		defer client.Close()

		_, err = client.Recv()
		if err != nil {
			t.Fatal(err)
		}

		err = client.Send(name)
		if err != nil {
			t.Fatal(err)
		}

		msg, err := client.Recv()
		if err != nil {
			t.Fatal(err)
		}

		names := make([]string, 0, len(clients))
		for name := range clients {
			names = append(names, name)
		}

		assertServerMessage(t, msg, names...)

		clients[name] = client
	}
}

func TestClientMessagesAreBroadcastedToAllClients(t *testing.T) {
	server, err := budgetchat.Serve("localhost:")
	if err != nil {
		t.Fatalf("Failed to start server: %s\n", err)
	}
	defer server.Close()

	// Create clients
	clientNames := []string{"alice", "bob", "charlie"}
	clients := make(map[string]*client, len(clientNames))

	for _, name := range clientNames {
		client, err := makeClient(server.Addr().String())
		if err != nil {
			t.Fatalf("Error constructing client: %s", err)
		}
		defer client.Close()

		_, err = client.Recv()
		if err != nil {
			t.Fatal(err)
		}

		err = client.Send(name)
		if err != nil {
			t.Fatal(err)
		}

		// Receive message with existing users
		_, err = client.Recv()
		if err != nil {
			t.Fatal(err)
		}

		clients[name] = client
	}

	clients["alice"].Send("Hello")
	expected := "[alice] Hello"

	// Both bob and charlie should be able to see the message just sent by Alice
	for _, user := range []string{"bob", "charlie"} {
		for {
			msg, err := clients[user].Recv()
			if err != nil {
				t.Fatal(err)
			}
			// Ignore server messages
			if strings.HasPrefix(msg, "*") {
				continue
			}

			if msg != expected {
				t.Fatalf("Expected to receive %s, got %s instead", expected, msg)
			}
			break
		}
	}
}

func TestAnnouncesWhenAUserJoinsAndLeavesToOtherUsers(t *testing.T) {
	server, err := budgetchat.Serve("localhost:")
	if err != nil {
		t.Fatalf("Failed to start server: %s\n", err)
	}
	defer server.Close()

	// Create clients
	clientNames := []string{"alice", "bob", "charlie"}
	clients := make(map[string]*client, len(clientNames))

	for _, name := range clientNames {
		client, err := makeClient(server.Addr().String())
		if err != nil {
			t.Fatalf("Error constructing client: %s", err)
		}
		defer client.Close()

		_, err = client.Recv()
		if err != nil {
			t.Fatal(err)
		}

		err = client.Send(name)
		if err != nil {
			t.Fatal(err)
		}

		// Receive message with existing users
		_, err = client.Recv()
		if err != nil {
			t.Fatal(err)
		}

		clients[name] = client
	}

	// alice and bob see charlie join
	msg, err := clients["alice"].Recv()
	if err != nil {
		t.Fatal(err)
	}
	assertServerMessage(t, msg, "bob")

	msg, err = clients["alice"].Recv()
	if err != nil {
		t.Fatal(err)
	}
	assertServerMessage(t, msg, "charlie")

	msg, err = clients["bob"].Recv()
	if err != nil {
		t.Fatal(err)
	}
	assertServerMessage(t, msg, "charlie")

	// alice leaves and bob and charlie get notified
	clients["alice"].Close()

	msg, err = clients["bob"].Recv()
	if err != nil {
		t.Fatal(err)
	}
	assertServerMessage(t, msg, "alice")

	msg, err = clients["charlie"].Recv()
	if err != nil {
		t.Fatal(err)
	}
	assertServerMessage(t, msg, "alice")
}

func TestExampleSession(t *testing.T) {
	server, err := budgetchat.Serve("localhost:")
	if err != nil {
		t.Fatalf("Failed to start server: %s\n", err)
	}
	defer server.Close()

	// Create clients
	clientNames := []string{"bob", "charlie", "dave", "alice"}
	clients := make(map[string]*client, len(clientNames))

	for _, name := range clientNames {
		client, err := makeClient(server.Addr().String())
		if err != nil {
			t.Fatalf("Error constructing client: %s", err)
		}
		defer client.Close()

		_, err = client.Recv()
		if err != nil {
			t.Fatal(err)
		}

		err = client.Send(name)
		if err != nil {
			t.Fatal(err)
		}

		// Receive message with existing users
		msg, err := client.Recv()
		if err != nil {
			t.Fatal(err)
		}

		// Alice learns of who is in the room
		if name == "alice" {
			assertServerMessage(t, msg, "bob", "charlie", "dave")
		}

		clients[name] = client
	}

	// Alice says something to the room
	err = clients["alice"].Send("Hello everyone")
	if err != nil {
		t.Fatal(err)
	}

	// Bob reads message and responds
	_, err = clients["bob"].Recv()
	if err != nil {
		t.Fatal(err)
	}
	err = clients["bob"].Send("hi alice")
	if err != nil {
		t.Fatal(err)
	}
	// ... and alice reads it
	msg, err := clients["alice"].Recv()
	if err != nil {
		t.Fatal(err)
	}
	expected := "[bob] hi alice"
	if msg != expected {
		t.Fatalf("Expected %s but received %s", expected, msg)
	}

	// Charlie reads message and responds
	_, err = clients["charlie"].Recv()
	if err != nil {
		t.Fatal(err)
	}
	err = clients["charlie"].Send("hello alice")
	if err != nil {
		t.Fatal(err)
	}
	// ... and alice reads it
	msg, err = clients["alice"].Recv()
	if err != nil {
		t.Fatal(err)
	}
	expected = "[charlie] hello alice"
	if msg != expected {
		t.Fatalf("Expected %s but received %s", expected, msg)
	}

	// dave leaves, alice sees notification
	clients["dave"].Close()
	msg, err = clients["alice"].Recv()
	if err != nil {
		t.Fatal(err)
	}
	assertServerMessage(t, msg, "dave")
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
