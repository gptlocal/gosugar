package main

import (
	"io"
	"net/http"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}

	// 创建一个新的请求，将原始请求的方法、URL 和头信息复制过来
	req, err := http.NewRequest(r.Method, r.URL.String(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header = r.Header

	// 向目标服务器发送请求
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 将响应头复制回客户端
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	// 使用 io.Pipe 创建一个管道
	pr, pw := io.Pipe()
	go func() {
		// 将目标服务器的响应写入管道
		defer pw.Close()
		io.Copy(pw, resp.Body)
	}()

	// 从管道中读取数据并写回客户端
	io.Copy(w, pr)
}

func main() {
	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(":8080", nil)
}
