package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"strings"
)

var directory string

func handleConnection(conn net.Conn, directory string) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Чтение строки запроса
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}
	requestLine = strings.TrimSpace(requestLine)

	// Разбор строки запроса на метод, путь и версию протокола
	parts := strings.Split(requestLine, " ")
	if len(parts) < 3 {
		fmt.Println("Invalid request line:", requestLine)
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}
	method := parts[0]
	path := parts[1]

	// Чтение заголовков
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

	// Обработка запроса
	if method == "GET" {
		parts := strings.Split(path, "/")
		fmt.Println(parts)
		switch parts[1] {
		case "files":
			filename := parts[2]
			fullPath := filepath.Join(directory, filename)
			contents, err := ioutil.ReadFile(fullPath)
			if err != nil {
				fmt.Println("Error opening file:", err)
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				return
			}
			response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n", len(contents))
			conn.Write([]byte(response))
			conn.Write(contents)
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

	} else {
		conn.Write([]byte("HTTP/1.1 405 Method Not Allowed\r\n\r\n"))
	}
}

func main() {
	// Чтение флага командной строки
	flag.StringVar(&directory, "directory", ".", "Directory to serve files from")
	flag.Parse()

	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", ":4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		return
	}
	defer l.Close()

	fmt.Println("Server listening on port 4221")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn, directory)
	}
}
