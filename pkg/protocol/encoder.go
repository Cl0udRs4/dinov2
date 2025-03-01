package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
)

// Constants for TLV structure
const (
	// TLV type constants
	TLVTypeCommand     byte = 1
	TLVTypeResponse    byte = 2
	TLVTypeError       byte = 3
	TLVTypeModuleData  byte = 4
	TLVTypeMetadata    byte = 5
	TLVTypeExtension   byte = 6
	TLVTypeReserved    byte = 7
	
	// Header constants
	HeaderSize         int  = 12 // Version(1) + EncAlgorithm(1) + Type(1) + TaskID(4) + Checksum(4) + Reserved(1)
	MaxPacketSize      int  = 65535
	MaxFragmentSize    int  = 4096
	
	// Protocol flags
	FlagFragmented     byte = 0x01
	FlagLastFragment   byte = 0x02
	FlagCompressed     byte = 0x04
	FlagEncrypted      byte = 0x08
	FlagUrgent         byte = 0x10
	FlagReserved1      byte = 0x20
	FlagReserved2      byte = 0x40
	FlagReserved3      byte = 0x80
)

// TLV represents a Type-Length-Value structure
type TLV struct {
	Type   byte
	Length uint16
	Value  []byte
}

// NewTLV creates a new TLV structure
func NewTLV(tlvType byte, value []byte) *TLV {
	return &TLV{
		Type:   tlvType,
		Length: uint16(len(value)),
		Value:  value,
	}
}

// EncodeTLV encodes a TLV structure to bytes
func EncodeTLV(tlv *TLV) []byte {
	buf := new(bytes.Buffer)
	
	// Write type (1 byte)
	buf.WriteByte(tlv.Type)
	
	// Write length (2 bytes)
	binary.Write(buf, binary.BigEndian, tlv.Length)
	
	// Write value
	buf.Write(tlv.Value)
	
	return buf.Bytes()
}

// DecodeTLV decodes bytes into a TLV structure
func DecodeTLV(data []byte) (*TLV, int, error) {
	if len(data) < 3 { // Minimum size: Type(1) + Length(2)
		return nil, 0, errors.New("data too short for TLV")
	}
	
	tlv := &TLV{
		Type: data[0],
	}
	
	// Read length (2 bytes)
	tlv.Length = binary.BigEndian.Uint16(data[1:3])
	
	// Validate length
	if int(tlv.Length) > len(data)-3 {
		return nil, 0, errors.New("TLV length exceeds available data")
	}
	
	// Read value
	tlv.Value = make([]byte, tlv.Length)
	copy(tlv.Value, data[3:3+tlv.Length])
	
	// Return TLV and total bytes consumed
	return tlv, 3 + int(tlv.Length), nil
}

