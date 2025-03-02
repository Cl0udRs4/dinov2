package client

// ProtocolType defines the type of protocol used for communication
type ProtocolType string

// Protocol types
const (
	ProtocolTypeTCP       ProtocolType = "tcp"
	ProtocolTypeHTTP      ProtocolType = "http"
	ProtocolTypeWebSocket ProtocolType = "websocket"
	ProtocolTypeDNS       ProtocolType = "dns"
	ProtocolTypeICMP      ProtocolType = "icmp"
)

// ConnectionState defines the state of the client connection
type ConnectionState int

// Connection states
const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateReconnecting
	StateSwitchingProtocol
)
