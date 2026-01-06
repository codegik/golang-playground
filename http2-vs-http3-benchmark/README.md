# HTTP/2 vs HTTP/3 Performance Benchmark

A comprehensive benchmark suite comparing HTTP/2 and HTTP/3 performance across multiple scenarios.

## What This Benchmarks

This tool runs real-world performance tests comparing HTTP/2 (over TCP+TLS) and HTTP/3 (over QUIC) protocols:

### Test Scenarios

1. **Single Request Latency** - Measures round-trip time for individual requests
2. **Concurrent Requests** - Tests performance with 10 simultaneous requests
3. **Large File Download** - Downloads 10MB files to test throughput
4. **Many Small Requests** - 500 requests of 1KB each (simulates web page assets)
5. **Mixed Workload** - Combination of different request sizes and types

### Metrics Collected

For each scenario, the benchmark measures:
- **Latency**: Min, Average, P50, P95, P99, Max
- **Throughput**: MB/s for data transfer
- **Requests/sec**: Request handling rate
- **Success/Failure rate**: Request reliability

## Quick Start

```bash
cd /Users/iklassman/sources/golang-playground/http2-vs-http3-benchmark

go mod download

go run . -mode benchmark
```

This will:
1. Start HTTP/2 server on port 8443
2. Start HTTP/3 server on port 9443
3. Run all benchmark scenarios
4. Display comparative results

## Usage

### Run Full Benchmark (Recommended)

```bash
go run . -mode benchmark
```

### Run Servers Separately

Terminal 1 (HTTP/2):
```bash
go run . -mode http2-server -http2-port 8443
```

Terminal 2 (HTTP/3):
```bash
go run . -mode http3-server -http3-port 9443
```

Terminal 3 (Benchmark):
```bash
go run . -mode benchmark -http2-port 8443 -http3-port 9443
```

### Custom Ports

```bash
go run . -mode benchmark -http2-port 7443 -http3-port 8443
```

## Build Binary

```bash
go build -o benchmark .

./benchmark -mode benchmark
```

## Understanding the Results

### Output Format

The benchmark produces three sections:

#### 1. Detailed Results by Scenario
```
[Single Request Latency]
HTTP/2:
  Total Requests:     100
  Success:            100
  Requests/sec:       45.23

  Latency:
    Avg:              22ms
    P95:              35ms
    P99:              42ms

HTTP/3:
  Total Requests:     100
  Success:            100
  Requests/sec:       52.18

  Latency:
    Avg:              19ms
    P95:              28ms
    P99:              33ms
```

#### 2. Comparative Analysis
```
HTTP/3 vs HTTP/2:
  Average Latency:     ↓ 13.6% better (HTTP/2: 22ms, HTTP/3: 19ms)
  P95 Latency:         ↓ 20.0% better (HTTP/2: 35ms, HTTP/3: 28ms)
  Requests/sec:        ↑ 15.4% better (HTTP/2: 45.23, HTTP/3: 52.18)
```

#### 3. Winner by Scenario
```
Single Request Latency                   Winner: HTTP/3
Concurrent Requests (10)                 Winner: HTTP/3
Large File (10MB)                        Winner: HTTP/3
Many Small Requests (500x1KB)            Winner: HTTP/3
Mixed Workload                           Winner: HTTP/3

Overall Score:
  HTTP/2: 0 wins
  HTTP/3: 5 wins

🏆 HTTP/3 is the overall winner!
```

## When HTTP/3 Wins

HTTP/3 typically performs better in:

✓ **High latency networks** - 0-RTT connection resumption eliminates round trips
✓ **Packet loss scenarios** - Independent streams avoid head-of-line blocking
✓ **Mobile networks** - Connection migration survives network switches
✓ **Concurrent requests** - True multiplexing without TCP blocking
✓ **Many small requests** - Lower overhead per request

## When HTTP/2 Might Win

HTTP/2 can be competitive in:

- **Low latency, stable networks** - When network is perfect, TCP overhead is minimal
- **CPU-constrained scenarios** - QUIC encryption is slightly more CPU intensive
- **First connection** - No 0-RTT benefit on initial handshake
- **Very large files on perfect network** - TCP optimization mature

## Expected Results

On typical networks, you should see:

| Metric | HTTP/3 Advantage |
|--------|-----------------|
| Average Latency | 10-20% faster |
| P95/P99 Latency | 15-30% faster |
| Requests/sec | 10-25% higher |
| Throughput (large files) | 5-15% higher |

**Note:** Results vary based on network conditions, hardware, and OS network stack.

## Architecture

```
http2-vs-http3-benchmark/
├── main.go             - Entry point and CLI
├── http2_server.go     - HTTP/2 server implementation
├── http3_server.go     - HTTP/3 server implementation
├── benchmark.go        - Benchmark execution and metrics collection
├── results.go          - Results analysis and reporting
├── go.mod              - Go module dependencies
└── README.md           - This file
```

### Server Endpoints

Both servers implement identical endpoints:

- `GET /ping` - Simple ping/pong response
- `GET /json` - Returns JSON with timestamp
- `GET /data?size=N` - Returns N bytes of data
- `GET /delay?ms=N` - Delays N milliseconds before responding

## Technical Details

### HTTP/2 Implementation
- TLS 1.3
- TCP transport
- Connection multiplexing
- Server push disabled (for fair comparison)

### HTTP/3 Implementation
- TLS 1.3 integrated with QUIC
- UDP transport
- 0-RTT enabled for connection resumption
- Independent stream multiplexing

### Benchmark Methodology
- Warmup period before measurements
- Multiple iterations for statistical significance
- Concurrent load simulation
- Percentile calculations for latency distribution

## Limitations

This benchmark:
- Runs on localhost (no real network latency/loss)
- Uses self-signed certificates
- Measures application-level performance
- Does not simulate middlebox interference

For realistic results with network conditions, consider:
- Running servers on separate machines
- Using tools like `tc` (traffic control) for latency/loss simulation
- Testing across real internet connections

## Dependencies

- `github.com/quic-go/quic-go` - HTTP/3 and QUIC implementation
- Go 1.21+ standard library

## Interpreting Results for Your Use Case

**For Web Applications:**
- Focus on "Many Small Requests" scenario
- Look at P95/P99 latencies (user experience)

**For API Services:**
- "Concurrent Requests" scenario most relevant
- Requests/sec indicates scaling capability

**For File Downloads:**
- "Large File Download" shows throughput
- HTTP/3 advantage smaller on perfect networks

**For Mobile Apps:**
- HTTP/3 will show larger real-world advantage
- Connection migration not tested here (requires network switching)

## Contributing

To add new test scenarios:

1. Add test method in `benchmark.go`
2. Register in `RunAll()` scenarios list
3. Ensure both HTTP/2 and HTTP/3 tested identically

## License

MIT
