package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"os"
)

func main() {
	mode := flag.String("mode", "all", "Mode: all, packets, encryption, 0rtt, multiplex, migration, server, client")
	serverAddr := flag.String("server", "localhost:4433", "Server address for client mode")
	flag.Parse()

	switch *mode {
	case "packets":
		demonstratePackets()
	case "encryption":
		demonstrateEncryption()
	case "0rtt":
		Demonstrate0RTT()
	case "multiplex":
		DemonstrateMultiplexing()
	case "migration":
		DemonstrateConnectionMigration()
		DemonstrateStatelessReset()
	case "server":
		RunInstrumentedServer()
	case "client":
		paths := []string{"/api/users", "/api/data", "/api/stats"}
		if len(flag.Args()) > 0 {
			paths = flag.Args()
		}
		RunInstrumentedClient(*serverAddr, paths)
	case "all":
		demonstratePackets()
		demonstrateEncryption()
		Demonstrate0RTT()
		DemonstrateMultiplexing()
		DemonstrateConnectionMigration()
		DemonstrateStatelessReset()
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		flag.Usage()
		os.Exit(1)
	}
}

func demonstratePackets() {
	fmt.Println("\n\n")
	fmt.Println("███████████████████████████████████████████████████")
	fmt.Println("█                                                 █")
	fmt.Println("█        QUIC PACKET STRUCTURE ANATOMY            █")
	fmt.Println("█                                                 █")
	fmt.Println("███████████████████████████████████████████████████")

	destConnID := make([]byte, 8)
	srcConnID := make([]byte, 8)
	rand.Read(destConnID)
	rand.Read(srcConnID)

	fmt.Println("\n[1] Initial Packet (Client → Server)")
	initialPkt := CreateInitialPacket()
	cryptoFrame := &Frame{
		Type:   FrameTypeCrypto,
		Offset: 0,
		Data:   make([]byte, 200),
	}
	PrintPacketStructure(initialPkt, []*Frame{cryptoFrame})

	fmt.Println("\n[2] Handshake Packet (with multiple frames)")
	handshakePkt := CreateHandshakePacket(destConnID, srcConnID, 1)
	frames := []*Frame{
		{Type: FrameTypeCrypto, Offset: 0, Data: make([]byte, 100)},
		{Type: FrameTypeAck, Data: []byte{}},
	}
	PrintPacketStructure(handshakePkt, frames)

	fmt.Println("\n[3] 0-RTT Packet (Early Data)")
	zeroRTTPkt := Create0RTTPacket(destConnID, srcConnID, 2)
	streamFrame := &Frame{
		Type:     FrameTypeStream,
		StreamID: 0,
		Offset:   0,
		Data:     []byte("GET /api/data HTTP/3\r\n"),
	}
	PrintPacketStructure(zeroRTTPkt, []*Frame{streamFrame})

	fmt.Println("\n[4] Packet with Path Challenge (Connection Migration)")
	migrationPkt := CreateHandshakePacket(destConnID, srcConnID, 3)
	pathFrames := []*Frame{
		{Type: FrameTypePathChallenge, Data: make([]byte, 8)},
		{Type: FrameTypePathResponse, Data: make([]byte, 8)},
		{Type: FrameTypeNewConnectionID, Data: make([]byte, 16)},
	}
	PrintPacketStructure(migrationPkt, pathFrames)

	fmt.Println("\n🔑 KEY TAKEAWAYS:")
	fmt.Println("─────────────────────────────────────────────────")
	fmt.Println("• Long header: Contains version, connection IDs")
	fmt.Println("• Short header: Minimal overhead for 1-RTT packets")
	fmt.Println("• Connection ID: Decouples connection from IP/port")
	fmt.Println("• Frames: Multiplexed within single UDP datagram")
	fmt.Println("• Packet Number: Per encryption level, never reused")
}

func demonstrateEncryption() {
	fmt.Println("\n\n")
	fmt.Println("███████████████████████████████████████████████████")
	fmt.Println("█                                                 █")
	fmt.Println("█     QUIC ENCRYPTION & TLS 1.3 INTEGRATION       █")
	fmt.Println("█                                                 █")
	fmt.Println("███████████████████████████████████████████████████")

	connID := make([]byte, 8)
	rand.Read(connID)

	DemonstrateEncryption(connID)

	fmt.Println("\n🔑 KEY TAKEAWAYS:")
	fmt.Println("─────────────────────────────────────────────────")
	fmt.Println("• Initial keys: Derived from Dest Connection ID")
	fmt.Println("• Key hierarchy: Initial → Handshake → 0-RTT/1-RTT")
	fmt.Println("• AEAD: AES-128-GCM or ChaCha20-Poly1305")
	fmt.Println("• Header protection: Hides packet number from observers")
	fmt.Println("• TLS 1.3 integrated: Not layered on top")
}
