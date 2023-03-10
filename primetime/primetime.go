package primetime

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
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

	s := bufio.NewScanner(conn)
	s.Split(bufio.ScanLines)

	for s.Scan() {
		input := new(inputNumber)
		err := json.Unmarshal(s.Bytes(), input)

		if err != nil {
			fmt.Printf("Error when parsing request: %s\n", err)
			io.WriteString(conn, "{}\n")
			break
		}
		b, err := json.Marshal(response{
			Method: "isPrime",
			Prime:  isPrime(*input),
		})
		if err != nil {
			fmt.Printf("Error serializing response: %s\n", err)
			io.WriteString(conn, "{}\n")
			continue
		}
		conn.Write(b)
		io.WriteString(conn, "\n")

	}
}

type response struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

type inputNumber int

func (n *inputNumber) UnmarshalJSON(data []byte) error {
	request := make(map[string]interface{})
	err := json.Unmarshal(data, &request)
	if err != nil {
		return err
	}

	method, ok := request["method"]
	if !ok {
		return fmt.Errorf("method field is not present")
	}

	if method, ok := method.(string); !ok || method != "isPrime" {
		return fmt.Errorf("invalid value for method field")
	}

	number, ok := request["number"]
	if !ok {
		return fmt.Errorf("number field is not present")
	}
	number, ok = number.(float64)
	if !ok {
		return fmt.Errorf("number field is not a number")
	}

	*n = inputNumber(number.(float64))

	return nil
}

func isPrime(n inputNumber) bool {
	if n <= 1 {
		return false
	}

	for divisor := 2; divisor <= int(math.Sqrt(float64(n))); divisor++ {
		if int(n)%divisor == 0 {
			return false
		}
	}
	return true
}
