package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

type BenchmarkResult struct {
	Protocol         string
	Scenario         string
	TotalRequests    int
	SuccessRequests  int
	FailedRequests   int
	TotalDuration    time.Duration
	MinLatency       time.Duration
	MaxLatency       time.Duration
	AvgLatency       time.Duration
	P50Latency       time.Duration
	P95Latency       time.Duration
	P99Latency       time.Duration
	Throughput       float64
	RequestsPerSec   float64
	TotalBytesRecv   int64
}

type Benchmark struct {
	http2Port int
	http3Port int
}

func NewBenchmark(http2Port, http3Port int) *Benchmark {
	return &Benchmark{
		http2Port: http2Port,
		http3Port: http3Port,
	}
}

func (b *Benchmark) createHTTP2Client() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          10,
			MaxIdleConnsPerHost:   10,
			MaxConnsPerHost:       100,
			IdleConnTimeout:       30 * time.Second,
			DisableCompression:    false,
			ResponseHeaderTimeout: 10 * time.Second,
		},
		Timeout: 60 * time.Second,
	}
}

func (b *Benchmark) createHTTP3Client() *http.Client {
	return &http.Client{
		Transport: &http3.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				NextProtos:         []string{"h3"},
			},
			QUICConfig: &quic.Config{
				MaxIdleTimeout:                30 * time.Second,
				MaxIncomingStreams:            1000,
				MaxIncomingUniStreams:         1000,
				EnableDatagrams:               true,
				Allow0RTT:                     true,
				MaxStreamReceiveWindow:        6 * 1024 * 1024,
				MaxConnectionReceiveWindow:    15 * 1024 * 1024,
			},
			DisableCompression: false,
		},
		Timeout: 60 * time.Second,
	}
}

func (b *Benchmark) RunAll() []BenchmarkResult {
	results := []BenchmarkResult{}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  HTTP/2 vs HTTP/3 Performance Benchmark")
	fmt.Println(strings.Repeat("=", 60) + "\n")

	time.Sleep(2 * time.Second)

	scenarios := []struct {
		name string
		fn   func() []BenchmarkResult
	}{
		{"Single Request Latency", b.benchmarkSingleRequest},
		{"High Concurrency (50 concurrent)", b.benchmarkHighConcurrency},
		{"Very High Concurrency (100 concurrent)", b.benchmarkVeryHighConcurrency},
		{"Massive Small Requests (2000x1KB)", b.benchmarkMassiveSmallRequests},
		{"Simulated Network Latency (50ms)", b.benchmarkWithLatency},
		{"Large File Download", b.benchmarkLargeFile},
		{"Mixed Workload", b.benchmarkMixedWorkload},
	}

	for _, scenario := range scenarios {
		fmt.Printf("\n[Scenario] %s\n", scenario.name)
		fmt.Println(strings.Repeat("-", 60))
		scenarioResults := scenario.fn()
		results = append(results, scenarioResults...)
		time.Sleep(1 * time.Second)
	}

	return results
}

func (b *Benchmark) benchmarkSingleRequest() []BenchmarkResult {
	results := []BenchmarkResult{}

	http2Client := b.createHTTP2Client()
	http3Client := b.createHTTP3Client()

	iterations := 100

	http2Result := b.runRequests(http2Client, fmt.Sprintf("https://localhost:%d/ping", b.http2Port),
		"HTTP/2", "Single Request", iterations, 1)
	results = append(results, http2Result)

	http3Result := b.runRequests(http3Client, fmt.Sprintf("https://localhost:%d/ping", b.http3Port),
		"HTTP/3", "Single Request", iterations, 1)
	results = append(results, http3Result)

	return results
}

func (b *Benchmark) benchmarkConcurrentRequests() []BenchmarkResult {
	results := []BenchmarkResult{}

	http2Client := b.createHTTP2Client()
	http3Client := b.createHTTP3Client()

	iterations := 100
	concurrency := 10

	http2Result := b.runRequests(http2Client, fmt.Sprintf("https://localhost:%d/json", b.http2Port),
		"HTTP/2", "Concurrent Requests (10)", iterations, concurrency)
	results = append(results, http2Result)

	http3Result := b.runRequests(http3Client, fmt.Sprintf("https://localhost:%d/json", b.http3Port),
		"HTTP/3", "Concurrent Requests (10)", iterations, concurrency)
	results = append(results, http3Result)

	return results
}

