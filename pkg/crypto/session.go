package crypto

import (
	"fmt"
	"sync"
)

// Session represents an encryption session
type Session struct {
	ID       SessionID
	Encryptor Encryptor
}

// SessionManager manages encryption sessions
type SessionManager struct {
	sessions map[SessionID]*Session
	mutex    sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[SessionID]*Session),
	}
}

// CreateSession creates a new session
func (m *SessionManager) CreateSession(sessionID SessionID, algorithm Algorithm) (*Session, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Check if session already exists
	if _, exists := m.sessions[sessionID]; exists {
		return nil, fmt.Errorf("session already exists")
	}
	
	// Create encryptor
	encryptor, err := NewEncryptor(algorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}
	
	// Create session
	session := &Session{
		ID:       sessionID,
		Encryptor: encryptor,
	}
	
	// Add session to map
	m.sessions[sessionID] = session
	
	return session, nil
}

// GetSession gets a session by ID
func (m *SessionManager) GetSession(sessionID SessionID) (*Session, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Check if session exists
	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}
	
	return session, nil
}

// DeleteSession deletes a session
func (m *SessionManager) DeleteSession(sessionID SessionID) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Delete session from map
	delete(m.sessions, sessionID)
}
