package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type SessionTicket struct {
	Ticket           []byte
	ServerName       string
	IssuedAt         time.Time
	Lifetime         time.Duration
	EarlyDataAllowed bool
	MaxEarlyData     uint32
	ALPN             string
	ResumptionSecret []byte
}

func (t *SessionTicket) IsValid() bool {
	return time.Since(t.IssuedAt) < t.Lifetime
}

func (t *SessionTicket) String() string {
	s := "\n┌─ Session Ticket ─────────────────────────────────┐\n"
	s += fmt.Sprintf("│ Server: %s\n", t.ServerName)
	s += fmt.Sprintf("│ Issued: %s\n", t.IssuedAt.Format("15:04:05"))
	s += fmt.Sprintf("│ Lifetime: %v\n", t.Lifetime)
	s += fmt.Sprintf("│ Early Data Allowed: %v\n", t.EarlyDataAllowed)
	s += fmt.Sprintf("│ Max Early Data: %d bytes\n", t.MaxEarlyData)
	s += fmt.Sprintf("│ ALPN: %s\n", t.ALPN)
	s += fmt.Sprintf("│ Ticket: %s...\n", hex.EncodeToString(t.Ticket[:16]))
	s += "└──────────────────────────────────────────────────┘\n"
	return s
}

func CreateSessionTicket(serverName string) *SessionTicket {
	ticket := make([]byte, 32)
	secret := make([]byte, 32)
	rand.Read(ticket)
	rand.Read(secret)

	return &SessionTicket{
		Ticket:           ticket,
		ServerName:       serverName,
		IssuedAt:         time.Now(),
		Lifetime:         24 * time.Hour,
		EarlyDataAllowed: true,
		MaxEarlyData:     0xffffffff,
		ALPN:             "h3",
		ResumptionSecret: secret,
	}
}

type ConnectionState struct {
	Phase           string
	RTT             int
	PacketsSent     int
	BytesSent       int
	EncryptionLevel EncryptionLevel
}

func (cs *ConnectionState) String() string {
	return fmt.Sprintf("[%s | RTT:%d | Packets:%d | Bytes:%d | Encryption:%s]",
		cs.Phase, cs.RTT, cs.PacketsSent, cs.BytesSent, cs.EncryptionLevel)
}

