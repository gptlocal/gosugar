package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
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
		go handleHttpConn(conn)
	}
}

func handleHttpConn(conn net.Conn) {
	defer conn.Close()
	r, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		log.Printf("Can't read http request: %v", err)
		sendBadRequest(conn)
		return
	}

	targetURL, err := url.Parse(r.URL.String())
	if err != nil {
		log.Printf("Can't parse request url: %v", err)
		sendBadRequest(conn)
		return
	}

	host, port, err := net.SplitHostPort(targetURL.Host)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			host = targetURL.Host
			port = "80"
			if r.URL.Scheme == "https" {
				port = "443"
			}
		} else {
			log.Printf("Can't get port: %v", err)
			sendBadRequest(conn)
			return
		}
	}

	targetConn, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		log.Printf("Can't dial connection: %v", err)
		sendServiceUnavailable(conn)
		return
	}
	defer targetConn.Close()

	err = r.Write(targetConn)
	if err != nil {
		log.Printf("Can't send request: %v", err)
		sendServiceUnavailable(conn)
		return
	}

	resp, err := http.ReadResponse(bufio.NewReader(targetConn), r)
	if err != nil {
		log.Printf("Can't read response: %v", err)
		sendServiceUnavailable(conn)
		return
	}
	defer resp.Body.Close()

	// Write the status line
	conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", resp.StatusCode, resp.Status)))

	// Write the headers
	for k, vv := range resp.Header {
		for _, v := range vv {
			conn.Write([]byte(fmt.Sprintf("%s: %s\r\n", k, v)))
		}
	}

	// Write the empty line
	conn.Write([]byte("\r\n"))

	io.Copy(conn, resp.Body)
}

func sendBadRequest(conn net.Conn) {
	badRequest := "HTTP/1.1 400 Bad Request\r\nContent-Type: text/plain\r\nContent-Length: 0\r\n\r\n"
	conn.Write([]byte(badRequest))
}

func sendServiceUnavailable(conn net.Conn) {
	badRequest := "HTTP/1.1 503 Service Unavailable\r\nContent-Type: text/plain\r\nContent-Length: 0\r\n\r\n"
	conn.Write([]byte(badRequest))
}
