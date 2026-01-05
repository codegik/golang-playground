package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type ConnectionID struct {
	ID               []byte
	SequenceNumber   uint64
	RetireBy         uint64
	StatelessResetToken []byte
}

func (c *ConnectionID) String() string {
	return hex.EncodeToString(c.ID)
}

type NetworkPath struct {
	LocalAddr  string
	RemoteAddr string
	Active     bool
	RTT        time.Duration
	LossPct    float64
}

func (n *NetworkPath) String() string {
	status := "  "
	if n.Active {
		status = "✓ "
	}
	return fmt.Sprintf("%s%s -> %s (RTT: %dms, Loss: %.1f%%)",
		status, n.LocalAddr, n.RemoteAddr, n.RTT.Milliseconds(), n.LossPct)
}

type Connection struct {
	LocalConnIDs  []*ConnectionID
	RemoteConnIDs []*ConnectionID
	CurrentPath   *NetworkPath
	Paths         []*NetworkPath
	State         string
}

func NewConnection() *Connection {
	initialConnID := make([]byte, 8)
	rand.Read(initialConnID)

	return &Connection{
		LocalConnIDs: []*ConnectionID{
			{
				ID:             initialConnID,
				SequenceNumber: 0,
			},
		},
		RemoteConnIDs: []*ConnectionID{},
		State:         "Established",
	}
}

func (c *Connection) AddConnectionID(seqNum uint64) *ConnectionID {
	connID := make([]byte, 8)
	resetToken := make([]byte, 16)
	rand.Read(connID)
	rand.Read(resetToken)

	newCID := &ConnectionID{
		ID:                  connID,
		SequenceNumber:      seqNum,
		StatelessResetToken: resetToken,
	}

	c.LocalConnIDs = append(c.LocalConnIDs, newCID)
	return newCID
}

func DemonstrateConnectionMigration() {
	fmt.Println("\n═══════════════════════════════════════════════════")
	fmt.Println("     Connection Migration")
	fmt.Println("═══════════════════════════════════════════════════")

	fmt.Println("\n[SCENARIO] Mobile device moving from WiFi to Cellular")
	fmt.Println("─────────────────────────────────────────────────")

	conn := NewConnection()

	wifiPath := &NetworkPath{
		LocalAddr:  "192.168.1.100:51234",
		RemoteAddr: "203.0.113.10:443",
		Active:     true,
		RTT:        20 * time.Millisecond,
		LossPct:    0.1,
	}

	cellularPath := &NetworkPath{
		LocalAddr:  "10.20.30.40:51234",
		RemoteAddr: "203.0.113.10:443",
		Active:     false,
		RTT:        50 * time.Millisecond,
		LossPct:    1.0,
	}

	conn.CurrentPath = wifiPath
	conn.Paths = []*NetworkPath{wifiPath, cellularPath}

	serverConnID := make([]byte, 8)
	rand.Read(serverConnID)
	conn.RemoteConnIDs = append(conn.RemoteConnIDs, &ConnectionID{
		ID:             serverConnID,
		SequenceNumber: 0,
	})

	fmt.Println("\n[Step 1] Established connection on WiFi")
	fmt.Printf("Local Connection ID:  %s\n", conn.LocalConnIDs[0])
	fmt.Printf("Remote Connection ID: %s\n", conn.RemoteConnIDs[0])
	fmt.Println("\nActive paths:")
	for _, path := range conn.Paths {
		fmt.Printf("  %s\n", path)
	}

	fmt.Println("\n[Step 2] WiFi signal degrading...")
	time.Sleep(100 * time.Millisecond)
	wifiPath.RTT = 200 * time.Millisecond
	wifiPath.LossPct = 15.0

	fmt.Println("Path quality changed:")
	for _, path := range conn.Paths {
		fmt.Printf("  %s\n", path)
	}

	fmt.Println("\n[Step 3] Device acquires cellular IP address")
	fmt.Printf("New IP: %s\n", cellularPath.LocalAddr)

	newLocalConnID := conn.AddConnectionID(1)
	fmt.Printf("Generated new Connection ID: %s\n", newLocalConnID)

	fmt.Println("\n[Step 4] Path validation on new path")
	fmt.Println("\nSending PATH_CHALLENGE on cellular:")

	challenge := make([]byte, 8)
	rand.Read(challenge)

	pathChallengeFrame := &Frame{
		Type: FrameTypePathChallenge,
		Data: challenge,
	}
	fmt.Print(pathChallengeFrame)

	challengePkt := CreateInitialPacket()
	challengePkt.DestConnID = serverConnID
	challengePkt.SrcConnID = newLocalConnID.ID

	fmt.Println("\nPacket sent from NEW path:")
	fmt.Printf("  Source: %s (NEW)\n", cellularPath.LocalAddr)
	fmt.Printf("  Dest: %s\n", cellularPath.RemoteAddr)
	fmt.Printf("  Using Dest Conn ID: %s (unchanged)\n", hex.EncodeToString(serverConnID))
	fmt.Printf("  Using Src Conn ID: %s (NEW)\n", newLocalConnID)

	time.Sleep(50 * time.Millisecond)

	fmt.Println("\n[Step 5] Server validates new path")
	pathResponseFrame := &Frame{
		Type: FrameTypePathResponse,
		Data: challenge,
	}
	fmt.Print(pathResponseFrame)

	fmt.Println("\n✓ Path validation successful!")

	fmt.Println("\n[Step 6] Migrate to new path")
	wifiPath.Active = false
	cellularPath.Active = true
	conn.CurrentPath = cellularPath

	fmt.Println("\nConnection migrated:")
	for _, path := range conn.Paths {
		fmt.Printf("  %s\n", path)
	}

	fmt.Println("\n[Step 7] Retire old Connection ID")
	retireFrame := &Frame{
		Type: FrameTypeRetireConnectionID,
		Data: []byte{0},
	}
	fmt.Print(retireFrame)

	fmt.Printf("Retired Connection ID: %s\n", conn.LocalConnIDs[0])

	fmt.Println("\n═════════════════════════════════════════════════")
	fmt.Println("\n🔑 KEY INSIGHTS:")
	fmt.Println("─────────────────────────────────────────────────")
	fmt.Println("\n1. Connection IDs decouple connection from network path")
	fmt.Println("   • TCP: tuple (src_ip, src_port, dst_ip, dst_port)")
	fmt.Println("   • QUIC: Connection ID (independent of IP/port)")

	fmt.Println("\n2. Migration is transparent to application")
	fmt.Println("   • No broken connections")
	fmt.Println("   • No re-authentication")
	fmt.Println("   • Seamless for user")

	fmt.Println("\n3. Path validation prevents attacks")
	fmt.Println("   • Challenge-response mechanism")
	fmt.Println("   • Prevents address spoofing")
	fmt.Println("   • Confirms connectivity on new path")

	fmt.Println("\n4. Multiple Connection IDs")
	fmt.Println("   • Server provides spare IDs")
	fmt.Println("   • Client chooses which to use")
	fmt.Println("   • Load balancing friendly")

	fmt.Println("\n═════════════════════════════════════════════════")

	fmt.Println("\n[BONUS] Demonstrating NAT Rebinding")
	fmt.Println("─────────────────────────────────────────────────")

	demonstrateNATRebinding(conn)
}