func Demonstrate0RTT() {
	fmt.Println("\n═══════════════════════════════════════════════════")
	fmt.Println("     0-RTT Connection Establishment")
	fmt.Println("═══════════════════════════════════════════════════")

	fmt.Println("\n[STEP 1] Initial Connection (1-RTT)")
	fmt.Println("─────────────────────────────────────────────────")

	destConnID := make([]byte, 8)
	srcConnID := make([]byte, 8)
	rand.Read(destConnID)
	rand.Read(srcConnID)

	state := &ConnectionState{
		Phase:           "Initial",
		RTT:             0,
		PacketsSent:     0,
		BytesSent:       0,
		EncryptionLevel: EncryptionInitial,
	}

	fmt.Printf("\nClient %s\n", state)
	initialPkt := CreateInitialPacket()
	fmt.Println(initialPkt)

	cryptoFrame := &Frame{
		Type:   FrameTypeCrypto,
		Offset: 0,
		Data:   []byte("ClientHello + TLS extensions"),
	}
	fmt.Println("Sending ClientHello in CRYPTO frame:")
	fmt.Print(cryptoFrame)

	state.Phase = "WaitingServerHello"
	state.RTT = 1
	state.PacketsSent = 1
	fmt.Printf("\nClient %s\n", state)

	time.Sleep(50 * time.Millisecond)

	fmt.Println("\n[Server Response]")
	handshakePkt := CreateHandshakePacket(srcConnID, destConnID, 0)
	state.EncryptionLevel = EncryptionHandshake
	fmt.Println(handshakePkt)

	serverHelloFrame := &Frame{
		Type:   FrameTypeCrypto,
		Offset: 0,
		Data:   []byte("ServerHello + Certificate + CertificateVerify + Finished"),
	}
	fmt.Print(serverHelloFrame)

	time.Sleep(50 * time.Millisecond)

	state.Phase = "Connected"
	state.RTT = 2
	state.EncryptionLevel = Encryption1RTT
	fmt.Printf("\nClient %s\n", state)

	ticket := CreateSessionTicket("localhost:4433")
	fmt.Println("\n[Server Issues Session Ticket]")
	fmt.Print(ticket)

	fmt.Println("\n═════════════════════════════════════════════════")
	fmt.Println("\n[STEP 2] Resumption with 0-RTT")
	fmt.Println("─────────────────────────────────────────────────")

	if !ticket.IsValid() {
		fmt.Println("Ticket expired!")
		return
	}

	newDestConnID := make([]byte, 8)
	newSrcConnID := make([]byte, 8)
	rand.Read(newDestConnID)
	rand.Read(newSrcConnID)

	state = &ConnectionState{
		Phase:           "Resuming",
		RTT:             0,
		PacketsSent:     0,
		BytesSent:       0,
		EncryptionLevel: EncryptionInitial,
	}

	fmt.Printf("\nClient %s\n", state)
	fmt.Println("\nSending Initial packet WITH session ticket:")
	initialResumePkt := CreateInitialPacket()
	initialResumePkt.TokenLength = uint64(len(ticket.Ticket))
	initialResumePkt.Token = ticket.Ticket
	fmt.Println(initialResumePkt)

	resumeCryptoFrame := &Frame{
		Type:   FrameTypeCrypto,
		Offset: 0,
		Data:   []byte("ClientHello + PSK extension (ticket)"),
	}
	fmt.Print(resumeCryptoFrame)

	fmt.Println("\n[CRITICAL] Immediately sending 0-RTT packet (NO WAIT):")
	zeroRTTPkt := Create0RTTPacket(newDestConnID, newSrcConnID, 1)
	state.EncryptionLevel = Encryption0RTT
	fmt.Println(zeroRTTPkt)

	httpRequest := &Frame{
		Type:     FrameTypeStream,
		StreamID: 0,
		Offset:   0,
		Data:     []byte("GET /api/data HTTP/3\r\nHost: localhost\r\n\r\n"),
	}
	fmt.Println("0-RTT Application Data (HTTP/3 request):")
	fmt.Print(httpRequest)

	state.PacketsSent = 2
	state.BytesSent = 1200 + 1200
	fmt.Printf("\nClient %s\n", state)

	fmt.Println("\n⚡ KEY BENEFIT: Application data sent IMMEDIATELY")
	fmt.Println("   No waiting for server response!")
	fmt.Println("   Reduces latency by 1 full RTT")

	time.Sleep(50 * time.Millisecond)

	fmt.Println("\n[Server Response - 1 RTT later]")
	state.Phase = "Connected"
	state.RTT = 1
	state.EncryptionLevel = Encryption1RTT
	fmt.Printf("\nClient %s\n", state)

	fmt.Println("\nServer confirms 0-RTT acceptance:")
	handshakeDone := &Frame{
		Type: FrameTypeHandshakeDone,
	}
	fmt.Print(handshakeDone)

	responseStream := &Frame{
		Type:     FrameTypeStream,
		StreamID: 0,
		Offset:   0,
		Data:     []byte("HTTP/3 200 OK\r\n{\"data\": \"...\"}\r\n"),
	}
	fmt.Print(responseStream)

	fmt.Println("\n═════════════════════════════════════════════════")
	fmt.Println("\n[COMPARISON]")
	fmt.Println("─────────────────────────────────────────────────")
	fmt.Println("1-RTT handshake: 2 RTTs until application data")
	fmt.Println("0-RTT handshake: 0 RTTs until application data ⚡")
	fmt.Println("")
	fmt.Println("Perfect for:")
	fmt.Println("  • API requests")
	fmt.Println("  • Mobile reconnections")
	fmt.Println("  • High-latency networks")
	fmt.Println("\n⚠️  0-RTT Replay Protection:")
	fmt.Println("  • Server MUST reject duplicate tickets")
	fmt.Println("  • Use only for idempotent requests")
	fmt.Println("  • Anti-replay window tracking")
	fmt.Println("═════════════════════════════════════════════════")
}
