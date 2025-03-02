package protocol

import (
	"dinoc2/pkg/crypto"
	"fmt"
)

// ProtocolHandler handles protocol operations
type ProtocolHandler struct {
	sessionManager *crypto.SessionManager
}

// NewProtocolHandler creates a new protocol handler
func NewProtocolHandler() *ProtocolHandler {
	return &ProtocolHandler{
		sessionManager: crypto.NewSessionManager(),
	}
}

// CreateSession creates a new session
func (h *ProtocolHandler) CreateSession(sessionID crypto.SessionID, algorithm crypto.Algorithm) error {
	_, err := h.sessionManager.CreateSession(sessionID, algorithm)
	return err
}

// ProcessIncomingPacket processes an incoming packet
func (h *ProtocolHandler) ProcessIncomingPacket(data []byte, sessionID crypto.SessionID) (*Packet, error) {
	// Get session
	session, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	// Decrypt data if needed
	if session.Encryptor.Algorithm() != crypto.AlgorithmAES {
		decryptedData, err := session.Encryptor.Decrypt(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt data: %w", err)
		}
		data = decryptedData
	}
	
	// Decode packet
	packet, err := DecodePacket(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode packet: %w", err)
	}
	
	return packet, nil
}

// PrepareOutgoingPacket prepares an outgoing packet
func (h *ProtocolHandler) PrepareOutgoingPacket(packet *Packet, sessionID crypto.SessionID, encrypt bool) ([][]byte, error) {
	// Encode packet
	data, err := EncodePacket(packet)
	if err != nil {
		return nil, fmt.Errorf("failed to encode packet: %w", err)
	}
	
	// Get session
	session, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	// Encrypt data if needed
	if encrypt && session.Encryptor.Algorithm() != crypto.AlgorithmAES {
		encryptedData, err := session.Encryptor.Encrypt(data)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt data: %w", err)
		}
		data = encryptedData
	}
	
	// Return data as a single fragment
	return [][]byte{data}, nil
}