// CalculateChecksum calculates CRC32 checksum for the given data
func CalculateChecksum(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

// EncodePacket encodes a packet into bytes with proper checksum
func EncodePacket(p *Packet) []byte {
	buf := new(bytes.Buffer)
	
	// Reserve space for header
	headerBytes := make([]byte, HeaderSize)
	
	// Set header fields
	headerBytes[0] = p.Header.Version
	headerBytes[1] = byte(p.Header.EncAlgorithm)
	headerBytes[2] = byte(p.Header.Type)
	
	// Set TaskID (4 bytes)
	binary.BigEndian.PutUint32(headerBytes[3:7], p.Header.TaskID)
	
	// Reserve space for checksum (4 bytes)
	// Will be calculated after writing data
	
	// Write header to buffer
	buf.Write(headerBytes)
	
	// Write data to buffer
	buf.Write(p.Data)
	
	// Get the full packet bytes
	packetBytes := buf.Bytes()
	
	// Calculate checksum (excluding the checksum field itself)
	// Create a copy with zero checksum for calculation
	checksumData := make([]byte, len(packetBytes))
	copy(checksumData, packetBytes)
	
	// Zero out the checksum field in the copy
	binary.BigEndian.PutUint32(checksumData[7:11], 0)
	
	// Calculate checksum
	checksum := CalculateChecksum(checksumData)
	
	// Set checksum in the original packet bytes
	binary.BigEndian.PutUint32(packetBytes[7:11], checksum)
	
	return packetBytes
}

// DecodePacket decodes bytes into a packet and verifies the checksum
func DecodePacket(data []byte) (*Packet, error) {
	if len(data) < HeaderSize {
		return nil, errors.New("data too short for packet header")
	}
	
	// Extract header fields
	version := data[0]
	encAlgorithm := EncryptionAlgorithm(data[1])
	packetType := PacketType(data[2])
	taskID := binary.BigEndian.Uint32(data[3:7])
	checksum := binary.BigEndian.Uint32(data[7:11])
	
	// Create a copy with zero checksum for verification
	checksumData := make([]byte, len(data))
	copy(checksumData, data)
	
	// Zero out the checksum field in the copy
	binary.BigEndian.PutUint32(checksumData[7:11], 0)
	
	// Calculate checksum
	calculatedChecksum := CalculateChecksum(checksumData)
	
	// Verify checksum
	if checksum != calculatedChecksum {
		return nil, errors.New("checksum verification failed")
	}
	
	// Create packet
	packet := &Packet{
		Header: PacketHeader{
			Version:      version,
			EncAlgorithm: encAlgorithm,
			Type:         packetType,
			TaskID:       taskID,
			Checksum:     checksum,
		},
		Data: data[HeaderSize:],
	}
	
	return packet, nil
}

// FragmentPacket splits a large packet into multiple smaller fragments
func FragmentPacket(p *Packet, maxFragmentSize int) []*Packet {
	if len(p.Data) <= maxFragmentSize {
		// No fragmentation needed
		return []*Packet{p}
	}
	
	// Calculate number of fragments needed
	dataSize := len(p.Data)
	fragmentCount := (dataSize + maxFragmentSize - 1) / maxFragmentSize
	fragments := make([]*Packet, 0, fragmentCount)
	
	// Create fragments
	for i := 0; i < fragmentCount; i++ {
		start := i * maxFragmentSize
		end := start + maxFragmentSize
		if end > dataSize {
			end = dataSize
		}
		
		// Create fragment packet
		fragment := &Packet{
			Header: PacketHeader{
				Version:      p.Header.Version,
				EncAlgorithm: p.Header.EncAlgorithm,
				Type:         p.Header.Type,
				TaskID:       p.Header.TaskID,
				Checksum:     0, // Will be calculated during encoding
			},
			Data: make([]byte, end-start+2), // +2 for fragment header
		}
		
		// Set fragment header
		// First byte: fragment index
		// Second byte: flags (FlagFragmented + FlagLastFragment if last fragment)
		fragment.Data[0] = byte(i)
		fragment.Data[1] = FlagFragmented
		if i == fragmentCount-1 {
			fragment.Data[1] |= FlagLastFragment
		}
		
		// Copy fragment data
		copy(fragment.Data[2:], p.Data[start:end])
		
		fragments = append(fragments, fragment)
	}
	
	return fragments
}

// ReassemblePacket reconstructs a packet from fragments
func ReassemblePacket(fragments []*Packet) (*Packet, error) {
	if len(fragments) == 0 {
		return nil, errors.New("no fragments provided")
	}
	
	if len(fragments) == 1 && fragments[0].Data[1]&FlagFragmented == 0 {
		// Not a fragmented packet
		return fragments[0], nil
	}
	
	// Sort fragments by index
	// First determine the maximum index to properly size our array
	maxIndex := 0
	for _, fragment := range fragments {
		if len(fragment.Data) < 2 {
			return nil, errors.New("invalid fragment: too short")
		}
		
		index := int(fragment.Data[0])
		if index > maxIndex {
			maxIndex = index
		}
	}
	
	// Create sorted array with proper size
	sortedFragments := make([]*Packet, maxIndex+1)
	for _, fragment := range fragments {
		index := int(fragment.Data[0])
		sortedFragments[index] = fragment
	}
	
	// Verify all fragments are present
	for i, fragment := range sortedFragments {
		if fragment == nil {
			return nil, fmt.Errorf("missing fragment: %d", i)
		}
	}
	
	// Calculate total data size
	totalSize := 0
	for _, fragment := range sortedFragments {
		totalSize += len(fragment.Data) - 2 // Subtract fragment header size
	}
	
	// Create reassembled packet
	packet := &Packet{
		Header: PacketHeader{
			Version:      fragments[0].Header.Version,
			EncAlgorithm: fragments[0].Header.EncAlgorithm,
			Type:         fragments[0].Header.Type,
			TaskID:       fragments[0].Header.TaskID,
			Checksum:     0, // Will be calculated during encoding
		},
		Data: make([]byte, totalSize),
	}
	
	// Copy data from fragments
	offset := 0
	for _, fragment := range sortedFragments {
		fragmentData := fragment.Data[2:] // Skip fragment header
		copy(packet.Data[offset:], fragmentData)
		offset += len(fragmentData)
	}
	
	return packet, nil
}
