package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// setupHandlers creates and returns a configured http.ServeMux with all benchmark endpoints
func setupHandlers(serverName string) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	mux.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		sizeStr := r.URL.Query().Get("size")
		size := 1024
		if sizeStr != "" {
			if parsed, err := strconv.Atoi(sizeStr); err == nil {
				size = parsed
			}
		}

		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.Itoa(size))
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	mux.HandleFunc("/delay", func(w http.ResponseWriter, r *http.Request) {
		delayStr := r.URL.Query().Get("ms")
		delayMs := 100
		if delayStr != "" {
			if parsed, err := strconv.Atoi(delayStr); err == nil {
				delayMs = parsed
			}
		}

		time.Sleep(time.Duration(delayMs) * time.Millisecond)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("delayed %dms", delayMs)))
	})

	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		response := fmt.Sprintf(`{"timestamp":"%s","server":"%s","status":"ok"}`,
			time.Now().Format(time.RFC3339), serverName)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	})

	return mux
}
