package budgetchat

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"protohackers/protos"
	"regexp"
	"strings"
	"sync"
)

func Serve(address string) (protos.Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("Failed to listen: %s\n", err)
	}
	c := &chatRoom{}
	go protos.Serve(listener, c.handleConnection)

	return &server{Listener: listener}, nil
}

type server struct {
	net.Listener
}

type chatRoom struct {
	users map[string](chan string)
	mu    sync.Mutex
}

func (c *chatRoom) handleConnection(conn net.Conn) {
	defer conn.Close()

	err := writeLine(conn, "Welcome to budgetchat! What shall I call you?")
	if err != nil {
		fmt.Printf("Something went wrong when sending a message to the client: %s\n", err)
		return
	}

	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.ScanLines)

	// Handle user registration
	if !scanner.Scan() {
		fmt.Printf("Failed to read message from client, disconnecting: %s\n", scanner.Err())
		return
	}
	name := trimMessage(scanner.Text())

	recvChannel, err := c.join(name)
	if err != nil {
		writeLine(conn, "* %s", err)
		return
	}
	defer c.leave(name)

	// Send received messages to the client in a separate goroutine
	go func() {
		for msg := range recvChannel {
			writeLine(conn, msg)
		}
	}()

	// Read loop
	for scanner.Scan() {
		msg := trimMessage(scanner.Text())
		c.broadcast(fmt.Sprintf("[%s] %s", name, msg), name)
	}

	// Leave
	fmt.Printf("Failed to read message from client, disconnecting: %s\n", scanner.Err())
	c.broadcast(fmt.Sprintf("* %s has left the room", name), name)
	return
}

func (c *chatRoom) broadcast(msg string, exceptions ...string) {
	// Wrap the manipulation of the users in a mutex.
	c.mu.Lock()
	defer c.mu.Unlock()

	for n, ch := range c.users {
		skip := false
		for _, e := range exceptions {
			if e == n {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		ch <- msg
	}
}

func (c *chatRoom) join(name string) (chan string, error) {
	if !isValidName(name) {
		return nil, fmt.Errorf("Illegal name provided, disconnecting!")
	}

	// Wrap the manipulation of the users in a mutex.
	c.mu.Lock()
	defer c.mu.Unlock()

	for n := range c.users {
		if n == name {
			return nil, fmt.Errorf("Name already in use, disconnecting!")
		}
	}

	usernames := make([]string, 0, len(c.users))
	for n := range c.users {
		usernames = append(usernames, n)
	}

	recvChannel := make(chan string, 10)

	recvChannel <- fmt.Sprintf("* The room contains: %s", strings.Join(usernames, ", "))

	// Announce to others in the room
	c.mu.Unlock()
	c.broadcast(fmt.Sprintf("* %s has entered the room", name))
	c.mu.Lock()

	// Ensure map creation before assignment
	if c.users == nil {
		c.users = make(map[string](chan string))
	}

	c.users[name] = recvChannel
	return recvChannel, nil
}

func (c *chatRoom) leave(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.users, name)
}

func writeLine(w io.Writer, s string, args ...any) error {
	_, err := io.WriteString(w, fmt.Sprintf(s, args...)+"\n")
	return err
}

func isValidName(name string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, name)
	return matched
}

func trimMessage(msg string) string {
	return strings.TrimRight(msg, " \n\t\r\n")
}