func (b *Benchmark) benchmarkLargeFile() []BenchmarkResult {
	results := []BenchmarkResult{}

	http2Client := b.createHTTP2Client()
	http3Client := b.createHTTP3Client()

	size := 10 * 1024 * 1024
	iterations := 20

	http2Result := b.runRequests(http2Client, fmt.Sprintf("https://localhost:%d/data?size=%d", b.http2Port, size),
		"HTTP/2", "Large File (10MB)", iterations, 1)
	results = append(results, http2Result)

	http3Result := b.runRequests(http3Client, fmt.Sprintf("https://localhost:%d/data?size=%d", b.http3Port, size),
		"HTTP/3", "Large File (10MB)", iterations, 1)
	results = append(results, http3Result)

	return results
}

func (b *Benchmark) benchmarkManySmallRequests() []BenchmarkResult {
	results := []BenchmarkResult{}

	http2Client := b.createHTTP2Client()
	http3Client := b.createHTTP3Client()

	iterations := 500
	concurrency := 20

	http2Result := b.runRequests(http2Client, fmt.Sprintf("https://localhost:%d/data?size=1024", b.http2Port),
		"HTTP/2", "Many Small Requests (500x1KB)", iterations, concurrency)
	results = append(results, http2Result)

	http3Result := b.runRequests(http3Client, fmt.Sprintf("https://localhost:%d/data?size=1024", b.http3Port),
		"HTTP/3", "Many Small Requests (500x1KB)", iterations, concurrency)
	results = append(results, http3Result)

	return results
}

func (b *Benchmark) benchmarkMixedWorkload() []BenchmarkResult {
	results := []BenchmarkResult{}

	http2Client := b.createHTTP2Client()
	http3Client := b.createHTTP3Client()

	http2Result := b.runMixedRequests(http2Client, b.http2Port, "HTTP/2", "Mixed Workload")
	results = append(results, http2Result)

	http3Result := b.runMixedRequests(http3Client, b.http3Port, "HTTP/3", "Mixed Workload")
	results = append(results, http3Result)

	return results
}

func (b *Benchmark) benchmarkHighConcurrency() []BenchmarkResult {
	results := []BenchmarkResult{}

	http2Client := b.createHTTP2Client()
	http3Client := b.createHTTP3Client()

	iterations := 500
	concurrency := 50

	http2Result := b.runRequests(http2Client, fmt.Sprintf("https://localhost:%d/json", b.http2Port),
		"HTTP/2", "High Concurrency (50 concurrent)", iterations, concurrency)
	results = append(results, http2Result)

	http3Result := b.runRequests(http3Client, fmt.Sprintf("https://localhost:%d/json", b.http3Port),
		"HTTP/3", "High Concurrency (50 concurrent)", iterations, concurrency)
	results = append(results, http3Result)

	return results
}

func (b *Benchmark) benchmarkVeryHighConcurrency() []BenchmarkResult {
	results := []BenchmarkResult{}

	http2Client := b.createHTTP2Client()
	http3Client := b.createHTTP3Client()

	iterations := 1000
	concurrency := 100

	http2Result := b.runRequests(http2Client, fmt.Sprintf("https://localhost:%d/json", b.http2Port),
		"HTTP/2", "Very High Concurrency (100 concurrent)", iterations, concurrency)
	results = append(results, http2Result)

	http3Result := b.runRequests(http3Client, fmt.Sprintf("https://localhost:%d/json", b.http3Port),
		"HTTP/3", "Very High Concurrency (100 concurrent)", iterations, concurrency)
	results = append(results, http3Result)

	return results
}

func (b *Benchmark) benchmarkMassiveSmallRequests() []BenchmarkResult {
	results := []BenchmarkResult{}

	http2Client := b.createHTTP2Client()
	http3Client := b.createHTTP3Client()

	iterations := 2000
	concurrency := 50

	http2Result := b.runRequests(http2Client, fmt.Sprintf("https://localhost:%d/data?size=1024", b.http2Port),
		"HTTP/2", "Massive Small Requests (2000x1KB)", iterations, concurrency)
	results = append(results, http2Result)

	http3Result := b.runRequests(http3Client, fmt.Sprintf("https://localhost:%d/data?size=1024", b.http3Port),
		"HTTP/3", "Massive Small Requests (2000x1KB)", iterations, concurrency)
	results = append(results, http3Result)

	return results
}

func (b *Benchmark) benchmarkWithLatency() []BenchmarkResult {
	results := []BenchmarkResult{}

	http2Client := b.createHTTP2Client()
	http3Client := b.createHTTP3Client()

	iterations := 200
	concurrency := 20
	delayMs := 50

	http2Result := b.runRequests(http2Client, fmt.Sprintf("https://localhost:%d/delay?ms=%d", b.http2Port, delayMs),
		"HTTP/2", "Simulated Network Latency (50ms)", iterations, concurrency)
	results = append(results, http2Result)

	http3Result := b.runRequests(http3Client, fmt.Sprintf("https://localhost:%d/delay?ms=%d", b.http3Port, delayMs),
		"HTTP/3", "Simulated Network Latency (50ms)", iterations, concurrency)
	results = append(results, http3Result)

	return results
}

