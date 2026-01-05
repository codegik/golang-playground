package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

type EncryptionLevel uint8

const (
	EncryptionInitial EncryptionLevel = iota
	EncryptionHandshake
	Encryption0RTT
	Encryption1RTT
)

func (e EncryptionLevel) String() string {
	switch e {
	case EncryptionInitial:
		return "Initial"
	case EncryptionHandshake:
		return "Handshake"
	case Encryption0RTT:
		return "0-RTT"
	case Encryption1RTT:
		return "1-RTT"
	default:
		return "Unknown"
	}
}

type Keys struct {
	Key []byte
	IV  []byte
	HP  []byte
}

func DeriveInitialKeys(connID []byte, isClient bool) *Keys {
	initialSalt := []byte{
		0x38, 0x76, 0x2c, 0xf7, 0xf5, 0x59, 0x34, 0xb3,
		0x4d, 0x17, 0x9a, 0xe6, 0xa4, 0xc8, 0x0c, 0xad,
		0xcc, 0xbb, 0x7f, 0x0a,
	}

	initialSecret := hkdf.Extract(sha256.New, connID, initialSalt)

	var label []byte
	if isClient {
		label = []byte("client in")
	} else {
		label = []byte("server in")
	}

	secret := hkdfExpandLabel(initialSecret, label, 32)
	key := hkdfExpandLabel(secret, []byte("quic key"), 16)
	iv := hkdfExpandLabel(secret, []byte("quic iv"), 12)
	hp := hkdfExpandLabel(secret, []byte("quic hp"), 16)

	return &Keys{
		Key: key,
		IV:  iv,
		HP:  hp,
	}
}

func hkdfExpandLabel(secret, label []byte, length int) []byte {
	hkdfLabel := make([]byte, 0, 2+1+len("tls13 ")+len(label)+1)
	hkdfLabel = append(hkdfLabel, byte(length>>8), byte(length))
	hkdfLabel = append(hkdfLabel, byte(len("tls13 ")+len(label)))
	hkdfLabel = append(hkdfLabel, []byte("tls13 ")...)
	hkdfLabel = append(hkdfLabel, label...)
	hkdfLabel = append(hkdfLabel, 0)

	out := make([]byte, length)
	r := hkdf.Expand(sha256.New, secret, hkdfLabel)
	io.ReadFull(r, out)

	return out
}

func EncryptPayload(payload, key, iv []byte, packetNumber uint64) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
	copy(nonce, iv)
	for i := 0; i < 8; i++ {
		nonce[12-1-i] ^= byte(packetNumber >> (8 * i))
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, payload, nil)
	return ciphertext, nil
}

func ProtectHeader(header, hp []byte, sample []byte) ([]byte, error) {
	block, err := aes.NewCipher(hp)
	if err != nil {
		return nil, err
	}

	mask := make([]byte, aes.BlockSize)
	block.Encrypt(mask, sample)

	protected := make([]byte, len(header))
	copy(protected, header)

	if (header[0] & 0x80) == 0x80 {
		protected[0] ^= mask[0] & 0x0f
	} else {
		protected[0] ^= mask[0] & 0x1f
	}

	pnOffset := len(header) - 4
	for i := 0; i < 4; i++ {
		protected[pnOffset+i] ^= mask[1+i]
	}

	return protected, nil
}

func PrintEncryptionDetails(level EncryptionLevel, keys *Keys, packetNumber uint64) {
	fmt.Printf("\n┌─ Encryption Details ─────────────────────────────┐\n")
	fmt.Printf("│ Level: %s\n", level)
	fmt.Printf("│ Packet Number: %d\n", packetNumber)
	fmt.Println("├──────────────────────────────────────────────────┤")
	fmt.Printf("│ Key:  %s\n", hex.EncodeToString(keys.Key))
	fmt.Printf("│ IV:   %s\n", hex.EncodeToString(keys.IV))
	fmt.Printf("│ HP:   %s\n", hex.EncodeToString(keys.HP))
	fmt.Println("├──────────────────────────────────────────────────┤")
	fmt.Println("│ Process:")
	fmt.Println("│ 1. Derive nonce: IV XOR packet_number")
	fmt.Println("│ 2. Encrypt payload: AES-128-GCM")
	fmt.Println("│ 3. Sample ciphertext for header protection")
	fmt.Println("│ 4. Protect header: AES-128-ECB mask")
	fmt.Println("└──────────────────────────────────────────────────┘")
}

func DemonstrateEncryption(connID []byte) {
	fmt.Println("\n═══════════════════════════════════════════════════")
	fmt.Println("     QUIC Packet Encryption Process")
	fmt.Println("═══════════════════════════════════════════════════")

	clientKeys := DeriveInitialKeys(connID, true)

	fmt.Println("\n[1] Initial Key Derivation (using Destination Connection ID)")
	fmt.Printf("    Connection ID: %s\n", hex.EncodeToString(connID))

	PrintEncryptionDetails(EncryptionInitial, clientKeys, 0)

	payload := []byte("CRYPTO frame with ClientHello")
	fmt.Printf("\n[2] Plaintext Payload (%d bytes):\n", len(payload))
	fmt.Printf("    %s\n", string(payload))

	ciphertext, _ := EncryptPayload(payload, clientKeys.Key, clientKeys.IV, 0)
	fmt.Printf("\n[3] Encrypted Payload (%d bytes):\n", len(ciphertext))
	fmt.Printf("    %s...\n", hex.EncodeToString(ciphertext[:min(32, len(ciphertext))]))

	fmt.Println("\n[4] Header Protection:")
	fmt.Println("    Sample 16 bytes from ciphertext")
	fmt.Println("    Generate mask using AES-ECB with HP key")
	fmt.Println("    XOR first byte and packet number bytes")

	fmt.Println("\n═══════════════════════════════════════════════════")
}
