package meanstoanend

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"protohackers/protos"
)

const MESSAGE_SIZE = 9

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

	data := map[int32]int32{}

	for {
		buf := make([]byte, MESSAGE_SIZE)
		n, err := io.ReadFull(conn, buf)
		if err != nil {
			fmt.Printf("Unable to get the full response: %s", err)
			break
		}
		if n != MESSAGE_SIZE {
			fmt.Printf("Expected to get %d bytes in response, got %d", MESSAGE_SIZE, n)
			break
		}

		first, second, err := decodeRequestNumbers(buf)
		if err != nil {
			fmt.Printf("%s", err)
			break
		}

		switch buf[0] {
		case 'I':
			timestamp, price := first, second
			data[timestamp] = price

		case 'Q':
			minTime, maxTime := first, second

			result := averageData(data, minTime, maxTime)
			response := &bytes.Buffer{}
			binary.Write(response, binary.BigEndian, result)

			io.Copy(conn, response)
		}
	}
}

func decodeRequestNumbers(request []byte) (int32, int32, error) {
	a, err := decodeInt32(request[1:5])
	if err != nil {
		return 0, 0, fmt.Errorf("Failed to decode int32 value: %s", err)
	}
	b, err := decodeInt32(request[5:9])
	if err != nil {
		return 0, 0, fmt.Errorf("Failed to decode int32 value: %s", err)
	}
	return a, b, nil
}

func decodeInt32(b []byte) (int32, error) {
	var result int32
	err := binary.Read(bytes.NewBuffer(b), binary.BigEndian, &result)
	return result, err
}

func averageData(data map[int32]int32, minTime int32, maxTime int32) int32 {
	total := 0
	n := 0
	for ts, val := range data {
		if ts >= minTime && ts <= maxTime {
			total += int(val)
			n++
		}
	}

	if n == 0 {
		return 0
	}

	return int32(total / n)
}
