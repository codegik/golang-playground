.PHONY: build run-all packets encryption 0rtt multiplex migration server client clean

build:
	go build -o http3-demo .

run-all: build
	./http3-demo -mode all

packets: build
	./http3-demo -mode packets

encryption: build
	./http3-demo -mode encryption

0rtt: build
	./http3-demo -mode 0rtt

multiplex: build
	./http3-demo -mode multiplex

migration: build
	./http3-demo -mode migration

server: build
	./http3-demo -mode server

client: build
	./http3-demo -mode client

test: build
	@echo "Testing packet structure..."
	@./http3-demo -mode packets > /dev/null && echo "✓ Packets OK"
	@echo "Testing encryption..."
	@./http3-demo -mode encryption > /dev/null && echo "✓ Encryption OK"
	@echo "Testing 0-RTT..."
	@./http3-demo -mode 0rtt > /dev/null && echo "✓ 0-RTT OK"
	@echo "Testing multiplexing..."
	@./http3-demo -mode multiplex > /dev/null && echo "✓ Multiplexing OK"
	@echo "Testing migration..."
	@./http3-demo -mode migration > /dev/null && echo "✓ Migration OK"
	@echo "All tests passed!"

clean:
	rm -f http3-demo

install-deps:
	go mod download
	go mod tidy

help:
	@echo "HTTP/3 Deep Dive - Make targets:"
	@echo ""
	@echo "  make build       - Build the binary"
	@echo "  make run-all     - Run all demonstrations"
	@echo "  make packets     - Show packet structure"
	@echo "  make encryption  - Show encryption process"
	@echo "  make 0rtt        - Show 0-RTT handshake"
	@echo "  make multiplex   - Show stream multiplexing"
	@echo "  make migration   - Show connection migration"
	@echo "  make server      - Run HTTP/3 server"
	@echo "  make client      - Run HTTP/3 client"
	@echo "  make test        - Run all demos (quick check)"
	@echo "  make clean       - Remove binaries"
	@echo "  make help        - Show this message"
