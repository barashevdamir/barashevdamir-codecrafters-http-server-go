package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {

	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", ":4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		return
	}
	for {
		conn, err := l.Accept()
		reader := bufio.NewReader(conn)
		fmt.Println(conn.LocalAddr())
		requestLine, err := reader.ReadString('\n')

		if err != nil {
			fmt.Println("Error reading request:", err)
			return
		}

		headers := make(map[string]string)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading header line:", err)
				return
			}
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}
			headerParts := strings.SplitN(line, ": ", 2)
			if len(headerParts) == 2 {
				headers[headerParts[0]] = headerParts[1]
			}
		}

		requestLine = strings.TrimSpace(requestLine)
		parts := strings.Split(requestLine, " ")
		parts = strings.Split(parts[1], "/")
		target := parts[1]
		switch target {
		case "user-agent":
			UserAgent := headers["User-Agent"]
			conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprintf("%d", len(UserAgent)) + "\r\n\r\n" + UserAgent))
		case "echo":
			echo := parts[2]
			EchoLen := len(echo)
			conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprintf("%d", EchoLen) + "\r\n\r\n" + echo))
		case "":
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		default:
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}

		fmt.Println("Received request: ", requestLine)
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
	}
}
