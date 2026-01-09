package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

type HTTP3Server struct {
	port int
}

func NewHTTP3Server(port int) *HTTP3Server {
	return &HTTP3Server{port: port}
}

func (s *HTTP3Server) Start() error {
	mux := setupHandlers("http3")

	tlsConfig := generateHTTP3TLSConfig()

	quicConfig := &quic.Config{
		MaxIdleTimeout:  30 * time.Second,
		Allow0RTT:       true,
		EnableDatagrams: true,
	}

	server := &http3.Server{
		Handler:    mux,
		Addr:       fmt.Sprintf(":%d", s.port),
		TLSConfig:  tlsConfig,
		QUICConfig: quicConfig,
	}

	fmt.Printf("[HTTP/3 Server] Starting on port %d\n", s.port)
	return server.ListenAndServe()
}

func generateHTTP3TLSConfig() *tls.Config {
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
		NextProtos:   []string{"h3"},
	}
}
