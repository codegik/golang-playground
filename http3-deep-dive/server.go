package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

var (
	connMutex sync.Mutex
	connCount int32
	reqCount  int32
	startTime time.Time
)

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
		NextProtos:   []string{"h3"},
	}
}

func RunInstrumentedServer() {
	fmt.Println("\n═══════════════════════════════════════════════════")
	fmt.Println("     HTTP/3 Server with Instrumentation")
	fmt.Println("═══════════════════════════════════════════════════")

	startTime = time.Now()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqNum := atomic.AddInt32(&reqCount, 1)
		elapsed := time.Since(startTime)

		fmt.Printf("\n[HTTP REQUEST #%d] +%dms\n", reqNum, elapsed.Milliseconds())
		fmt.Printf("  %s %s\n", r.Method, r.URL.Path)
		fmt.Printf("  Proto: %s\n", r.Proto)
		fmt.Printf("  Remote: %s\n", r.RemoteAddr)
		fmt.Printf("  Headers:\n")

		for key, values := range r.Header {
			for _, value := range values {
				fmt.Printf("    %s: %s\n", key, value)
			}
		}

		if r.Body != nil {
			body, _ := io.ReadAll(r.Body)
			if len(body) > 0 {
				fmt.Printf("  Body: %d bytes\n", len(body))
			}
		}

		response := fmt.Sprintf(`{"path": "%s", "method": "%s", "timestamp": "%s", "request_number": %d}`,
			r.URL.Path, r.Method, time.Now().Format(time.RFC3339), reqNum)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Request-Number", fmt.Sprintf("%d", reqNum))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))

		fmt.Printf("[HTTP RESPONSE #%d] 200 OK (%d bytes)\n", reqNum, len(response))
	})

	tlsConfig := generateTLSConfig()

	quicConfig := &quic.Config{
		MaxIdleTimeout:  30 * time.Second,
		Allow0RTT:       true,
		EnableDatagrams: true,
	}

	server := http3.Server{
		Handler:    handler,
		Addr:       ":4433",
		TLSConfig:  tlsConfig,
		QUICConfig: quicConfig,
	}

	fmt.Println("\nServer configuration:")
	fmt.Println("  Address: :4433")
	fmt.Println("  Protocol: HTTP/3 (h3)")
	fmt.Println("  0-RTT: Enabled")
	fmt.Println("  Datagrams: Enabled")
	fmt.Println("  Max Idle: 30s")
	fmt.Println("\nKey Features:")
	fmt.Println("  • QUIC transport over UDP")
	fmt.Println("  • TLS 1.3 integrated encryption")
	fmt.Println("  • Stream multiplexing")
	fmt.Println("  • Connection migration support")
	fmt.Println("\nWaiting for connections...\n")
	fmt.Println("Try: go run . -mode client")
	fmt.Println("")

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
