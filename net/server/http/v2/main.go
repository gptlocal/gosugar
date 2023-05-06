package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "OK")
	})

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
