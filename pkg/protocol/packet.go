package protocol

// PacketType represents the type of packet being transmitted
type PacketType byte

const (
	PacketTypeHeartbeat PacketType = iota
	PacketTypeCommand
	PacketTypeResponse
	PacketTypeModuleData
	PacketTypeModuleResponse
	PacketTypeError
	PacketTypeKeyExchange
	PacketTypeProtocolSwitch
)

// EncryptionAlgorithm represents the encryption algorithm used
type EncryptionAlgorithm byte

const (
	EncryptionAlgorithmNone EncryptionAlgorithm = iota
	EncryptionAlgorithmAES
	EncryptionAlgorithmChacha20
)

// ProtocolVersion represents the version of the protocol
const ProtocolVersion byte = 1

// PacketHeader represents the header of a packet
type PacketHeader struct {
	Version      byte               // Protocol version
	EncAlgorithm EncryptionAlgorithm // Encryption algorithm
	Type         PacketType         // Packet type
	TaskID       uint32             // Task identifier
	Checksum     uint32             // Packet checksum
}

// Packet represents a complete packet with header and data
type Packet struct {
	Header PacketHeader
	Data   []byte
}

// NewPacket creates a new packet with the specified type and data
func NewPacket(packetType PacketType, data []byte) *Packet {
	// TODO: Calculate checksum
	return &Packet{
		Header: PacketHeader{
			Version:      ProtocolVersion,
			EncAlgorithm: EncryptionAlgorithmNone, // Default to no encryption
			Type:         packetType,
			TaskID:       0, // Will be set by the task manager
			Checksum:     0, // Will be calculated before sending
		},
		Data: data,
	}
}

// SetEncryptionAlgorithm sets the encryption algorithm for the packet
func (p *Packet) SetEncryptionAlgorithm(algorithm EncryptionAlgorithm) {
	p.Header.EncAlgorithm = algorithm
}

// SetTaskID sets the task ID for the packet
func (p *Packet) SetTaskID(taskID uint32) {
	p.Header.TaskID = taskID
}

// CalculateChecksum calculates and sets the checksum for the packet
func (p *Packet) CalculateChecksum() {
	// Use the CalculateChecksum function from encoder.go
	// Create a copy of the packet with zero checksum for calculation
	packetCopy := *p
	packetCopy.Header.Checksum = 0
	
	// Encode the packet without checksum
	encodedPacket := packetCopy.Encode()
	
	// Calculate checksum
	p.Header.Checksum = CalculateChecksum(encodedPacket)
}

// Encode serializes the packet into a byte slice
func (p *Packet) Encode() []byte {
	// Use the EncodePacket function from encoder.go
	return EncodePacket(p)
}

// Decode deserializes a byte slice into a packet
func Decode(data []byte) (*Packet, error) {
	// Use the DecodePacket function from encoder.go
	return DecodePacket(data)
}
