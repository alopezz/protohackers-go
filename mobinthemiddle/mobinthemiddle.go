package mobinthemiddle

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"protohackers/protos"
	"regexp"
	"strings"
)

const TONY_ADDRESS = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"

var boguscoinRegexp = regexp.MustCompile(`^7[a-zA-Z0-9]{25,34}$`)

func Serve(address string, upstreamAddress string) (protos.Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("Failed to listen: %s\n", err)
	}

	handler := func(conn net.Conn) {
		handler(conn, upstreamAddress)
	}

	go protos.Serve(listener, handler)

	return &server{Listener: listener}, nil
}

type server struct {
	net.Listener
}

func handler(conn net.Conn, addr string) {
	upstream, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("Error establishing a connection to upstream server: %s", err)
		return
	}

	go Tamper(conn, upstream)
	go Tamper(upstream, conn)
}

func Tamper(src io.ReadCloser, dst io.WriteCloser) {
	s := bufio.NewScanner(src)
	s.Split(scanCompleteLines)

	for s.Scan() {
		err := writeLine(dst, rewriteBogus(s.Text()))
		if err != nil {
			src.Close()
			return
		}
	}

	dst.Close()
}

func writeLine(w io.Writer, s string) error {
	_, err := io.WriteString(w, s+"\n")
	return err
}

func rewriteBogus(in string) string {
	words := strings.Split(in, " ")
	for i, word := range words {
		if boguscoinRegexp.MatchString(word) {
			words[i] = TONY_ADDRESS
		}
	}
	return strings.Join(words, " ")
}

// A modification over the standard `ScanLines` that won't yield incomplete lines
func scanCompleteLines(data []byte, atEOF bool) (int, []byte, error) {
	// Stop and don't return anything if we hit EOF
	if atEOF {
		return 0, nil, nil
	}

	// For everything else, delegate to `bufio.ScanLines`
	return bufio.ScanLines(data, atEOF)
}
