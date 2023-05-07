package main

import (
	"io"
	"net/http"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// 向目标服务器发送请求
	resp, err := http.DefaultTransport.RoundTrip(r)
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
	w.WriteHeader(resp.StatusCode)

	// 以下代码等价于 io.Copy(w, resp.Body), 使用io.Pipe() 使得代码反而会变得更复杂

	// 使用 io.Pipe 创建一个管道
	src, sink := io.Pipe()
	go func() {
		// 将目标服务器的响应写入管道
		defer sink.Close()
		io.Copy(sink, resp.Body)
	}()

	// 从管道中读取数据并写回客户端
	io.Copy(w, src)
}

func main() {
	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(":8080", nil)
}