func (b *Benchmark) runRequests(client *http.Client, url, protocol, scenario string, iterations, concurrency int) BenchmarkResult {
	latencies := make([]time.Duration, 0, iterations)
	var totalBytes int64
	var successCount, failCount int
	var mu sync.Mutex

	start := time.Now()

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		sem <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			reqStart := time.Now()
			resp, err := client.Get(url)
			if err != nil {
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			reqDuration := time.Since(reqStart)

			mu.Lock()
			if err != nil || resp.StatusCode != 200 {
				failCount++
			} else {
				successCount++
				latencies = append(latencies, reqDuration)
				totalBytes += int64(len(body))
			}
			mu.Unlock()
		}()
	}

	wg.Wait()
	totalDuration := time.Since(start)

	return b.calculateResults(protocol, scenario, latencies, totalBytes, successCount, failCount, totalDuration)
}

func (b *Benchmark) runMixedRequests(client *http.Client, port int, protocol, scenario string) BenchmarkResult {
	urls := []string{
		fmt.Sprintf("https://localhost:%d/ping", port),
		fmt.Sprintf("https://localhost:%d/json", port),
		fmt.Sprintf("https://localhost:%d/data?size=10240", port),
		fmt.Sprintf("https://localhost:%d/data?size=102400", port),
	}

	latencies := make([]time.Duration, 0, 200)
	var totalBytes int64
	var successCount, failCount int
	var mu sync.Mutex

	start := time.Now()

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for i := 0; i < 200; i++ {
		wg.Add(1)
		sem <- struct{}{}

		url := urls[i%len(urls)]

		go func(url string) {
			defer wg.Done()
			defer func() { <-sem }()

			reqStart := time.Now()
			resp, err := client.Get(url)
			if err != nil {
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			reqDuration := time.Since(reqStart)

			mu.Lock()
			if err != nil || resp.StatusCode != 200 {
				failCount++
			} else {
				successCount++
				latencies = append(latencies, reqDuration)
				totalBytes += int64(len(body))
			}
			mu.Unlock()
		}(url)
	}

	wg.Wait()
	totalDuration := time.Since(start)

	return b.calculateResults(protocol, scenario, latencies, totalBytes, successCount, failCount, totalDuration)
}

func (b *Benchmark) calculateResults(protocol, scenario string, latencies []time.Duration, totalBytes int64, successCount, failCount int, totalDuration time.Duration) BenchmarkResult {
	if len(latencies) == 0 {
		return BenchmarkResult{
			Protocol:        protocol,
			Scenario:        scenario,
			TotalRequests:   successCount + failCount,
			SuccessRequests: successCount,
			FailedRequests:  failCount,
			TotalDuration:   totalDuration,
		}
	}

	var sum time.Duration
	min := latencies[0]
	max := latencies[0]

	for _, lat := range latencies {
		sum += lat
		if lat < min {
			min = lat
		}
		if lat > max {
			max = lat
		}
	}

	avg := sum / time.Duration(len(latencies))

	sortLatencies := make([]time.Duration, len(latencies))
	copy(sortLatencies, latencies)
	for i := 0; i < len(sortLatencies); i++ {
		for j := i + 1; j < len(sortLatencies); j++ {
			if sortLatencies[i] > sortLatencies[j] {
				sortLatencies[i], sortLatencies[j] = sortLatencies[j], sortLatencies[i]
			}
		}
	}

	p50 := sortLatencies[len(sortLatencies)*50/100]
	p95 := sortLatencies[len(sortLatencies)*95/100]
	p99 := sortLatencies[len(sortLatencies)*99/100]

	throughput := float64(totalBytes) / totalDuration.Seconds() / 1024 / 1024
	reqPerSec := float64(successCount) / totalDuration.Seconds()

	return BenchmarkResult{
		Protocol:        protocol,
		Scenario:        scenario,
		TotalRequests:   successCount + failCount,
		SuccessRequests: successCount,
		FailedRequests:  failCount,
		TotalDuration:   totalDuration,
		MinLatency:      min,
		MaxLatency:      max,
		AvgLatency:      avg,
		P50Latency:      p50,
		P95Latency:      p95,
		P99Latency:      p99,
		Throughput:      throughput,
		RequestsPerSec:  reqPerSec,
		TotalBytesRecv:  totalBytes,
	}
}
