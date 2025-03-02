package crypto

import (
	"errors"
	"sync"
	"time"
)

// SessionID is defined in crypto.go

// Session represents an encryption session with a client
type Session struct {
	ID            SessionID
	Encryptor     Encryptor
	CreatedAt     time.Time
	LastActivity  time.Time
	LastRotation  time.Time
	RotationCount int
}

// SessionManager manages encryption sessions for multiple clients
type SessionManager struct {
	sessions      map[SessionID]*Session
	mutex         sync.RWMutex
	rotationTimer *time.Timer
	rotationDone  chan struct{}
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	manager := &SessionManager{
		sessions:     make(map[SessionID]*Session),
		rotationDone: make(chan struct{}),
	}
	
	// Start the key rotation timer
	manager.startRotationTimer()
	
	return manager
}

// CreateSession creates a new encryption session with the specified algorithm
func (m *SessionManager) CreateSession(id SessionID, algorithm Algorithm) (*Session, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Check if session already exists
	if _, exists := m.sessions[id]; exists {
		return nil, errors.New("session already exists")
	}
	
	// Create a new encryptor
	encryptor, err := Factory(algorithm)
	if err != nil {
		return nil, err
	}
	
	// Create a new session
	now := time.Now()
	session := &Session{
		ID:            id,
		Encryptor:     encryptor,
		CreatedAt:     now,
		LastActivity:  now,
		LastRotation:  now,
		RotationCount: 0,
	}
	
	// Add the session to the map
	m.sessions[id] = session
	
	return session, nil
}

// GetSession retrieves an existing session
func (m *SessionManager) GetSession(id SessionID) (*Session, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	session, exists := m.sessions[id]
	if !exists {
		return nil, errors.New("session not found")
	}
	
	// Update last activity time
	session.LastActivity = time.Now()
	
	return session, nil
}

// RemoveSession removes a session
func (m *SessionManager) RemoveSession(id SessionID) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if _, exists := m.sessions[id]; !exists {
		return errors.New("session not found")
	}
	
	delete(m.sessions, id)
	return nil
}

// RotateSessionKey rotates the key for a specific session
func (m *SessionManager) RotateSessionKey(id SessionID) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	session, exists := m.sessions[id]
	if !exists {
		return errors.New("session not found")
	}
	
	// Rotate the key
	if err := session.Encryptor.RotateKey(); err != nil {
		return err
	}
	
	// Update session metadata
	session.LastRotation = time.Now()
	session.RotationCount++
	
	return nil
}

// RotateAllKeys rotates keys for all active sessions
func (m *SessionManager) RotateAllKeys() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	now := time.Now()
	
	for _, session := range m.sessions {
		// Skip inactive sessions (no activity in the last hour)
		if now.Sub(session.LastActivity) > time.Hour {
			continue
		}
		
		// Rotate the key
		if err := session.Encryptor.RotateKey(); err != nil {
			// Log the error but continue with other sessions
			continue
		}
		
		// Update session metadata
		session.LastRotation = now
		session.RotationCount++
	}
}

// startRotationTimer starts a timer to periodically rotate keys
func (m *SessionManager) startRotationTimer() {
	// Rotate keys every 12 hours
	rotationInterval := 12 * time.Hour
	
	m.rotationTimer = time.NewTimer(rotationInterval)
	
	go func() {
		for {
			select {
			case <-m.rotationTimer.C:
				// Rotate all keys
				m.RotateAllKeys()
				
				// Reset the timer
				m.rotationTimer.Reset(rotationInterval)
			case <-m.rotationDone:
				// Stop the timer
				if !m.rotationTimer.Stop() {
					<-m.rotationTimer.C
				}
				return
			}
		}
	}()
}

// Shutdown stops the session manager and cleans up resources
func (m *SessionManager) Shutdown() {
	// Signal the rotation timer to stop
	close(m.rotationDone)
	
	// Clear all sessions
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Clear all sessions
	m.sessions = make(map[SessionID]*Session)
}

// GetSessionCount returns the number of active sessions
func (m *SessionManager) GetSessionCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return len(m.sessions)
}

// GetActiveSessions returns a list of active session IDs
func (m *SessionManager) GetActiveSessions() []SessionID {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	ids := make([]SessionID, 0, len(m.sessions))
	for id := range m.sessions {
		ids = append(ids, id)
	}
	
	return ids
}
