package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// ProtocolVersion is the current protocol version
const ProtocolVersion = 1

// PacketType represents a packet type
type PacketType uint8

// Packet types
const (
	PacketTypeHandshake PacketType = iota
	PacketTypeCommand
	PacketTypeResponse
	PacketTypeData
	PacketTypeError
)

// EncryptionAlgorithm represents an encryption algorithm
type EncryptionAlgorithm uint8

// Encryption algorithms
const (
	EncryptionAlgorithmNone EncryptionAlgorithm = iota
	EncryptionAlgorithmAES
	EncryptionAlgorithmChacha20
)

// HeaderSize is the size of the packet header in bytes
const HeaderSize = 12

// PacketHeader represents a packet header
type PacketHeader struct {
	Version      uint8
	EncAlgorithm EncryptionAlgorithm
	Type         PacketType
	TaskID       uint32
	Checksum     uint32
}

// Packet represents a protocol packet
type Packet struct {
	Header PacketHeader
	Data   []byte
}

// EncodePacket encodes a packet into a byte slice
func EncodePacket(packet *Packet) ([]byte, error) {
	// Calculate total size
	totalSize := HeaderSize + len(packet.Data)
	
	// Create buffer
	buffer := bytes.NewBuffer(make([]byte, 0, totalSize))
	
	// Write header
	err := binary.Write(buffer, binary.LittleEndian, packet.Header.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to write version: %w", err)
	}
	
	err = binary.Write(buffer, binary.LittleEndian, packet.Header.EncAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to write encryption algorithm: %w", err)
	}
	
	err = binary.Write(buffer, binary.LittleEndian, packet.Header.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to write type: %w", err)
	}
	
	err = binary.Write(buffer, binary.LittleEndian, packet.Header.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to write task ID: %w", err)
	}
	
	err = binary.Write(buffer, binary.LittleEndian, packet.Header.Checksum)
	if err != nil {
		return nil, fmt.Errorf("failed to write checksum: %w", err)
	}
	
	// Write data
	_, err = buffer.Write(packet.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to write data: %w", err)
	}
	
	return buffer.Bytes(), nil
}

// DecodePacket decodes a byte slice into a packet
func DecodePacket(data []byte) (*Packet, error) {
	// Check if data is long enough
	if len(data) < HeaderSize {
		return nil, fmt.Errorf("data too short")
	}
	
	// Create buffer
	buffer := bytes.NewBuffer(data)
	
	// Create packet
	packet := &Packet{}
	
	// Read header
	err := binary.Read(buffer, binary.LittleEndian, &packet.Header.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}
	
	err = binary.Read(buffer, binary.LittleEndian, &packet.Header.EncAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to read encryption algorithm: %w", err)
	}
	
	err = binary.Read(buffer, binary.LittleEndian, &packet.Header.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to read type: %w", err)
	}
	
	err = binary.Read(buffer, binary.LittleEndian, &packet.Header.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to read task ID: %w", err)
	}
	
	err = binary.Read(buffer, binary.LittleEndian, &packet.Header.Checksum)
	if err != nil {
		return nil, fmt.Errorf("failed to read checksum: %w", err)
	}
	
	// Read data
	packet.Data = make([]byte, buffer.Len())
	_, err = buffer.Read(packet.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}
	
	return packet, nil
}
