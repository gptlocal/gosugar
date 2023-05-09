package main

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

// 用于将数据从一个连接传输到另一个连接
func transfer(dst, src net.Conn) {
	defer dst.Close()
	defer src.Close()
	io.Copy(dst, src)
}

func handleConnection(conn net.Conn, tlsConfig *tls.Config) {
	// 使用TLS加密连接
	tlsConn := tls.Server(conn, tlsConfig)
	defer tlsConn.Close()

	// 读取并解析Trojan协议请求
	targetAddress, err := parseTrojanRequest(tlsConn)
	if err != nil {
		log.Printf("Error parsing request: %v", err)
		return
	}

	// 连接到目标服务器
	targetConn, err := net.Dial("tcp", targetAddress)
	if err != nil {
		log.Printf("Error connecting to target: %v", err)
		return
	}
	defer targetConn.Close()

	// 传输数据
	go transfer(targetConn, tlsConn)
	transfer(tlsConn, targetConn)
}

func parseTrojanRequest(tlsConn *tls.Conn) (string, error) {
	// 读取第一个字节，该字节表示地址类型
	addrType := make([]byte, 1)
	_, err := io.ReadFull(tlsConn, addrType)
	if err != nil {
		return "", err
	}

	var buf []byte

	// 根据地址类型读取地址和端口
	switch addrType[0] {
	case 1: // IPv4
		buf = make([]byte, 4+2) // 4 bytes for IPv4 address, 2 bytes for port
	case 3: // 域名
		// 读取域名长度
		domainLen := make([]byte, 1)
		_, err = io.ReadFull(tlsConn, domainLen)
		if err != nil {
			return "", err
		}
		// 读取域名和端口
		buf = make([]byte, int(domainLen[0])+2)
	case 4: // IPv6
		buf = make([]byte, 16+2) // 16 bytes for IPv6 address, 2 bytes for port
	default:
		return "", fmt.Errorf("unsupported address type: %d", addrType[0])
	}

	// 读取地址和端口
	_, err = io.ReadFull(tlsConn, buf)
	if err != nil {
		return "", err
	}

	// 解析目标地址和端口
	var targetAddress string
	switch addrType[0] {
	case 1: // IPv4
		targetAddress = fmt.Sprintf("%d.%d.%d.%d:%d", buf[0], buf[1], buf[2], buf[3], binary.BigEndian.Uint16(buf[4:6]))
	case 3: // 域名
		targetAddress = fmt.Sprintf("%s:%d", string(buf[:len(buf)-2]), binary.BigEndian.Uint16(buf[len(buf)-2:]))
	case 4: // IPv6
		targetAddress = fmt.Sprintf("[%x:%x:%x:%x:%x:%x:%x:%x]:%d",
			binary.BigEndian.Uint16(buf[0:2]),
			binary.BigEndian.Uint16(buf[2:4]),
			binary.BigEndian.Uint16(buf[4:6]),
			binary.BigEndian.Uint16(buf[6:8]),
			binary.BigEndian.Uint16(buf[8:10]),
			binary.BigEndian.Uint16(buf[10:12]),
			binary.BigEndian.Uint16(buf[12:14]),
			binary.BigEndian.Uint16(buf[14:16]),
			binary.BigEndian.Uint16(buf[16:18]))
	}

	return targetAddress, nil
}

func main() {
	// 生成TLS配置
	cert, err := tls.LoadX509KeyPair("/opt/etc/ssl/localgpt.net/localgpt.net.cer", "/opt/etc/ssl/localgpt.net/localgpt.net.key")
	if err != nil {
		log.Fatalf("Error loading TLS key pair: %s", err)
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	// 创建TCP监听器
	listener, err := net.Listen("tcp", ":443")
	if err != nil {
		log.Fatalf("Error listening on port 443: %s", err)
	}

	// 处理传入的连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %s", err)
			continue
		}
		go handleConnection(conn, tlsConfig)
	}
}
