package main

import (
	"log"
	"net/http"
)

func main() {
	port := "8080"

	// 设置文件服务器处理器，将要服务的目录设置为当前目录（"."）
	fs := http.FileServer(http.Dir("."))

	// 将文件服务器处理器绑定到根路径 "/"
	http.Handle("/", fs)

	// 开始监听和服务
	log.Printf("Serving files on :%s...", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
