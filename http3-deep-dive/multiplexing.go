package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Stream struct {
	ID       uint64
	Priority uint8
	State    string
	BytesSent int
	BytesRecv int
	Blocked   bool
	StartTime time.Time
	EndTime   time.Time
}

func (s *Stream) Duration() time.Duration {
	if s.EndTime.IsZero() {
		return time.Since(s.StartTime)
	}
	return s.EndTime.Sub(s.StartTime)
}

func (s *Stream) String() string {
	status := "✓"
	if s.Blocked {
		status = "⚠"
	}
	if s.State != "Completed" {
		status = "⏳"
	}

	return fmt.Sprintf("Stream %2d %s | State: %-10s | Sent: %5d | Recv: %5d | Time: %4dms",
		s.ID, status, s.State, s.BytesSent, s.BytesRecv,
		s.Duration().Milliseconds())
}

type PacketLoss struct {
	PacketNumber uint64
	StreamID     uint64
	DetectedAt   time.Time
}

func DemonstrateMultiplexing() {
	fmt.Println("\n═══════════════════════════════════════════════════")
	fmt.Println("     Stream Multiplexing: HTTP/3 vs HTTP/2")
	fmt.Println("═══════════════════════════════════════════════════")

	fmt.Println("\n[HTTP/2 over TCP - Head-of-Line Blocking]")
	fmt.Println("─────────────────────────────────────────────────")
	demonstrateHTTP2HOL()

	fmt.Println("\n\n[HTTP/3 over QUIC - No Head-of-Line Blocking]")
	fmt.Println("─────────────────────────────────────────────────")
	demonstrateHTTP3NoHOL()

	fmt.Println("\n═══════════════════════════════════════════════════")
	fmt.Println("     Per-Stream Flow Control")
	fmt.Println("═══════════════════════════════════════════════════")
	demonstrateFlowControl()
}

func demonstrateHTTP2HOL() {
	streams := []*Stream{
		{ID: 1, State: "Active", Priority: 1, StartTime: time.Now()},
		{ID: 3, State: "Active", Priority: 1, StartTime: time.Now()},
		{ID: 5, State: "Active", Priority: 1, StartTime: time.Now()},
		{ID: 7, State: "Active", Priority: 1, StartTime: time.Now()},
	}

	fmt.Println("\nInitial state: 4 concurrent streams")
	for _, s := range streams {
		fmt.Printf("  %s\n", s)
	}

	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n[TCP Packet Loss Detected!]")
	fmt.Println("  Packet #42 lost (contains Stream 3 data)")
	fmt.Println("  TCP must retransmit and wait for ACK...")

	streams[0].State = "Blocked"
	streams[0].Blocked = true
	streams[1].State = "Blocked"
	streams[1].Blocked = true
	streams[2].State = "Blocked"
	streams[2].Blocked = true
	streams[3].State = "Blocked"
	streams[3].Blocked = true

	fmt.Println("\n⚠️  ALL streams blocked (Head-of-Line Blocking):")
	for _, s := range streams {
		fmt.Printf("  %s\n", s)
	}

	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n[Retransmission successful after 200ms]")
	for _, s := range streams {
		s.State = "Active"
		s.Blocked = false
		s.BytesSent = 5000 + rand.Intn(3000)
	}

	for _, s := range streams {
		fmt.Printf("  %s\n", s)
	}

	fmt.Println("\n❌ Problem: Even though only Stream 3 lost data,")
	fmt.Println("   ALL streams were blocked at TCP layer")
}

