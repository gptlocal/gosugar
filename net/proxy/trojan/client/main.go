package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func handleClient(clientConn net.Conn) {
	defer clientConn.Close()

	// 从客户端读取目标地址和端口（这里假设使用方传入的是以 "example.com:80" 格式的字符串）
	reader := bufio.NewReader(clientConn)
	target, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading target:", err)
		return
	}
	target = strings.TrimSuffix(target, "\n")

	// 配置 TLS 连接
	tlsConfig := &tls.Config{
		ServerName: "gfw.localgpt.net",
	}

	// 与 Trojan 服务器建立 TLS 连接
	remoteConn, err := tls.Dial("tcp", "gfw.localgpt.net:443", tlsConfig)
	if err != nil {
		fmt.Println("Error dialing remote:", err)
		return
	}
	defer remoteConn.Close()

	// 发送 Trojan 请求（此处省略实际 Trojan 请求，只发送目标地址和端口）
	_, err = remoteConn.Write([]byte(target + "\r\n"))
	if err != nil {
		fmt.Println("Error dialing remote:", err)
		return
	}

	// 开始在客户端和远程服务器之间转发数据
	go io.Copy(remoteConn, clientConn)
	io.Copy(clientConn, remoteConn)
}

func main() {
	localAddr := "localhost:1080"
	remoteAddr := "gfw.localgpt.net:443"

	listener, err := net.Listen("tcp", localAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", localAddr, err)
	}

	log.Printf("Listening on %s, forwarding to %s", localAddr, remoteAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go handleClient(conn)
	}
}
