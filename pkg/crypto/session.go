package crypto

import (
	"fmt"
	"sync"
	"time"
)

// Session represents an encryption session
type Session struct {
	ID        SessionID
	Encryptor Encryptor
	Created   time.Time
	LastUsed  time.Time
}

// SessionManager manages encryption sessions
type SessionManager struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new encryption session
func (m *SessionManager) CreateSession(sessionID SessionID, algorithm Algorithm) (*Session, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Check if session already exists
	if _, exists := m.sessions[string(sessionID)]; exists {
		return nil, fmt.Errorf("session already exists")
	}
	
	// Create encryptor
	encryptor, err := NewEncryptor(algorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}
	
	// Create session
	session := &Session{
		ID:        sessionID,
		Encryptor: encryptor,
		Created:   time.Now(),
		LastUsed:  time.Now(),
	}
	
	// Add session to map
	m.sessions[string(sessionID)] = session
	
	return session, nil
}

// GetSession gets a session by ID
func (m *SessionManager) GetSession(sessionID SessionID) (*Session, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Check if session exists
	session, exists := m.sessions[string(sessionID)]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}
	
	// Update last used time
	session.LastUsed = time.Now()
	
	return session, nil
}

// RemoveSession removes a session
func (m *SessionManager) RemoveSession(sessionID SessionID) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Check if session exists
	if _, exists := m.sessions[string(sessionID)]; !exists {
		return fmt.Errorf("session not found")
	}
	
	// Remove session from map
	delete(m.sessions, string(sessionID))
	
	return nil
}

// CleanupSessions removes expired sessions
func (m *SessionManager) CleanupSessions(maxAge time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Get current time
	now := time.Now()
	
	// Check each session
	for id, session := range m.sessions {
		// Check if session has expired
		if now.Sub(session.LastUsed) > maxAge {
			// Remove session
			delete(m.sessions, id)
		}
	}
}

// Shutdown shuts down the session manager
func (m *SessionManager) Shutdown() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Clear all sessions
	m.sessions = make(map[string]*Session)
}