func demonstrateNATRebinding(conn *Connection) {
	fmt.Println("\nScenario: Mobile device NAT timeout")

	oldPath := &NetworkPath{
		LocalAddr:  "10.0.0.5:51234",
		RemoteAddr: "203.0.113.10:443",
		Active:     true,
		RTT:        30 * time.Millisecond,
		LossPct:    0.5,
	}

	fmt.Printf("\nOriginal path: %s\n", oldPath)

	fmt.Println("\n[NAT binding expires after 30 seconds of inactivity]")
	time.Sleep(100 * time.Millisecond)

	newPath := &NetworkPath{
		LocalAddr:  "10.0.0.5:62891",
		RemoteAddr: "203.0.113.10:443",
		Active:     true,
		RTT:        30 * time.Millisecond,
		LossPct:    0.5,
	}

	fmt.Printf("New NAT binding: %s\n", newPath)

	fmt.Println("\nClient sends packet from new port:")
	fmt.Printf("  Using SAME Connection ID: %s\n", conn.RemoteConnIDs[0])
	fmt.Println("  Server sees packet from new 4-tuple")
	fmt.Println("  Server automatically updates path")

	fmt.Println("\n✓ Connection continues without interruption")
	fmt.Println("  • No PATH_CHALLENGE needed (same network)")
	fmt.Println("  • Server just updates return path")
	fmt.Println("  • This happens automatically in QUIC")

	fmt.Println("\n⚠️  TCP would break:")
	fmt.Println("   • Connection defined by 4-tuple")
	fmt.Println("   • NAT rebinding = new 4-tuple")
	fmt.Println("   • Connection RST, application must reconnect")
}

func DemonstrateStatelessReset() {
	fmt.Println("\n═══════════════════════════════════════════════════")
	fmt.Println("     Stateless Reset (Server Restart)")
	fmt.Println("═══════════════════════════════════════════════════")

	connID := make([]byte, 8)
	resetToken := make([]byte, 16)
	rand.Read(connID)
	rand.Read(resetToken)

	fmt.Println("\n[Step 1] Connection established")
	fmt.Printf("Connection ID: %s\n", hex.EncodeToString(connID))
	fmt.Printf("Stateless Reset Token: %s\n", hex.EncodeToString(resetToken[:8]))

	fmt.Println("\n[Step 2] Server crashes and restarts")
	fmt.Println("  • All connection state lost")
	fmt.Println("  • Server has no memory of this connection")

	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n[Step 3] Client sends packet")
	fmt.Printf("  Using Connection ID: %s\n", hex.EncodeToString(connID))

	fmt.Println("\n[Step 4] Server response")
	fmt.Println("  • Server doesn't recognize Connection ID")
	fmt.Println("  • Cannot send CONNECTION_CLOSE (no crypto state)")
	fmt.Println("  • Sends Stateless Reset instead")

	fmt.Println("\nStateless Reset packet structure:")
	fmt.Println("  ┌─ Short Header (looks like 1-RTT packet)")
	fmt.Printf("  │  Random bits: %s...\n", hex.EncodeToString(make([]byte, 16))[:20])
	fmt.Printf("  │  Stateless Reset Token: %s\n", hex.EncodeToString(resetToken))
	fmt.Println("  └─ (Last 16 bytes)")

	fmt.Println("\n[Step 5] Client recognizes reset token")
	fmt.Println("  ✓ Connection immediately closed")
	fmt.Println("  ✓ Application notified")
	fmt.Println("  ✓ Can establish new connection if needed")

	fmt.Println("\n🔑 Benefits:")
	fmt.Println("  • No state needed on server")
	fmt.Println("  • Prevents hung connections")
	fmt.Println("  • Fast failure detection")
	fmt.Println("═════════════════════════════════════════════════")
}
