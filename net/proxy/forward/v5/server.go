package v5

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type Server struct {
	tcpListener net.Listener
	cmd         *exec.Cmd
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewServer creates a transport layer server
func NewServer(ctx context.Context, listenAddress string) (*Server, error) {
	var cmd *exec.Cmd
	tcpListener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	server := &Server{
		tcpListener: tcpListener,
		cmd:         cmd,
		ctx:         ctx,
		cancel:      cancel,
	}
	go server.acceptLoop()
	return server, nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.tcpListener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				log.Printf("transport accept error: %v", err)
				time.Sleep(time.Millisecond * 100)
				continue
			}
		}

		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	reqBufReader := bufio.NewReader(io.NopCloser(conn))
	req, err := http.ReadRequest(reqBufReader)
	if err != nil {
		log.Printf("not a valid http request: %v", err)
		return
	}

	if strings.ToUpper(req.Method) == "CONNECT" { // CONNECT
		addr, err := parseAddr(req)
		if err != nil {
			log.Printf("invalid http dest address: %v", err)
			return
		}
		resp := fmt.Sprintf("HTTP/%d.%d 200 Connection established\r\n\r\n", req.ProtoMajor, req.ProtoMinor)
		_, err = conn.Write([]byte(resp))
		if err != nil {
			log.Printf("http failed to respond connect request: %v", err)
			return
		}

		//destConn, err := net.DialTimeout("tcp", addr, 10*time.Second)
		//if err != nil {
		//	log.Printf("can't dial %s: %v", addr, err)
		//	return
		//}

	} else { // GET, POST, PUT...
		for {
			reqReader, reqWriter := io.Pipe()
			rspReader, rspWriter := io.Pipe()

			addr, err := parseAddr(req)
			if err != nil {
				log.Printf("invalid http dest address: %v", err)
				return
			}

			err = req.Write(reqWriter) // write request to the remote
			if err != nil {
				log.Printf("http failed to write http request: %v", err)
				return
			}

			rspBufReader := bufio.NewReader(io.NopCloser(rspReader)) // read response from the remote
			rsp, err := http.ReadResponse(rspBufReader, req)
			if err != nil {
				log.Printf("http failed to read http response: %v", err)
				return
			}
			defer rsp.Body.Close()

			err = rsp.Write(conn) // send the response back to the local
			if err != nil {
				log.Printf("http failed to write the response back: %v", err)
				return
			}

			req, err = http.ReadRequest(reqBufReader) // read the next http request from local
			if err != nil {
				log.Printf("http failed to the read request from local: %v", err)
				return
			}
		}
	}
}

func (s *Server) Close() error {
	s.cancel()
	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
	}
	return s.tcpListener.Close()
}

func parseAddr(r *http.Request) (string, error) {
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			host = r.Host
			port = "80"
			if r.URL.Scheme == "https" {
				port = "443"
			}
		} else {
			return "", err
		}
	}
	return net.JoinHostPort(host, port), nil
}
