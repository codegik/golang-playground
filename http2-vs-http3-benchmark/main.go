package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	mode := flag.String("mode", "benchmark", "Mode: benchmark, http2-server, http3-server")
	http2Port := flag.Int("http2-port", 8443, "HTTP/2 server port")
	http3Port := flag.Int("http3-port", 9443, "HTTP/3 server port")
	flag.Parse()

	switch *mode {
	case "http2-server":
		runHTTP2Server(*http2Port)
	case "http3-server":
		runHTTP3Server(*http3Port)
	case "benchmark":
		runBenchmark(*http2Port, *http3Port)
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		flag.Usage()
		os.Exit(1)
	}
}

func runHTTP2Server(port int) {
	server := NewHTTP2Server(port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n[HTTP/2 Server] Shutting down...")
		os.Exit(0)
	}()

	if err := server.Start(); err != nil {
		fmt.Printf("HTTP/2 server error: %v\n", err)
		os.Exit(1)
	}
}

func runHTTP3Server(port int) {
	server := NewHTTP3Server(port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n[HTTP/3 Server] Shutting down...")
		os.Exit(0)
	}()

	if err := server.Start(); err != nil {
		fmt.Printf("HTTP/3 server error: %v\n", err)
		os.Exit(1)
	}
}

func runBenchmark(http2Port, http3Port int) {
	fmt.Println("Starting HTTP/2 and HTTP/3 servers...")

	http2Server := NewHTTP2Server(http2Port)
	go func() {
		if err := http2Server.Start(); err != nil {
			fmt.Printf("HTTP/2 server error: %v\n", err)
		}
	}()

	http3Server := NewHTTP3Server(http3Port)
	go func() {
		if err := http3Server.Start(); err != nil {
			fmt.Printf("HTTP/3 server error: %v\n", err)
		}
	}()

	fmt.Println("Waiting for servers to start...")
	time.Sleep(3 * time.Second)

	benchmark := NewBenchmark(http2Port, http3Port)
	results := benchmark.RunAll()

	report := NewResultsReport(results)
	report.Print()

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("  Benchmark Complete!")
	fmt.Println(strings.Repeat("=", 80) + "\n")
}
