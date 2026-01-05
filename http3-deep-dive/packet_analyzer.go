package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

type PacketType uint8

const (
	PacketTypeInitial   PacketType = 0x00
	PacketType0RTT      PacketType = 0x01
	PacketTypeHandshake PacketType = 0x02
	PacketTypeRetry     PacketType = 0x03
)

type QuicPacketHeader struct {
	HeaderForm     uint8
	FixedBit       uint8
	LongPacketType PacketType
	TypeSpecific   uint8
	Version        uint32
	DestConnIDLen  uint8
	DestConnID     []byte
	SrcConnIDLen   uint8
	SrcConnID      []byte
	TokenLength    uint64
	Token          []byte
	Length         uint64
	PacketNumber   uint32
}

func (h *QuicPacketHeader) String() string {
	s := "\n┌─ QUIC Packet Header ─────────────────────────────┐\n"
	s += fmt.Sprintf("│ Header Form: %d (1=Long, 0=Short)\n", h.HeaderForm)
	s += fmt.Sprintf("│ Fixed Bit: %d (MUST be 1)        \n", h.FixedBit)

	if h.HeaderForm == 1 {
		var ptype string
		switch h.LongPacketType {
		case PacketTypeInitial:
			ptype = "Initial"
		case PacketType0RTT:
			ptype = "0-RTT"
		case PacketTypeHandshake:
			ptype = "Handshake"
		case PacketTypeRetry:
			ptype = "Retry"
		}
		s += fmt.Sprintf("│ Packet Type: %s\n", ptype)
		s += fmt.Sprintf("│ Version: 0x%08x (QUIC v1)\n", h.Version)
		s += fmt.Sprintf("│ Dest Conn ID Len: %d\n", h.DestConnIDLen)
		s += fmt.Sprintf("│ Dest Conn ID: %s\n", hex.EncodeToString(h.DestConnID))
		s += fmt.Sprintf("│ Src Conn ID Len: %d n", h.SrcConnIDLen)
		s += fmt.Sprintf("│ Src Conn ID: %s\n", hex.EncodeToString(h.SrcConnID))

		if h.LongPacketType == PacketTypeInitial {
			s += fmt.Sprintf("│ Token Length: %d\n", h.TokenLength)
			if h.TokenLength > 0 {
				s += fmt.Sprintf("│ Token: %s...\n", hex.EncodeToString(h.Token[:min(8, len(h.Token))]))
			}
		}
		s += fmt.Sprintf("│ Packet Length: %d bytes\n", h.Length)
	}
	s += fmt.Sprintf("│ Packet Number: %d\n", h.PacketNumber)
	s += "└──────────────────────────────────────────────────┘\n"
	return s
}

func CreateInitialPacket() *QuicPacketHeader {
	destConnID := make([]byte, 8)
	srcConnID := make([]byte, 8)
	rand.Read(destConnID)
	rand.Read(srcConnID)

	return &QuicPacketHeader{
		HeaderForm:     1,
		FixedBit:       1,
		LongPacketType: PacketTypeInitial,
		TypeSpecific:   0,
		Version:        0x00000001,
		DestConnIDLen:  8,
		DestConnID:     destConnID,
		SrcConnIDLen:   8,
		SrcConnID:      srcConnID,
		TokenLength:    0,
		Token:          []byte{},
		Length:         1200,
		PacketNumber:   0,
	}
}

func Create0RTTPacket(destConnID, srcConnID []byte, pn uint32) *QuicPacketHeader {
	return &QuicPacketHeader{
		HeaderForm:     1,
		FixedBit:       1,
		LongPacketType: PacketType0RTT,
		TypeSpecific:   0,
		Version:        0x00000001,
		DestConnIDLen:  uint8(len(destConnID)),
		DestConnID:     destConnID,
		SrcConnIDLen:   uint8(len(srcConnID)),
		SrcConnID:      srcConnID,
		Length:         1200,
		PacketNumber:   pn,
	}
}

