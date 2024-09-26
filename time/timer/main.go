package main

import (
	"fmt"
	"io"
	"sync"
	"time"
)

func main() {
	r, w := io.Pipe()

	var once sync.Once
	defer once.Do(func() {
		w.Close()
		r.Close()
	})

	timer := time.AfterFunc(time.Second*10, func() {
		once.Do(func() {
			fmt.Println("Timeout")
			w.Close()
			r.Close()
		})
	})
	defer timer.Stop()

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(n)
}
