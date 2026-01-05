package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

func RunInstrumentedClient(serverAddr string, paths []string) {
	fmt.Println("\n═══════════════════════════════════════════════════")
	fmt.Println("     HTTP/3 Client with Protocol Tracing")
	fmt.Println("═══════════════════════════════════════════════════")

	client := &http.Client{
		Transport: &http3.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				NextProtos:         []string{"h3"},
			},
			QUICConfig: &quic.Config{
				MaxIdleTimeout:  30 * time.Second,
				EnableDatagrams: true,
			},
		},
		Timeout: 10 * time.Second,
	}

	fmt.Printf("\nTarget: %s\n", serverAddr)
	fmt.Printf("Paths to request: %d\n", len(paths))
	fmt.Println("\nHTTP/3 Features Being Used:")
	fmt.Println("  • QUIC transport (UDP-based)")
	fmt.Println("  • TLS 1.3 encryption")
	fmt.Println("  • Stream multiplexing")
	fmt.Println("  • No head-of-line blocking")
	fmt.Println("\n─────────────────────────────────────────────────")

	overallStart := time.Now()

	for i, path := range paths {
		fmt.Printf("\n[REQUEST %d/%d] GET %s\n", i+1, len(paths), path)

		url := fmt.Sprintf("https://%s%s", serverAddr, path)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			continue
		}

		requestStart := time.Now()
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		requestElapsed := time.Since(requestStart)

		if err != nil {
			fmt.Printf("Error reading body: %v\n", err)
			continue
		}

		fmt.Printf("[RESPONSE %d/%d] %s\n", i+1, len(paths), resp.Status)
		fmt.Printf("  Time: %dms\n", requestElapsed.Milliseconds())
		fmt.Printf("  Protocol: %s\n", resp.Proto)
		fmt.Printf("  Content-Length: %d bytes\n", len(body))

		if reqNum := resp.Header.Get("X-Request-Number"); reqNum != "" {
			fmt.Printf("  Server Request #: %s\n", reqNum)
		}

		fmt.Printf("  Body: %s\n", string(body))

		if i < len(paths)-1 {
			fmt.Println("\n  (Stream multiplexing allows concurrent requests)")
			time.Sleep(200 * time.Millisecond)
		}
	}

	totalElapsed := time.Since(overallStart)

	fmt.Println("\n═══════════════════════════════════════════════════")
	fmt.Printf("\nTotal Time: %dms for %d requests\n", totalElapsed.Milliseconds(), len(paths))
	fmt.Printf("Average: %dms per request\n", totalElapsed.Milliseconds()/int64(len(paths)))

	fmt.Println("\n🔑 Key Observations:")
	fmt.Println("  • All requests used the same QUIC connection")
	fmt.Println("  • Each request used a different stream")
	fmt.Println("  • No TCP handshake overhead")
	fmt.Println("  • Connection survives network changes")
	fmt.Println("═══════════════════════════════════════════════════")
}
