# HTTP/3 Deep Dive: Under the Hood

This is a comprehensive implementation showcasing the internal workings of HTTP/3 and QUIC protocol. Built for technical talks that go beyond the basics.

## What This Covers

This project demonstrates the deepest technical aspects of HTTP/3:

1. **QUIC Packet Structure** - Anatomy of Initial, Handshake, 0-RTT, and 1-RTT packets
2. **Encryption & TLS 1.3 Integration** - Key derivation, AEAD, header protection
3. **0-RTT Connection Establishment** - Session tickets, replay protection, early data
4. **Stream Multiplexing** - Independent streams without head-of-line blocking
5. **Connection Migration** - Network switching, path validation, NAT rebinding
6. **Instrumented Server/Client** - Real HTTP/3 with deep protocol logging

## Prerequisites

```bash
go 1.21+
```

## Quick Start

```bash
cd http3-deep-dive

go mod download

go run .
```

This runs all demonstrations in sequence. For individual demos:

```bash
go run . -mode packets
go run . -mode encryption
go run . -mode 0rtt
go run . -mode multiplex
go run . -mode migration
```

## Live Server/Client

Terminal 1 (Server):
```bash
go run . -mode server
```

Terminal 2 (Client):
```bash
go run . -mode client
```

Watch the packet-level protocol exchange in real-time!

## Talk Flow

### Part 1: Packet Structure (5 min)

**Key Points:**
- Long header vs short header
- Connection IDs decouple connection from 4-tuple
- Frames are multiplexed within packets
- Variable-length packet numbers

**Live Demo:**
```bash
go run . -mode packets
```

**What to Show:**
- Initial packet with ClientHello in CRYPTO frame
- 0-RTT packet with application data
- Path Challenge/Response frames for migration

### Part 2: Encryption (7 min)

**Key Points:**
- Initial keys derived from Destination Connection ID
- TLS 1.3 is integrated, not layered
- Four encryption levels: Initial, Handshake, 0-RTT, 1-RTT
- Header protection hides packet numbers
- HKDF key derivation

**Live Demo:**
```bash
go run . -mode encryption
```

**What to Show:**
- Initial salt and key derivation
- AEAD encryption process (AES-128-GCM)
- Header protection masking
- How packet number is XORed into nonce

### Part 3: 0-RTT Handshake (8 min)

**Key Points:**
- Reduces connection time by 1 full RTT
- Session tickets enable resumption
- Early data sent with first flight
- Replay protection mechanisms
- Only for idempotent requests

**Live Demo:**
```bash
go run . -mode 0rtt
```

**What to Show:**
- Initial 1-RTT connection flow
- Server issuing session ticket
- Resumption with 0-RTT packet containing HTTP request
- Timeline comparison: 1-RTT vs 0-RTT

### Part 4: Stream Multiplexing (8 min)

**Key Points:**
- Each stream is independent at packet level
- No head-of-line blocking (unlike HTTP/2 over TCP)
- Packet loss only affects streams in that packet
- Per-stream and connection-level flow control

**Live Demo:**
```bash
go run . -mode multiplex
```

**What to Show:**
- HTTP/2: One lost packet blocks ALL streams
- HTTP/3: Lost packet only blocks affected stream
- Flow control with MAX_DATA and MAX_STREAM_DATA frames
- Parallel stream transmission

### Part 5: Connection Migration (10 min)

**Key Points:**
- Connection survives IP address changes
- Connection IDs enable migration
- Path validation with challenge-response
- NAT rebinding handled automatically
- Stateless reset for lost state

**Live Demo:**
```bash
go run . -mode migration
```

**What to Show:**
- WiFi to cellular migration
- PATH_CHALLENGE and PATH_RESPONSE frames
- NEW_CONNECTION_ID and RETIRE_CONNECTION_ID
- NAT rebinding scenario
- Stateless reset mechanism

### Part 6: Live Server/Client (12 min)

**Key Points:**
- Real HTTP/3 traffic with instrumentation
- See every packet, frame, and state change
- RTT measurement, congestion control
- Lost packet detection and recovery

**Live Demo:**

Terminal 1:
```bash
go run . -mode server
```

Terminal 2:
```bash
go run . -mode client
```

**What to Show:**
- Connection establishment sequence
- Initial, Handshake, 1-RTT packets
- HTTP/3 request/response on streams
- ACK frames and packet number acknowledgment
- RTT and congestion window updates
- Connection closure

## Architecture

```
packet_analyzer.go   - QUIC packet structure and frame types
encryption.go        - TLS 1.3 key derivation and AEAD encryption
zerortt.go          - 0-RTT handshake and session tickets
multiplexing.go     - Stream multiplexing and flow control
migration.go        - Connection migration and path validation
server.go           - Instrumented HTTP/3 server
client.go           - Instrumented HTTP/3 client
main.go             - CLI entrypoint
```

## Key Differentiators: HTTP/3 vs HTTP/2

| Feature | HTTP/2 (TCP) | HTTP/3 (QUIC) |
|---------|-------------|---------------|
| Transport | TCP | UDP |
| Handshake | TCP + TLS (2-3 RTT) | QUIC+TLS (1 RTT, 0 with resumption) |
| Head-of-line blocking | Yes (at TCP layer) | No (independent streams) |
| Connection migration | Breaks connection | Seamless with Connection IDs |
| Encryption | Optional (TLS on top) | Built-in (TLS 1.3 integrated) |
| Packet loss recovery | Affects all streams | Only affected streams |
| NAT rebinding | Breaks connection | Handled automatically |

## Deep Dive: Technical Details

### Why UDP?

TCP provides ordering guarantees that cause head-of-line blocking. When packet N is lost, packets N+1, N+2, etc. are held even if they contain data for different streams.

QUIC uses UDP and implements its own reliability with per-stream ordering. Stream 4 can make progress while Stream 8 waits for retransmission.

### Connection ID Purpose

In TCP, connections are identified by 4-tuple (src_ip, src_port, dst_ip, dst_port). Change your IP? Connection breaks.

QUIC uses Connection IDs - opaque identifiers independent of network path. Your phone switches from WiFi to 4G? Connection continues seamlessly.

### Why TLS 1.3?

TLS 1.3 reduces handshake to 1-RTT and enables 0-RTT resumption. QUIC integrates TLS 1.3 for key exchange, with encryption happening at the QUIC layer.

Initial packets use keys derived from the Destination Connection ID (public), allowing any server to decrypt. This enables connection migration and load balancing.

### Header Protection

Packet numbers are encrypted to prevent observers from tracking connections. A sample of the packet payload is used to generate a mask that XORs with the header.

This prevents middleboxes from interfering with congestion control or building flow tables.

### Frame Multiplexing

Multiple frame types can coexist in a single packet:
- STREAM frame (application data)
- ACK frame (acknowledgments)
- CRYPTO frame (handshake data)
- PATH_CHALLENGE (migration)

This reduces overhead and improves efficiency compared to TCP's single-purpose segments.

## Production Considerations

This is an educational implementation. For production:

- Use battle-tested libraries (quic-go, quiche, etc.)
- Implement proper congestion control (BBR, Cubic)
- Handle packet pacing and anti-amplification
- Implement connection ID rotation
- Add qlog support for debugging
- Handle PMTU discovery
- Implement connection coalescing

## Further Reading

- RFC 9000: QUIC Transport Protocol
- RFC 9001: Using TLS to Secure QUIC
- RFC 9002: QUIC Loss Detection and Congestion Control
- RFC 9114: HTTP/3
- RFC 9204: QPACK (Header Compression for HTTP/3)

## License

MIT
