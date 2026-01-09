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
	"time"
)

type HTTP2Server struct {
	port int
}

func NewHTTP2Server(port int) *HTTP2Server {
	return &HTTP2Server{port: port}
}

func (s *HTTP2Server) Start() error {
	mux := setupHandlers("http2")

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