func CreateHandshakePacket(destConnID, srcConnID []byte, pn uint32) *QuicPacketHeader {
	return &QuicPacketHeader{
		HeaderForm:     1,
		FixedBit:       1,
		LongPacketType: PacketTypeHandshake,
		TypeSpecific:   0,
		Version:        0x00000001,
		DestConnIDLen:  uint8(len(destConnID)),
		DestConnID:     destConnID,
		SrcConnIDLen:   uint8(len(srcConnID)),
		SrcConnID:      srcConnID,
		Length:         1200,
		PacketNumber:   pn,
	}
}

type FrameType uint8

const (
	FrameTypePadding            FrameType = 0x00
	FrameTypePing               FrameType = 0x01
	FrameTypeAck                FrameType = 0x02
	FrameTypeResetStream        FrameType = 0x04
	FrameTypeStopSending        FrameType = 0x05
	FrameTypeCrypto             FrameType = 0x06
	FrameTypeNewToken           FrameType = 0x07
	FrameTypeStream             FrameType = 0x08
	FrameTypeMaxData            FrameType = 0x10
	FrameTypeMaxStreamData      FrameType = 0x11
	FrameTypeMaxStreams         FrameType = 0x12
	FrameTypeDataBlocked        FrameType = 0x14
	FrameTypeStreamDataBlocked  FrameType = 0x15
	FrameTypeStreamsBlocked     FrameType = 0x16
	FrameTypeNewConnectionID    FrameType = 0x18
	FrameTypeRetireConnectionID FrameType = 0x19
	FrameTypePathChallenge      FrameType = 0x1a
	FrameTypePathResponse       FrameType = 0x1b
	FrameTypeConnectionClose    FrameType = 0x1c
	FrameTypeHandshakeDone      FrameType = 0x1e
)

type Frame struct {
	Type     FrameType
	Data     []byte
	StreamID uint64
	Offset   uint64
}

func (f *Frame) String() string {
	var typeName string
	switch f.Type {
	case FrameTypeCrypto:
		typeName = "CRYPTO"
	case FrameTypeStream:
		typeName = "STREAM"
	case FrameTypeAck:
		typeName = "ACK"
	case FrameTypePing:
		typeName = "PING"
	case FrameTypePathChallenge:
		typeName = "PATH_CHALLENGE"
	case FrameTypePathResponse:
		typeName = "PATH_RESPONSE"
	case FrameTypeNewConnectionID:
		typeName = "NEW_CONNECTION_ID"
	case FrameTypeHandshakeDone:
		typeName = "HANDSHAKE_DONE"
	default:
		typeName = fmt.Sprintf("0x%02x", f.Type)
	}

	s := fmt.Sprintf("  ├─ Frame Type: %s\n", typeName)
	if f.Type == FrameTypeStream {
		s += fmt.Sprintf("  │  Stream ID: %d\n", f.StreamID)
		s += fmt.Sprintf("  │  Offset: %d\n", f.Offset)
		s += fmt.Sprintf("  │  Length: %d bytes\n", len(f.Data))
	} else if f.Type == FrameTypeCrypto {
		s += fmt.Sprintf("  │  Offset: %d\n", f.Offset)
		s += fmt.Sprintf("  │  Length: %d bytes\n", len(f.Data))
	}
	return s
}

func EncodeVarint(v uint64) []byte {
	if v < 64 {
		return []byte{byte(v)}
	} else if v < 16384 {
		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, uint16(v))
		b[0] |= 0x40
		return b
	} else if v < 1073741824 {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(v))
		b[0] |= 0x80
		return b
	} else {
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, v)
		b[0] |= 0xc0
		return b
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func PrintPacketStructure(header *QuicPacketHeader, frames []*Frame) {
	fmt.Println(header)

	if len(frames) > 0 {
		fmt.Println("┌─ QUIC Packet Payload (Frames) ───────────────────┐")
		for i, frame := range frames {
			if i == len(frames)-1 {
				fmt.Print("  └")
			} else {
				fmt.Print("  │")
			}
			fmt.Print(frame)
		}
		fmt.Println("└──────────────────────────────────────────────────┘")
	}
}
