package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

func main() {
	// 监听指定端口
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Error starting tcp server: %v", err)
	}
	defer listener.Close()

	log.Println("TCP server started on port 8080")

	for {
		// 接受客户端连接
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		// 处理客户端请求
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		fmt.Printf("Error reading request: %v\n", err)
		return
	}

	response := http.Response{
		StatusCode:    200,
		ProtoMajor:    request.ProtoMajor,
		ProtoMinor:    request.ProtoMinor,
		Header:        http.Header{"Content-Type": []string{"text/plain"}},
		Body:          io.NopCloser(strings.NewReader("OK")),
		ContentLength: 2,
	}

	err = response.Write(conn)
	if err != nil {
		fmt.Printf("Error writing response: %v\n", err)
	}
}
