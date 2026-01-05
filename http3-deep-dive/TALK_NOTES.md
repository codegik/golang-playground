# HTTP/3 Talk Notes

## Opening (2 min)

"Today we're going deep into HTTP/3. Not what it is, but HOW it works. We'll look at packet structures, encryption, 0-RTT handshakes, stream multiplexing, and connection migration."

Show: `./http3-demo -mode all` running in background on second screen

## Part 1: Packet Structure (5 min)

### Key Message
QUIC packets are self-contained UDP datagrams with headers that enable connection migration and multiplexing.

### Run
```bash
./http3-demo -mode packets
```

### Talk Points
- Point to **Long Header** format
  - "Notice the Connection IDs - these are the key innovation"
  - "Version field allows protocol evolution"
  - "Four packet types: Initial, 0-RTT, Handshake, Retry"

- Point to **Frame multiplexing**
  - "CRYPTO + ACK + STREAM frames in single packet"
  - "More efficient than TCP's single-purpose segments"

- Point to **Packet Number**
  - "Never reused, per encryption level"
  - "Enables better loss detection than TCP sequence numbers"

### Visual Aid
Draw on whiteboard:
```
[UDP Datagram]
  ├─ QUIC Header (Connection IDs!)
  └─ Frames
      ├─ CRYPTO (handshake data)
      ├─ STREAM (application data)
      ├─ ACK (acknowledgments)
      └─ PATH_CHALLENGE (migration)
```

## Part 2: Encryption (7 min)

### Key Message
TLS 1.3 is integrated INTO QUIC, not layered on top. Initial keys derived from Connection ID enable any server to decrypt.

### Run
```bash
./http3-demo -mode encryption
```

### Talk Points
- Point to **Initial Salt**
  - "Standardized salt (RFC 9001)"
  - "Anyone can derive initial keys from Connection ID"
  - "Why? Enables load balancing and connection migration"

- Point to **HKDF Expansion**
  - "Keys, IV, and header protection key all derived"
  - "Separate keys for client and server"

- Point to **AEAD Encryption**
  - "AES-128-GCM: Authenticated Encryption with Associated Data"
  - "Nonce = IV XOR packet_number"
  - "Prevents replay attacks"

- Point to **Header Protection**
  - "Sample from ciphertext generates mask"
  - "XOR with header to hide packet number"
  - "Prevents observers from tracking flows"

### Ask Audience
"Why derive initial keys from Connection ID instead of random?"
Answer: "Load balancers can route without state! Any server can decrypt initial packets."

## Part 3: 0-RTT (8 min)

### Key Message
0-RTT eliminates round trip time for resumption, perfect for mobile and high-latency networks.

### Run
```bash
./http3-demo -mode 0rtt
```

### Talk Points
- Walk through **1-RTT first**
  - "Client → Initial (ClientHello)"
  - "Server → Handshake (ServerHello, certs, Finished)"
  - "Client → Handshake (Finished)"
  - "2 RTTs before application data"

- Point to **Session Ticket**
  - "Server provides resumption token"
  - "Contains encrypted state"
  - "Max early data limit: 4GB"

- Walk through **0-RTT resumption**
  - "Client → Initial WITH ticket"
  - "Client → 0-RTT WITH application data (HTTP request)"
  - "No waiting! Data sent immediately"
  - "1 RTT later: Server confirms"

### Critical Point
"⚠️ 0-RTT has replay risk!"
- "Only for idempotent operations"
- "Server MUST track and reject duplicate tickets"
- "Anti-replay window (8 seconds typical)"

### Use Case
"Perfect for:"
- Mobile reconnections (WiFi → 4G)
- High-latency networks (satellite, intercontinental)
- API calls (GET requests)

## Part 4: Stream Multiplexing (8 min)

### Key Message
Independent streams at packet level eliminate head-of-line blocking that plagues HTTP/2.

### Run
```bash
./http3-demo -mode multiplex
```

### Talk Points
- Show **HTTP/2 problem**
  - "TCP provides ordered byte stream"
  - "Packet loss blocks ALL streams"
  - "Even streams that received their data"

- Show **HTTP/3 solution**
  - "Each stream independent"
  - "Packet loss only affects contained streams"
  - "Other streams continue"

- Point to **Flow Control**
  - "Two levels: connection and stream"
  - "MAX_DATA: total connection limit"
  - "MAX_STREAM_DATA: per-stream limit"
  - "Receiver controls sender rate"

### Demo Impact
"Notice: In HTTP/2 sim, ALL streams stop. In HTTP/3, only affected stream stops."

