package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	targetURL, err := url.Parse(r.RequestURI)
	if err != nil {
		http.Error(w, "Invalid target URL", http.StatusBadRequest)
		return
	}

	targetRequest := r.Clone(r.Context())
	targetRequest.RequestURI = ""
	targetRequest.URL = targetURL

	targetResponse, err := http.DefaultTransport.RoundTrip(targetRequest)
	if err != nil {
		http.Error(w, "Error forwarding request", http.StatusInternalServerError)
		return
	}
	defer targetResponse.Body.Close()

	for k, vv := range targetResponse.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(targetResponse.StatusCode)
	io.Copy(w, targetResponse.Body)
}

func main() {
	http.HandleFunc("/", handleRequest)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("ListenAndServe error:", err)
	}
}
