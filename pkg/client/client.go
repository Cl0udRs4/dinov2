package client

import (
	"dinoc2/pkg/crypto"
	"time"
)

// Client represents a connected client
type Client struct {
	SessionID crypto.SessionID
	Address   string
	Protocol  string
	LastSeen  time.Time
	Info      map[string]interface{}
}

// NewClient creates a new client
func NewClient(sessionID crypto.SessionID, address, protocol string) *Client {
	return &Client{
		SessionID: sessionID,
		Address:   address,
		Protocol:  protocol,
		LastSeen:  time.Now(),
		Info:      make(map[string]interface{}),
	}
}

// UpdateLastSeen updates the last seen timestamp
func (c *Client) UpdateLastSeen() {
	c.LastSeen = time.Now()
}

// SetInfo sets client information
func (c *Client) SetInfo(key string, value interface{}) {
	if c.Info == nil {
		c.Info = make(map[string]interface{})
	}
	c.Info[key] = value
}

// GetInfo gets client information
func (c *Client) GetInfo(key string) (interface{}, bool) {
	if c.Info == nil {
		return nil, false
	}
	value, ok := c.Info[key]
	return value, ok
}
