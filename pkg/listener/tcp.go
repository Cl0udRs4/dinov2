package listener

import (
	"fmt"
	"net"
	"sync"
)

// TCPListener implements the Listener interface for TCP protocol
type TCPListener struct {
	config     ListenerConfig
	listener   net.Listener
	status     ListenerStatus
	statusLock sync.RWMutex
	stopChan   chan struct{}
}

// NewTCPListener creates a new TCP listener
func NewTCPListener(config ListenerConfig) *TCPListener {
	return &TCPListener{
		config:   config,
		status:   StatusStopped,
		stopChan: make(chan struct{}),
	}
}

// Start implements the Listener interface
func (l *TCPListener) Start() error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status == StatusRunning {
		return fmt.Errorf("listener is already running")
	}

	addr := fmt.Sprintf("%s:%d", l.config.Address, l.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		l.status = StatusError
		return fmt.Errorf("failed to start TCP listener on %s: %w", addr, err)
	}

	l.listener = listener
	l.status = StatusRunning
	l.stopChan = make(chan struct{})

	// Start accepting connections in a goroutine
	go l.acceptConnections()

	return nil
}

// Stop implements the Listener interface
func (l *TCPListener) Stop() error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status != StatusRunning {
		return nil // Already stopped
	}

	// Signal the accept loop to stop
	close(l.stopChan)

	// Close the listener
	if l.listener != nil {
		err := l.listener.Close()
		if err != nil {
			l.status = StatusError
			return fmt.Errorf("error closing TCP listener: %w", err)
		}
	}

	l.status = StatusStopped
	return nil
}

// Status implements the Listener interface
func (l *TCPListener) Status() ListenerStatus {
	l.statusLock.RLock()
	defer l.statusLock.RUnlock()
	return l.status
}

// Configure implements the Listener interface
func (l *TCPListener) Configure(config ListenerConfig) error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status == StatusRunning {
		return fmt.Errorf("cannot configure a running listener")
	}

	l.config = config
	return nil
}

// acceptConnections handles incoming TCP connections
func (l *TCPListener) acceptConnections() {
	for {
		select {
		case <-l.stopChan:
			return
		default:
			conn, err := l.listener.Accept()
			if err != nil {
				select {
				case <-l.stopChan:
					// Listener was closed intentionally, not an error
					return
				default:
					// Actual error occurred
					l.statusLock.Lock()
					l.status = StatusError
					l.statusLock.Unlock()
					fmt.Printf("Error accepting connection: %v\n", err)
					return
				}
			}

			// Handle the connection in a new goroutine
			go l.handleConnection(conn)
		}
	}
}

// handleConnection processes a client connection
func (l *TCPListener) handleConnection(conn net.Conn) {
	defer conn.Close()

	// TODO: Implement actual connection handling
	// This will be integrated with the protocol layer
	fmt.Printf("New connection from %s\n", conn.RemoteAddr())

	// For now, just read and discard data
	buffer := make([]byte, 1024)
	for {
		_, err := conn.Read(buffer)
		if err != nil {
			break
		}
		// Process data will be implemented later
	}
}
