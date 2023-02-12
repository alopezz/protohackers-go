package budgetchat

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"protohackers/protos"
	"regexp"
	"strings"
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
	users map[string]net.Conn
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
	name, err := readLine(scanner)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	if !isValidName(name) {
		writeLine(conn, "* Illegal name provided, disconnecting!")
		return
	}

	for n := range c.users {
		if n == name {
			writeLine(conn, "* Name already in use, disconnecting!")
			return
		}
	}

	usernames := make([]string, 0, len(c.users))
	for n := range c.users {
		usernames = append(usernames, n)
	}

	writeLine(conn, fmt.Sprintf("* The room contains: %s", strings.Join(usernames, ", ")))

	// Announce to others in the room
	for _, c := range c.users {
		writeLine(c, "* %s has entered the room", name)
	}

	// Ensure map creation before assignment
	if c.users == nil {
		c.users = make(map[string]net.Conn)
	}

	c.users[name] = conn

	// Main chat loop
	for scanner.Scan() {
		msg := strings.TrimRight(scanner.Text(), " \n\t\r\n")
		fmt.Printf("User %s sent: %s\n", name, msg)

		for n, c := range c.users {
			if n != name {
				writeLine(c, "[%s] %s", name, msg)
			}
		}
	}

	fmt.Printf("Failed to read message from client, disconnecting: %s\n", scanner.Err())
	for _, c := range c.users {
		writeLine(c, "* %s has left the room", name)
	}
}

func writeLine(w io.Writer, s string, args ...any) error {
	_, err := io.WriteString(w, fmt.Sprintf(s, args...)+"\n")
	return err
}

func readLine(s *bufio.Scanner) (string, error) {
	if !s.Scan() {
		return "", fmt.Errorf("Failed to read message from client, disconnection: %s\n", s.Err())
	}

	return strings.TrimRight(s.Text(), " \n\t\r\n"), nil
}

func isValidName(name string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]+$`, name)
	return matched
}
