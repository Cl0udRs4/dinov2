package client

import (
	"dinoc2/pkg/crypto"
	"dinoc2/pkg/protocol"
)

// Connection defines the interface for client connections
type Connection interface {
	// Connect establishes a connection to the server
	Connect() error

	// Close closes the connection
	Close() error

	// SendPacket sends a packet to the server
	SendPacket(packet *protocol.Packet) error

	// ReceivePacket receives a packet from the server
	ReceivePacket() (*protocol.Packet, error)

	// GetProtocolType returns the protocol type
	GetProtocolType() ProtocolType
}

// BaseConnection provides common functionality for all connection types
type BaseConnection struct {
	serverAddress   string
	protocolHandler *protocol.ProtocolHandler
	sessionID       crypto.SessionID
	protocolType    ProtocolType
	connected       bool
}

// NewBaseConnection creates a new base connection
func NewBaseConnection(serverAddress string, protocolHandler *protocol.ProtocolHandler, sessionID crypto.SessionID, protocolType ProtocolType) *BaseConnection {
	return &BaseConnection{
		serverAddress:   serverAddress,
		protocolHandler: protocolHandler,
		sessionID:       sessionID,
		protocolType:    protocolType,
		connected:       false,
	}
}

// GetProtocolType returns the protocol type
func (c *BaseConnection) GetProtocolType() ProtocolType {
	return c.protocolType
}