### Whiteboard
Draw side-by-side comparison:
```
HTTP/2 (TCP)          HTTP/3 (QUIC)
─────────────         ─────────────
[Lost Pkt]            [Lost Pkt]
  ↓                     ↓
ALL STREAMS          ONLY STREAM 4
BLOCKED              BLOCKED
```

## Part 5: Connection Migration (10 min)

### Key Message
Connection IDs decouple connection from network 4-tuple, enabling seamless network changes.

### Run
```bash
./http3-demo -mode migration
```

### Talk Points
- **The Problem (TCP)**
  - "Connection = (src_ip, src_port, dst_ip, dst_port)"
  - "Change IP? Connection breaks"
  - "Mobile networks: disaster"

- **The Solution (QUIC)**
  - "Connection = Connection ID"
  - "Independent of IP/port"
  - "Server maintains mapping: CID → connection state"

- Walk through **Migration Flow**
  1. "Client on WiFi: 192.168.1.100"
  2. "WiFi degrading... RTT 200ms, 15% loss"
  3. "Client acquires cellular: 10.20.30.40"
  4. "Generate new Connection ID"
  5. "Send PATH_CHALLENGE from new IP"
  6. "Server validates with PATH_RESPONSE"
  7. "Migration complete! Old CID retired"

- Point to **Path Validation**
  - "Challenge-response prevents address spoofing"
  - "8-byte random challenge"
  - "Server echoes in PATH_RESPONSE"

- Show **NAT Rebinding**
  - "Port change without IP change"
  - "Automatically handled"
  - "No PATH_CHALLENGE needed"

### Show **Stateless Reset**
  - "Server crashes, loses state"
  - "Can't send proper CONNECTION_CLOSE (no keys)"
  - "Sends stateless reset instead"
  - "Last 16 bytes = reset token"
  - "Client immediately closes"

### Real-World Impact
"This is huge for mobile:"
- Phone switches WiFi → 4G: seamless
- Walk between WiFi APs: seamless
- NAT timeout: handled automatically
- No application errors!

## Part 6: Live HTTP/3 (12 min)

### Key Message
See real HTTP/3 traffic in action.

### Run
Terminal 1:
```bash
./http3-demo -mode server
```

Terminal 2:
```bash
./http3-demo -mode client
```

### Talk Through Output
- **Connection establishment**
  - "Initial packet with ClientHello"
  - "Handshake packet with ServerHello"
  - "1-RTT packets with application data"

- **HTTP/3 request**
  - "Sent on Stream 0"
  - "HEADERS frame + DATA frame"
  - "Server responds on same stream"

- **Multiple requests**
  - "Each on different stream"
  - "Multiplexed over single connection"
  - "Independent progress"

### Interactive
Ask someone to curl while server running:
```bash
curl --http3 https://localhost:4433/api/test --insecure
```

Watch the output together!

## Closing (3 min)

### Summary
"HTTP/3 isn't just faster HTTP. It's a complete reimagining:"
- ✓ UDP-based transport (no TCP head-of-line blocking)
- ✓ Integrated TLS 1.3 (faster handshakes, 0-RTT)
- ✓ Connection IDs (migration, load balancing)
- ✓ Independent streams (true multiplexing)
- ✓ Better loss recovery (packet numbers, not sequence)

### When to Use
"HTTP/3 shines for:"
- Mobile applications
- High packet loss networks
- Long-distance connections
- Real-time applications

### Production Adoption
"Who's using it?"
- Google (all services)
- Facebook/Meta (all apps)
- Cloudflare (default enabled)
- ~30% of web traffic already HTTP/3

### Resources
"Dive deeper:"
- RFC 9000 (QUIC)
- RFC 9114 (HTTP/3)
- This code: github.com/yourrepo

### Q&A
Prepare for common questions:
- "Why not fix TCP?" → Middlebox ossification, can't evolve
- "UDP firewall issues?" → Falling back to HTTP/2 works
- "Performance vs HTTP/2?" → Better on mobile, mixed on wired
- "CPU overhead?" → Slightly higher, improving with hardware offload

## Backup Slides

Have ready if time permits:

### QPACK (Header Compression)
- Dynamic table like HPACK
- But with stream independence
- Encoder/decoder streams separate

### Congestion Control
- BBR default (Google's algorithm)
- Cubic also supported
- Per-connection, not global like TCP

### Packet Pacing
- Prevents bursts
- Smoother network utilization
- Better for wireless

### Future
- QUIC v2 (extensions)
- Multipath QUIC (multiple networks simultaneously)
- Unreliable QUIC (for gaming)
