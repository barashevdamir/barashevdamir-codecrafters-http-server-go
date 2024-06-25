package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var directory string
var data []byte

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
	var body []byte
	for {
		line, err := reader.ReadString('\n')
		fmt.Println(line)
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
	// Чтение тела запроса, если присутствует Content-Length
	if contentLength, ok := headers["Content-Length"]; ok {
		length, err := strconv.Atoi(contentLength)
		if err != nil {
			fmt.Println("Invalid Content-Length:", err)
			conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
			return
		}

		body := make([]byte, length)
		_, err = io.ReadFull(reader, body)
		if err != nil {
			fmt.Println("Error reading body:", err)
			return
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
			contents, err := os.ReadFile(fullPath)
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

	} else if method == "POST" {
		parts := strings.Split(path, "/")
		fmt.Println(parts)
		if parts[1] == "files" {
			filename := parts[2]
			fullPath := filepath.Join(directory, filename)
			err := os.WriteFile(fullPath, body, 0644)
			if err != nil {
				fmt.Println("Error opening file:", err)
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				return
			}
			conn.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
		} else {
			conn.Write([]byte("HTTP/1.1 405 Method Not Allowed\r\n\r\n"))
		}

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
