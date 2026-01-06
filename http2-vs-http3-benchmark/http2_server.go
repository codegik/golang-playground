package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

type HTTP2Server struct {
	port int
}

func NewHTTP2Server(port int) *HTTP2Server {
	return &HTTP2Server{port: port}
}

func (s *HTTP2Server) Start() error {
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
		response := fmt.Sprintf(`{"timestamp":"%s","server":"http2","status":"ok"}`,
			time.Now().Format(time.RFC3339))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	})

	tlsConfig := generateTLSConfig()

	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", s.port),
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	fmt.Printf("[HTTP/2 Server] Starting on port %d\n", s.port)
	return server.ListenAndServeTLS("", "")
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
		DNSNames:     []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"h2", "http/1.1"},
	}
}
