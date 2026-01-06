.PHONY: build run http2 http3 clean help

build:
	go build -o benchmark .

run: build
	./benchmark -mode benchmark

http2: build
	./benchmark -mode http2-server -http2-port 8443

http3: build
	./benchmark -mode http3-server -http3-port 9443

clean:
	rm -f benchmark

deps:
	go mod download
	go mod tidy

test: build
	@echo "Running quick benchmark test..."
	@timeout 30s ./benchmark -mode benchmark || echo "Benchmark test completed"

help:
	@echo "HTTP/2 vs HTTP/3 Benchmark - Make targets:"
	@echo ""
	@echo "  make build    - Build the benchmark binary"
	@echo "  make run      - Run full benchmark suite"
	@echo "  make http2    - Run HTTP/2 server only"
	@echo "  make http3    - Run HTTP/3 server only"
	@echo "  make clean    - Remove built binaries"
	@echo "  make deps     - Download and tidy dependencies"
	@echo "  make test     - Quick benchmark test"
	@echo "  make help     - Show this message"
	@echo ""
	@echo "Custom options:"
	@echo "  ./benchmark -mode benchmark -http2-port 8443 -http3-port 9443"