func demonstrateHTTP3NoHOL() {
	streams := []*Stream{
		{ID: 0, State: "Active", Priority: 1, StartTime: time.Now()},
		{ID: 4, State: "Active", Priority: 1, StartTime: time.Now()},
		{ID: 8, State: "Active", Priority: 1, StartTime: time.Now()},
		{ID: 12, State: "Active", Priority: 1, StartTime: time.Now()},
	}

	fmt.Println("\nInitial state: 4 concurrent streams")
	for _, s := range streams {
		fmt.Printf("  %s\n", s)
	}

	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n[QUIC Packet Loss Detected!]")
	loss := PacketLoss{
		PacketNumber: 42,
		StreamID:     4,
		DetectedAt:   time.Now(),
	}
	fmt.Printf("  Packet #%d lost (contains Stream %d data)\n",
		loss.PacketNumber, loss.StreamID)
	fmt.Println("  QUIC retransmits only Stream 4 data...")

	streams[1].State = "Recovering"
	streams[1].Blocked = true

	fmt.Println("\n✓ Only affected stream blocked:")
	for _, s := range streams {
		if s.ID != 4 {
			s.BytesSent = 8000 + rand.Intn(2000)
		}
		fmt.Printf("  %s\n", s)
	}

	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n[Stream 4 recovery complete]")
	streams[1].State = "Active"
	streams[1].Blocked = false
	streams[1].BytesSent = 7500

	for _, s := range streams {
		fmt.Printf("  %s\n", s)
	}

	fmt.Println("\n✓ Benefits:")
	fmt.Println("  • Independent stream recovery")
	fmt.Println("  • No cascade blocking")
	fmt.Println("  • Better throughput under packet loss")
}

func demonstrateFlowControl() {
	fmt.Println("\nQUIC has TWO levels of flow control:")
	fmt.Println("  1. Connection-level: MAX_DATA frame")
	fmt.Println("  2. Stream-level: MAX_STREAM_DATA frame")

	connectionLimit := uint64(1000000)
	streamLimits := map[uint64]uint64{
		0:  100000,
		4:  100000,
		8:  100000,
		12: 100000,
	}

	streams := []*Stream{
		{ID: 0, State: "Active", BytesSent: 0, StartTime: time.Now()},
		{ID: 4, State: "Active", BytesSent: 0, StartTime: time.Now()},
		{ID: 8, State: "Active", BytesSent: 0, StartTime: time.Now()},
		{ID: 12, State: "Active", BytesSent: 0, StartTime: time.Now()},
	}

	connectionUsed := uint64(0)

	fmt.Println("\nInitial limits:")
	fmt.Printf("  Connection: %d bytes\n", connectionLimit)
	for id, limit := range streamLimits {
		fmt.Printf("  Stream %d: %d bytes\n", id, limit)
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	fmt.Println("\n[Simulating parallel stream transmission]")

	for _, stream := range streams {
		wg.Add(1)
		go func(s *Stream) {
			defer wg.Done()

			for i := 0; i < 5; i++ {
				chunkSize := uint64(25000)

				mu.Lock()
				if connectionUsed+chunkSize > connectionLimit {
					s.State = "ConnBlocked"
					s.Blocked = true
					mu.Unlock()
					time.Sleep(50 * time.Millisecond)
					continue
				}

				if uint64(s.BytesSent)+chunkSize > streamLimits[s.ID] {
					s.State = "StreamBlocked"
					s.Blocked = true
					mu.Unlock()
					time.Sleep(50 * time.Millisecond)
					continue
				}

				s.BytesSent += int(chunkSize)
				connectionUsed += chunkSize
				s.State = "Active"
				s.Blocked = false
				mu.Unlock()

				time.Sleep(time.Duration(20+rand.Intn(30)) * time.Millisecond)
			}

			mu.Lock()
			s.State = "Completed"
			s.EndTime = time.Now()
			mu.Unlock()
		}(stream)
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	iteration := 0
	for {
		select {
		case <-ticker.C:
			iteration++
			mu.Lock()
			fmt.Printf("\n[Update %d] Connection used: %d/%d bytes\n",
				iteration, connectionUsed, connectionLimit)
			for _, s := range streams {
				fmt.Printf("  %s\n", s)
			}
			mu.Unlock()

		case <-done:
			ticker.Stop()
			mu.Lock()
			fmt.Printf("\n[Final] Connection used: %d/%d bytes\n",
				connectionUsed, connectionLimit)
			for _, s := range streams {
				fmt.Printf("  %s\n", s)
			}
			mu.Unlock()
			return
		}
	}
}
