package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	targetURL, err := url.Parse(r.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
	}

	targetConn, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer targetConn.Close()

	err = r.Write(targetConn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	resp, err := http.ReadResponse(bufio.NewReader(targetConn), r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)
}

func main() {
	http.HandleFunc("/", handleRequest)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("ListenAndServe error:", err)
	}
}
