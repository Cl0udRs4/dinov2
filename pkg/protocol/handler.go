package protocol

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"dinoc2/pkg/crypto"
)

// ProtocolHandler manages the protocol layer operations
type ProtocolHandler struct {
	sessionManager *crypto.SessionManager
	fragmentCache  map[uint32][]*Packet // Cache for packet fragments
	cacheMutex     sync.RWMutex
	jitterEnabled  bool
	jitterRange    [2]time.Duration // Min and max jitter delay
}

// NewProtocolHandler creates a new protocol handler
func NewProtocolHandler() *ProtocolHandler {
	return &ProtocolHandler{
		sessionManager: crypto.NewSessionManager(),
		fragmentCache:  make(map[uint32][]*Packet),
		jitterEnabled:  true,
		jitterRange:    [2]time.Duration{10 * time.Millisecond, 100 * time.Millisecond},
	}
}

// ProcessIncomingPacket processes an incoming packet
func (h *ProtocolHandler) ProcessIncomingPacket(data []byte, sessionID crypto.SessionID) (*Packet, error) {
	// Decode the packet
	packet, err := DecodePacket(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode packet: %w", err)
	}
	
	// Handle fragmented packets
	if len(packet.Data) > 0 && packet.Data[1]&FlagFragmented != 0 {
		return h.handleFragmentedPacket(packet)
	}
	
	// Handle encrypted packets
	if packet.Header.EncAlgorithm != EncryptionAlgorithmNone {
		decryptedPacket, err := h.decryptPacket(packet, sessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt packet: %w", err)
		}
		packet = decryptedPacket
	}
	
	return packet, nil
}

// PrepareOutgoingPacket prepares a packet for sending
func (h *ProtocolHandler) PrepareOutgoingPacket(packet *Packet, sessionID crypto.SessionID, encrypt bool) ([][]byte, error) {
	// Apply encryption if requested
	if encrypt {
		encryptedPacket, err := h.encryptPacket(packet, sessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt packet: %w", err)
		}
		packet = encryptedPacket
	}
	
	// Fragment the packet if needed
	fragments := FragmentPacket(packet, MaxFragmentSize)
	
	// Encode each fragment
	encodedFragments := make([][]byte, len(fragments))
	for i, fragment := range fragments {
		encodedFragments[i] = EncodePacket(fragment)
	}
	
	return encodedFragments, nil
}

// handleFragmentedPacket processes a fragmented packet
func (h *ProtocolHandler) handleFragmentedPacket(packet *Packet) (*Packet, error) {
	h.cacheMutex.Lock()
	defer h.cacheMutex.Unlock()
	
	// Get or create fragment cache for this task ID
	fragments, exists := h.fragmentCache[packet.Header.TaskID]
	if !exists {
		fragments = make([]*Packet, 0)
	}
	
	// Add the fragment to the cache
	fragments = append(fragments, packet)
	h.fragmentCache[packet.Header.TaskID] = fragments
	
	// Check if this is the last fragment
	if packet.Data[1]&FlagLastFragment != 0 {
		// Reassemble the packet
		reassembledPacket, err := ReassemblePacket(fragments)
		if err != nil {
			return nil, fmt.Errorf("failed to reassemble packet: %w", err)
		}
		
		// Remove the fragments from the cache
		delete(h.fragmentCache, packet.Header.TaskID)
		
		return reassembledPacket, nil
	}
	
	// Not the last fragment, wait for more
	return nil, errors.New("packet fragmented, waiting for more fragments")
}

// encryptPacket encrypts a packet using the specified session
func (h *ProtocolHandler) encryptPacket(packet *Packet, sessionID crypto.SessionID) (*Packet, error) {
	// Get the session
	session, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	// Encrypt the data
	encryptedData, err := session.Encryptor.Encrypt(packet.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}
	
	// Create a new packet with encrypted data
	encryptedPacket := &Packet{
		Header: PacketHeader{
			Version:      packet.Header.Version,
			EncAlgorithm: getEncryptionAlgorithm(session.Encryptor.Algorithm()),
			Type:         packet.Header.Type,
			TaskID:       packet.Header.TaskID,
			Checksum:     0, // Will be calculated during encoding
		},
		Data: encryptedData,
	}
	
	return encryptedPacket, nil
}

// decryptPacket decrypts a packet using the specified session
func (h *ProtocolHandler) decryptPacket(packet *Packet, sessionID crypto.SessionID) (*Packet, error) {
	// Get the session
	session, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	// Decrypt the data
	decryptedData, err := session.Encryptor.Decrypt(packet.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}
	
	// Create a new packet with decrypted data
	decryptedPacket := &Packet{
		Header: PacketHeader{
			Version:      packet.Header.Version,
			EncAlgorithm: EncryptionAlgorithmNone,
			Type:         packet.Header.Type,
			TaskID:       packet.Header.TaskID,
			Checksum:     0, // Will be calculated during encoding
		},
		Data: decryptedData,
	}
	
	return decryptedPacket, nil
}

// getEncryptionAlgorithm converts a crypto.Algorithm to an EncryptionAlgorithm
func getEncryptionAlgorithm(algorithm crypto.Algorithm) EncryptionAlgorithm {
	switch algorithm {
	case crypto.AlgorithmAES:
		return EncryptionAlgorithmAES
	case crypto.AlgorithmChacha20:
		return EncryptionAlgorithmChacha20
	default:
		return EncryptionAlgorithmNone
	}
}

// CreateSession creates a new encryption session
func (h *ProtocolHandler) CreateSession(sessionID crypto.SessionID, algorithm crypto.Algorithm) error {
	_, err := h.sessionManager.CreateSession(sessionID, algorithm)
	return err
}

// RemoveSession removes an encryption session
func (h *ProtocolHandler) RemoveSession(sessionID crypto.SessionID) error {
	return h.sessionManager.RemoveSession(sessionID)
}

// SetJitterEnabled enables or disables communication jitter
func (h *ProtocolHandler) SetJitterEnabled(enabled bool) {
	h.jitterEnabled = enabled
}

// SetJitterRange sets the min and max jitter delay
func (h *ProtocolHandler) SetJitterRange(min, max time.Duration) {
	h.jitterRange = [2]time.Duration{min, max}
}

// GenerateRandomInt64 generates a random int64 between 0 and max
func GenerateRandomInt64(max int64) int64 {
	if max <= 0 {
		return 0
	}
	
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0
	}
	
	return n.Int64()
}

// GetJitterDelay returns a random jitter delay within the configured range
func (h *ProtocolHandler) GetJitterDelay() time.Duration {
	if !h.jitterEnabled {
		return 0
	}
	
	min := int64(h.jitterRange[0])
	max := int64(h.jitterRange[1])
	
	// Generate a random duration between min and max
	jitter := min
	if max > min {
		jitter += GenerateRandomInt64(max - min)
	}
	
	return time.Duration(jitter)
}

// Shutdown shuts down the protocol handler
func (h *ProtocolHandler) Shutdown() {
	h.sessionManager.Shutdown()
}
